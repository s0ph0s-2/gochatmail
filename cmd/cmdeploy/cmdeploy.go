package main

import (
	"github.com/s0ph0s-2/gochatmail/internal/config"

	"bytes"
	"crypto/sha1"
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	go_qr "github.com/piglig/go-qr"
	"github.com/skratchdot/open-golang/open"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func do_init(fqdn string) {
	config := config.NewChatmailConfig(fqdn)
	err := config.Save("./chatmail.json")
	if err != nil {
		panic(err)
	}

	fmt.Println("Chatmail server configuration generated! Edit ./chatmail.json in your favorite text editor to change any of the default settings, if you would like.")
}

func copy_file(src string, dst string) error {
	data, r_err := os.ReadFile(src)
	if r_err != nil {
		return r_err
	}
	w_err := os.WriteFile(dst, data, 0644)
	if w_err != nil {
		return w_err
	}
	return nil
}

func splitext(filename string) (string, string) {
	stem, ext, _ := strings.Cut(filename, ".")
	return stem, ext
}

func make_page_name(stem string) string {
	if stem == "index" {
		return "home"
	} else {
		return stem
	}
}

type page_vars struct {
	Title      string
	AutoReload bool
	Config     config.ChatmailConfig
}

func make_markdown_renderer() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.Typographer),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
}

//go:embed delta-chat-bw.svg.tmpl
var dc_logo_tmpl string

func generate_qr_code(fqdn string) string {
	// Generate DCACCOUNT: URL to go in code
	new_acct_url := fmt.Sprintf("DCACCOUNT:https://%s/new", fqdn)
	// Set up QR code generator tool
	qr, err := go_qr.EncodeText(new_acct_url, go_qr.High)
	if err != nil {
		panic(err)
	}
	qr_scale := 1
	qr_border := 4
	config := go_qr.NewQrCodeImgConfig(qr_scale, qr_border)
	// Ask it to produce an SVG of the URL
	var svgbuf bytes.Buffer
	err = qr.WriteAsSVG(config, &svgbuf, "#FFFFFF", "#000000")
	// Load the DeltaChat logo as a template
	// Compute offset constants e & f for the matrix transform
	qr_dims := (qr.GetSize() * qr_scale) + (qr_border * 2)
	offset_xy := (float32(qr_dims) / 2.0) - 4.0
	// Insert templated DC logo into QR code SVG with transform params
	tmpl, err := template.New("dc_logo_tmpl").Parse(dc_logo_tmpl)
	if err != nil {
		panic(err)
	}
	var logo bytes.Buffer
	err = tmpl.Execute(&logo, offset_xy)
	qr_svg_str := svgbuf.String()
	split_point := strings.LastIndex(qr_svg_str, "<")
	if split_point < 0 {
		panic(err)
	}
	return qr_svg_str[0:split_point] + logo.String() + "</svg>"
}

func build_website(cm_config config.ChatmailConfig, input_dir string, output_dir string) {
	page_layout_file := filepath.Join(input_dir, "page-layout.html")
	templates, err := template.New("page_layout").ParseFiles(page_layout_file)
	if err != nil {
		panic(err)
	}
	contents, err := os.ReadDir(input_dir)
	if err != nil {
		panic(err)
	}
	qr_invite_file := filepath.Join(output_dir, "qr-chatmail-invite-"+cm_config.MailFullyQualifiedDomainName+".svg")
	qr_invite_data := generate_qr_code(cm_config.MailFullyQualifiedDomainName)
	err = os.WriteFile(qr_invite_file, []byte(qr_invite_data), 0644)
	if err != nil {
		panic(err)
	}
	md := make_markdown_renderer()
	for _, dirent := range contents {
		if dirent.IsDir() {
			continue
		}
		dirent_name := dirent.Name()
		input_file := filepath.Join(input_dir, dirent_name)
		if filepath.Ext(dirent_name) == ".md" {
			local_tmpls, err := templates.Clone()
			if err != nil {
				panic(err)
			}
			stem, _ := splitext(dirent_name)
			page_name := make_page_name(stem)
			page_content_md, err := os.ReadFile(input_file)
			if err != nil {
				panic(err)
			}
			var md_buf bytes.Buffer
			if err := md.Convert(page_content_md, &md_buf); err != nil {
				panic(err)
			}
			_, err = local_tmpls.New("PageContent").Parse(md_buf.String())
			if err != nil {
				panic(err)
			}
			this_page_vars := page_vars{page_name, true, cm_config}
			var html_buf bytes.Buffer
			err = local_tmpls.ExecuteTemplate(&html_buf, "page-layout.html", this_page_vars)
			if err != nil {
				panic(err)
			}
			output_file := filepath.Join(output_dir, stem+".html")
			err = os.WriteFile(output_file, html_buf.Bytes(), 0644)
			if err != nil {
				panic(err)
			}
		} else if dirent_name != "page-layout.html" {
			output_file := filepath.Join(output_dir, dirent_name)
			copy_err := copy_file(input_file, output_file)
			if copy_err != nil {
				panic(copy_err)
			}
		}
	}
}

