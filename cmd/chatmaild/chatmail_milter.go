package main

import (
	"mime"
	"mime/multipart"

	"github.com/s0ph0s-2/gochatmail/internal/config"

	"encoding/base64"
    "fmt"
	"io"
	"log"
	"net"
	"net/textproto"
    "net/mail"
	"slices"
	"strings"

	"github.com/emersion/go-milter"
)

type milter_server struct {
    server milter.Server
    listener net.Listener
}

func new_milter_server(listen_uri string) (milter_server, error) {
	server := milter.Server{
		NewMilter: func() milter.Milter {
			return &ChatmailMilter{}
		},
		Protocol: milter.OptNoConnect | milter.OptNoHelo,
	}
	ln, err := make_listener(listen_uri)
	if err != nil {
        return milter_server{}, fmt.Errorf("Failed to set up listener for milter: %q", err)
	}

	log.Printf("Using %s as milter listen socket\n", listen_uri)
    return milter_server{server, ln}, nil
}

func (ms *milter_server) serve() error {
    err := ms.server.Serve(ms.listener)
    if err != nil && err != milter.ErrServerClosed {
        log.Fatal("Failed to start milter: ", err)
    }
    return err
}

func (ms *milter_server) stop() error {
    return ms.server.Close()
}

type ChatmailMilter struct {
	mailFrom      string
    mimeFrom      string
	rcptTos       []string
	secureJoinHdr string
	subject       string
	content_type  string
	body          io.ReadWriter
	config        config.ChatmailConfig
}

// MARK: milter interface functions

func (cm *ChatmailMilter) Abort(m *milter.Modifier) error {
	return nil
}

func (cm *ChatmailMilter) Connect(host string, family string, port uint16, addr net.IP, m *milter.Modifier) (milter.Response, error) {
	return nil, nil
}

func (cm *ChatmailMilter) Helo(name string, m *milter.Modifier) (milter.Response, error) {
	return nil, nil
}

func (cm *ChatmailMilter) MailFrom(from string, m *milter.Modifier) (milter.Response, error) {
	cm.mailFrom = from
	return nil, nil
}

func (cm *ChatmailMilter) RcptTo(rcptTo string, m *milter.Modifier) (milter.Response, error) {
	cm.rcptTos = append(cm.rcptTos, rcptTo)
	return nil, nil
}

func (cm *ChatmailMilter) Header(name string, value string, m *milter.Modifier) (milter.Response, error) {
	if strings.EqualFold(name, "secure-join") {
		cm.secureJoinHdr = value
	} else if strings.EqualFold(name, "content-type") {
		cm.content_type = value
	} else if strings.EqualFold(name, "subject") {
		cm.subject = value
	} else if strings.EqualFold(name, "from") {
        cm.mimeFrom = value
    }
	return milter.RespContinue, nil
}

func (cm *ChatmailMilter) Headers(h textproto.MIMEHeader, m *milter.Modifier) (milter.Response, error) {
	return milter.RespContinue, nil
}

func (cm *ChatmailMilter) Body(m *milter.Modifier) (milter.Response, error) {
	log.Printf("milter_state: %v", cm)
	return milter.RespAccept, nil
}

func (cm *ChatmailMilter) BodyChunk(chunk []byte, m *milter.Modifier) (milter.Response, error) {
    _, err := cm.body.Write(chunk)
    if err != nil {
        return nil, err
    }
	return nil, nil
}

// MARK: testable logic functions

func (cm *ChatmailMilter) ValidateEmail() (milter.Response, error) {
	if slices.Contains(cm.config.PassthroughSendersList, cm.mailFrom) {
		return milter.RespAccept, nil
	}
	mail_encrypted, err := IsValidEncryptedMessage(
		cm.subject,
		cm.content_type,
		cm.body,
	)
	if err != nil {
		return nil, err
	}
    mime_from_addr, err := mail.ParseAddress(cm.mimeFrom)
    if err != nil {
        return nil, err
    }
    if !strings.EqualFold(mime_from_addr.Address, cm.mailFrom) {
        return milter.RespReject, nil
    }
	mime_from_parts := strings.Split(mime_from_addr.Address, "@")
	mime_from_domain := mime_from_parts[len(mime_from_parts)-1]
	for _, recipient := range cm.rcptTos {
		if cm.mailFrom == recipient {
			continue
		}
		if slices.Contains(cm.config.PassthroughRecipientsList, recipient) {
			continue
		}
		res := strings.Split(recipient, "@")
		if len(res) != 2 {
			return milter.RespReject, nil
		}
		recipient_domain := res[len(res)-1]
		is_outgoing := recipient_domain != mime_from_domain
		if is_outgoing && !mail_encrypted {
			is_securejoin := strings.EqualFold(cm.secureJoinHdr, "vc-request") || strings.EqualFold(cm.secureJoinHdr, "vg-request")
			if !is_securejoin {
				return milter.RespReject, nil
			}
		}
	}
	return milter.RespAccept, nil
}

