package replay

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
)

const statesByteTag = "statesByte"

func ExtractReplayBinaryData(in io.Reader) (io.Reader, error) {
	return extractBase64TagContent(in, statesByteTag)
}

func extractBase64TagContent(xmlr io.Reader, tag string) (io.Reader, error) {
	dec := xml.NewDecoder(xmlr)
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil, fmt.Errorf("%s not found", tag)
		}
		if err != nil {
			return nil, fmt.Errorf("xml decode: %w", err)
		}

		start, ok := tok.(xml.StartElement)
		if !ok || start.Name.Local != tag {
			continue
		}

		pr, pw := io.Pipe()

		go func() {
			defer pw.Close()

			depth := 1
			for depth > 0 {
				tok, err := dec.Token()
				if err != nil {
					pw.CloseWithError(fmt.Errorf("xml decode inside %s: %w", tag, err))
					return
				}
				switch x := tok.(type) {
				case xml.StartElement:
					depth++
				case xml.EndElement:
					depth--
				case xml.CharData:
					buf := filterBase64Whitespace([]byte(x))
					if len(buf) == 0 {
						continue
					}
					if _, err = pw.Write(buf); err != nil {
						return
					}

				}
			}
		}()
		return base64.NewDecoder(base64.StdEncoding, pr), nil
	}
}

func filterBase64Whitespace(b []byte) []byte {
	for _, c := range b {
		if c == ' ' || c == '\n' || c == '\r' || c == '\t' {
			out := make([]byte, 0, len(b))
			for _, c2 := range b {
				switch c2 {
				case ' ', '\n', '\r', '\t':
					continue
				default:
					out = append(out, c2)
				}
			}
			return out
		}
	}
	return b
}
