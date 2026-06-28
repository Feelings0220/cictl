package jenkins

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"
)

type Cloud struct {
	Name string `json:"name"`
	Kind string `json:"kind"` // simplified class name, e.g. "KubernetesCloud"
}

func (c *Client) ListClouds(ctx context.Context) ([]Cloud, error) {
	body, err := c.GET(ctx, "/config.xml", nil)
	if err != nil {
		return nil, fmt.Errorf("list clouds (root config): %w", err)
	}
	return parseCloudList(body)
}

func (c *Client) GetCloudConfig(ctx context.Context, name string) ([]byte, error) {
	body, err := c.GET(ctx, "/manage/cloud/"+name+"/config.xml", nil)
	if err != nil {
		return nil, fmt.Errorf("get cloud %s config: %w", name, err)
	}
	return body, nil
}

// parseCloudList walks <clouds> children, extracting <name>...</name> from each.
func parseCloudList(body []byte) ([]Cloud, error) {
	s := string(body)
	// Go's encoding/xml only supports XML 1.0; Jenkins emits version='1.1'.
	// Strip the XML processing instruction so the decoder can proceed.
	if trimmed := strings.TrimLeft(s, " \t\r\n"); strings.HasPrefix(trimmed, "<?xml") {
		if idx := strings.Index(s, "?>"); idx >= 0 {
			s = s[idx+2:]
		}
	}
	dec := xml.NewDecoder(strings.NewReader(s))
	var clouds []Cloud
	inClouds := false
	var current *Cloud
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "clouds" {
				inClouds = true
				continue
			}
			if inClouds && current == nil {
				// First child element inside <clouds> = a cloud entry
				current = &Cloud{Kind: simpleKind(t.Name.Local)}
				continue
			}
			if current != nil && t.Name.Local == "name" {
				var name string
				if err := dec.DecodeElement(&name, &t); err == nil {
					current.Name = name
				}
			}
		case xml.EndElement:
			if t.Name.Local == "clouds" {
				inClouds = false
			}
			if current != nil && t.Name.Local != "name" && simpleKind(t.Name.Local) == current.Kind {
				clouds = append(clouds, *current)
				current = nil
			}
		}
	}
	return clouds, nil
}

// simpleKind extracts the trailing class name (after the last dot).
func simpleKind(full string) string {
	if i := strings.LastIndex(full, "."); i >= 0 {
		return full[i+1:]
	}
	return full
}
