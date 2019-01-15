package csp

import (
	"fmt"
	"sort"
	"strings"
)

const (
	DispositionEnforce = "enforce"
	DispositionReport  = "report"

	ContentSecurityPolicy           = "Content-Security-Policy"
	ContentSecurityPolicyReportOnly = "Content-Security-Policy-Report-Only"
)

type Policy struct {
	Disposition string
	Directives  Directives
}

type Directives []Directive

func (d Directives) String() string {
	o := make([]string, 0)
	for _, v := range d {
		o = append(o, v.String())
	}
	return strings.Join(o, "; ")
}

type Directive struct {
	Name  string
	Value []string
}

func (d Directive) String() string {
	return d.Name + " " + strings.Join(d.Value, " ")
}

var (
	ErrDuplicateDirective      = fmt.Errorf("duplicate directive")
	ErrDirectiveNameUnknown    = fmt.Errorf("unknown directive name")
	ErrDirectiveNameDeprecated = fmt.Errorf("deprecated directive name")
)

func ParseDirectives(serializedPolicy string) (Directives, error) {
	d := make(Directives, 0)

	// For each token returned by strictly splitting serialized CSP
	// on the U+003B SEMICOLON character (;):
	tokens := strings.Split(serializedPolicy, ";")
	for _, t := range tokens {
		// Strip leading and trailing whitespace from token.
		t = strings.TrimSpace(t)

		// If token is an empty string, skip the remaining substeps
		// and continue to the next item.
		if len(t) == 0 {
			continue
		}

		// Let directive name be the result of collecting
		// a sequence of characters from token which are not space characters.
		x := strings.SplitN(t, " ", 2)

		// The name is a non-empty string
		if len(x) == 0 || len(x[0]) == 0 {
			continue
		}
		name := x[0]

		// Verify name
		// see also https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy
		// see also https://www.w3.org/TR/CSP3/#csp-directives
		switch strings.ToLower(name) {
		case "child-src":
		case "connect-src":
		case "default-src":
		case "font-src":
		case "frame-src":
		case "img-src":
		case "manifest-src":
		case "media-src":
		case "prefetch-src":
		case "object-src":
		case "script-src":
		case "script-src-elem":
		case "script-src-attr":
		case "style-src":
		case "style-src-elem":
		case "style-src-attr":
		case "worker-src":
		case "base-uri":
		case "plugin-types":
		case "sandbox":
		case "disown-opener":
		case "form-action":
		case "frame-ancestors":
		case "navigate-to":
		case "report-uri":
		case "report-to":
		case "upgrade-insecure-requests":
		case "block-all-mixed-content":
		case "require-sri-for":
			// ok

		case "reflected-xss":
			// ok, deprecated from CSP 2

		case "referrer":
			// ok, deprecated, use Referrer-Policy header

		case "policy-uri":
			return nil, &ParseError{
				Err:    ErrDirectiveNameDeprecated,
				Custom: "policy-uri has been removed and is not supported",
			}

		default:
			return nil, &ParseError{
				Err:    ErrDirectiveNameUnknown,
				Custom: fmt.Sprintf("directive name '%v' is unknown", name),
			}
		}

		// If the set of directives already contains a directive
		// whose name is a case insensitive match for directive name,
		// ignore this instance of the directive and continue to the next token.
		// The user agent SHOULD notify developers that a directive was ignored.
		for _, dx := range d {
			if strings.ToLower(dx.Name) == strings.ToLower(name) {
				return nil, &ParseError{
					Err:    ErrDuplicateDirective,
					Custom: fmt.Sprintf("directive '%v' is a duplicate", name),
				}
			}
		}

		// The value is a set of non-empty strings. The value set MAY be empty.
		values := make([]string, 0)
		if len(x) > 1 {
			for _, v := range strings.Split(x[1], " ") {
				if len(v) > 0 {
					values = append(values, v)
				}
			}
		}

		// Add directive to directive set.
		d = append(d, Directive{
			Name:  name,
			Value: values,
		})
	}

	return d, nil
}

func (d *Directives) AddDirective(v Directive) error {
	// add values to existing directive if already exists
	added := false
	for i := 0; i < len(*d); i++ {
		if (*d)[i].Name == v.Name {
			added = true
			var valmap = map[string]struct{}{}

			for _, val := range v.Value {
				valmap[val] = struct{}{}
			}

			for _, val := range (*d)[i].Value {
				valmap[val] = struct{}{}
			}

			var sorted sort.StringSlice
			for val := range valmap {
				sorted = append(sorted, val)
			}
			sorted.Sort()
			(*d)[i].Value = sorted
		}
	}

	// ... or add new directive
	if !added {
		var sorted = sort.StringSlice(v.Value)
		sorted.Sort()
		*d = append(*d, v)
	}

	return nil
}

func (d *Directives) RemoveDirectiveByName(name string) {
	x := make(Directives, 0)
	for _, v := range *d {
		if v.Name != name {
			x = append(x, v)
		}
	}
	*d = x
}
