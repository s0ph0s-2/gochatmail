package main

import (
	"encoding/json"
	"flag"
	"fmt"
    "io/fs"
	"os"
    "strings"
    "bytes"
    "github.com/yuin/goldmark"
    "path/filepath"
    "html/template"
)

type chatmail_config struct {
	MailFullyQualifiedDomainName string
    MaxEmailsPerMinutePerUser int
    MaxMailboxSizeMB int
    MaxMessageSizeB int
    DeleteMailsAfterDays int
    DeleteInactiveUsersAfterDays int
    UsernameMinLength int
    UsernameMaxLength int
    PasswordMinLength int
    PassthroughRecipientsList []string
    PrivacyContactPostalAddress string
    PrivacyContactEmailAddress string
    PrivacyDataOfficerPostalAddress string
    PrivacySupervisorPostalAddress string
}

func NewChatmailConfig(fqdn string) chatmail_config {
    return chatmail_config{
        fqdn,
        30,
        100,
        31457280,
        20,
        90,
        9,
        9,
        9,
        []string{"xstore@testrun.org"},
        "",
        "",
        "",
        "",
    }
}

func (config chatmail_config) Save(filename string) error {
	output_txt, m_err := json.MarshalIndent(config, "", "  ")
	if m_err != nil {
        return m_err
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	f.Write(output_txt)
    return nil
}

func LoadChatmailConfigFromFile(filename string, config *chatmail_config) error {
    data, r_err := os.ReadFile(filename)
    if r_err != nil {
        return r_err
    }
    j_err := json.Unmarshal(data, config)
    if j_err != nil {
        return j_err
    }
    return nil
}

func do_init(fqdn string) {
	config := NewChatmailConfig(fqdn)
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
    Title string
    Config chatmail_config
}

func build_website(config chatmail_config, input_dir string, output_dir string) {
    page_layout_file := filepath.Join(input_dir, "page-layout.html")
    templates, tmpl_err := template.New("page_layout").ParseFiles(page_layout_file)
    if tmpl_err != nil {
        panic(tmpl_err)
    }
    contents, rd_err := os.ReadDir(input_dir)
    if rd_err != nil {
        panic(rd_err)
    }
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
            if err := goldmark.Convert(page_content_md, &md_buf); err != nil {
                panic(err)
            }
            _, err = local_tmpls.New("PageContent").Parse(md_buf.String())
            if err != nil {
                panic(err)
            }
            this_page_vars := page_vars{ page_name, config }
            var html_buf bytes.Buffer
            err = local_tmpls.ExecuteTemplate(&html_buf, "page-layout.html", this_page_vars)
            if err != nil {
                panic(err)
            }
            output_file := filepath.Join(output_dir, stem + ".html")
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

func watch_for_changes() {
	// TODO: this
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
        var config chatmail_config
        config_file := filepath.Join(".", "chatmail.json")
        cl_err := LoadChatmailConfigFromFile(config_file, &config)
        if cl_err != nil {
            panic(cl_err)
        }
        input_dir := filepath.Join(".", "www", "src")
        output_dir := filepath.Join(".", "www", "build")
        os.RemoveAll(output_dir)
        os.Mkdir(output_dir, fs.ModeDir | 0755)
		build_website(config, input_dir, output_dir)
		watch_for_changes()
	default:
		fmt.Println("expected 'init' or 'webdev' subcommands")
		os.Exit(1)
	}
}