type file_analysis_result struct {
	Mtime int64
	Hash  []byte
}

func analyze_dir(dir string) map[string]file_analysis_result {
	// Compute hashes of every file in a directory (that isn't hidden or a vim
	// swap file).
	results := make(map[string]file_analysis_result)
	dir_contents, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, dirent := range dir_contents {
		if dirent.IsDir() {
			continue
		}
		name := dirent.Name()
		if filepath.Ext(name) == ".swp" {
			continue
		}
		info, err := dirent.Info()
		if err != nil {
			panic(err)
		}
		lastModified := info.ModTime().Unix()
		f, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			panic(err)
		}
		defer f.Close()
		h := sha1.New()
		if _, err = io.Copy(h, f); err != nil {
			panic(err)
		}
		results[name] = file_analysis_result{lastModified, h.Sum(nil)}
	}
	return results
}

func watch_for_changes(cm_config config.ChatmailConfig, input_dir string, output_dir string) {
	fmt.Printf("Watching for changes in %s...\n", input_dir)
	fmt.Println("Press Ctrl+C to stop watching once you're finished editing.")
	var current_state map[string]file_analysis_result = analyze_dir(input_dir)
	var next_state map[string]file_analysis_result
	for {
		next_state = analyze_dir(input_dir)
		statesEqual := reflect.DeepEqual(current_state, next_state)
		if statesEqual {
			time.Sleep(1 * time.Second)
			continue
		}
		current_state = next_state
		build_website(cm_config, input_dir, output_dir)
		fmt.Println("Changes detected! Pages have been regenerated.")
	}
}

func main() {
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)

	webdevCmd := flag.NewFlagSet("webdev", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Println("expected 'init' or 'webdev' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		initCmd.Parse(os.Args[2:])
		tail := initCmd.Args()
		if len(tail) < 1 {
			fmt.Println("you have to provide the fully qualified domain name of your new chat server")
			os.Exit(1)
		}
		fqdn := tail[0]
		do_init(fqdn)
	case "webdev":
		webdevCmd.Parse(os.Args[2:])
		var cm_config config.ChatmailConfig
		config_file := filepath.Join(".", "chatmail.json")
		cl_err := config.LoadChatmailConfigFromFile(config_file, &cm_config)
		if cl_err != nil {
			panic(cl_err)
		}
		input_dir := filepath.Join(".", "www", "src")
		output_dir := filepath.Join(".", "www", "build")
		os.RemoveAll(output_dir)
		os.Mkdir(output_dir, fs.ModeDir|0755)
		build_website(cm_config, input_dir, output_dir)
		index_html, err := filepath.Abs(filepath.Join(output_dir, "index.html"))
		if err != nil {
			panic(err)
		}
		open.Run("file://" + index_html)
		watch_for_changes(cm_config, input_dir, output_dir)
	default:
		fmt.Println("expected 'init' or 'webdev' subcommands")
		os.Exit(1)
	}
}