func IsEncryptedOpenPGPPayload(payload []byte) bool {
	i := 0
	for i < len(payload) {
        // Permit only OpenPGP formatted binary data.
		if payload[i]&0xC0 != 0xC0 {
			return false
		}
		packet_type_id := payload[i] & 0x3F
		i += 1
		var body_len int
		if payload[i] < 192 {
			body_len = int(payload[i])
			i += 1
		} else if payload[i] < 224 {
            if (i + 1) >= len(payload) {
                return false
            }
			body_len = ((int(payload[i]) - 192) << 8) + int(payload[i+1]) + 192
            i += 2
		} else if payload[i] == 255 {
            if (i + 4) >= len(payload) {
                return false
            }
			body_len = (int(payload[i+1]) << 24) | (int(payload[i+2]) << 16) | (int(payload[i+3]) << 8) | int(payload[i+4])
			i += 5
		} else {
			return false
		}
		i += body_len
		if i == len(payload) {
            // The last packet in the stream should be
            // "Symmetrically Encrypted and Integrity Protected Data Packet
            // (SEIDP)".
            // This is the only place in this function that is allowed to return
            // true.
			return packet_type_id == 18
		} else if packet_type_id != 1 && packet_type_id != 3 {
			return false
		}
	}
    return false
}

func IsValidEncryptedPayload(payload string) bool {
	const header = "-----BEGIN PGP MESSAGE-----\r\n\r\n"
	const footer = "-----END PGP MESSAGE-----\r\n\r\n"
	hasHeader := strings.HasPrefix(payload, header)
	hasFooter := strings.HasSuffix(payload, footer)
	if !(hasHeader && hasFooter) {
		return false
	}
	start_idx := len(header)
	crc24_start := strings.LastIndex(payload, "=")
	var end_idx int
	if crc24_start < 0 {
		end_idx = len(payload) - len(footer)
	} else {
		end_idx = crc24_start
	}
	b64_encoded := payload[start_idx:end_idx]
	b64_decoded := make([]byte, base64.StdEncoding.DecodedLen(len(b64_encoded)))
	n, err := base64.StdEncoding.Decode(b64_decoded, []byte(b64_encoded))
	if err != nil {
		return false
	}
	b64_decoded = b64_decoded[:n]
	return IsEncryptedOpenPGPPayload(b64_decoded)
}

func IsValidEncryptedMessage(subject string, content_type string, body io.Reader) (bool, error) {
	if !slices.Contains(CommonEncryptedSubjects, subject) {
		return false, nil
	}
	mediatype, params, err := mime.ParseMediaType(content_type)
	if err != nil {
		return false, err
	}
	if mediatype != "multipart/encrypted" {
		return false, nil
	}
	mpr := multipart.NewReader(body, params["boundary"])
	// TODO: figure out how to/whether it's necessary to decode non-UTF-8 encodings
	parts_count := 0
	for {
		part, err := mpr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, err
		}
		if parts_count == 0 {
			part_content_type := part.Header.Get("Content-Type")
			if part_content_type != "application/pgp-encrypted" {
				return false, nil
			}
			part_body, err := io.ReadAll(part)
			if err != nil {
				return false, err
			}
			if strings.TrimSpace(string(part_body)) != "Version: 1" {
				return false, nil
			}
		} else if parts_count == 1 {
			part_content_type := part.Header.Get("Content-Type")
			if !strings.HasPrefix(part_content_type, "application/octet-stream") {
				return false, nil
			}
			part_body, err := io.ReadAll(part)
			if err != nil {
				return false, err
			}
			if !IsValidEncryptedPayload(string(part_body)) {
				return false, nil
			}
		} else {
			return false, nil
		}
		parts_count += 1
	}
	return true, nil
}
