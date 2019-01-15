package csp

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		in               string
		expectDirectives []Directive
		expectErr        error
	}{
		// test some basic stuff
		{
			in: "  default-src  'self'  ;  script-src 'self'; connect-src ; object-src 'self';base-uri 'none';report-uri https://logs.templarbit.com/csp/foobar/reports;  ",
			expectDirectives: []Directive{
				{
					Name:  "default-src",
					Value: []string{"'self'"},
				},
				{
					Name:  "script-src",
					Value: []string{"'self'"},
				},
				{
					Name:  "connect-src",
					Value: []string{},
				},
				{
					Name:  "object-src",
					Value: []string{"'self'"},
				},
				{
					Name:  "base-uri",
					Value: []string{"'none'"},
				},
				{
					Name:  "report-uri",
					Value: []string{"https://logs.templarbit.com/csp/foobar/reports"},
				},
			},
			expectErr: nil,
		},

		// test duplicate directive
		{
			in:               "object-src 'self'; object-src 'none'",
			expectDirectives: nil,
			expectErr:        ErrDuplicateDirective,
		},
		{
			in:               "object-src 'self'; Object-src 'none'",
			expectDirectives: nil,
			expectErr:        ErrDuplicateDirective,
		},

		// test unknown directive name
		{
			in:               "bogus 'self'",
			expectDirectives: nil,
			expectErr:        ErrDirectiveNameUnknown,
		},
	}

	for i, tt := range tests {
		out, err := ParseDirectives(tt.in)
		if err != tt.expectErr {
			if xerr, ok := err.(*ParseError); ok {
				if xerr.Err != tt.expectErr {
					t.Fatalf("expect %v, got %v, in %v", tt.expectErr, xerr, i)
				}
			} else {
				t.Fatalf("expect %v, got %v, in %v", tt.expectErr, err, i)
			}
		}

		if len(out) != len(tt.expectDirectives) {
			t.Fatalf("expect len(%v), got len(%v), in %v", len(tt.expectDirectives), len(out), i)
		}

		for j := 0; j < len(out); j++ {
			if out[j].Name != tt.expectDirectives[j].Name {
				t.Errorf("expect %v, got %v, in %v", tt.expectDirectives[j].Name, out[j].Name, i)
			}

			if len(out[j].Value) != len(tt.expectDirectives[j].Value) {
				t.Fatalf("expect len(%v), got len(%v), in %v", len(tt.expectDirectives[j].Value), len(out[j].Value), i)
			}

			for k := 0; k < len(out[j].Value); k++ {
				if out[j].Value[k] != tt.expectDirectives[j].Value[k] {
					t.Errorf("expect %v, got %v, in %v", tt.expectDirectives[j].Value[k], out[j].Value[k], i)
				}
			}
		}
	}
}

func TestDirectivesToString(t *testing.T) {
	e := "default-src 'self'; script-src 'self'; object-src 'self' http://; base-uri 'none'; report-uri https://logs.templarbit.com/csp/xxkey/reports"
	directives, err := ParseDirectives(e)
	if err != nil {
		t.Fatal(err)
	}

	o := directives.String()
	if o != e {
		t.Errorf("\nexpected: %v\n     got: %v", e, o)
	}
}

func TestAddDirective(t *testing.T) {
	d := make(Directives, 0)
	d.AddDirective(Directive{"foo", []string{"b", "a", "r"}})
	d.AddDirective(Directive{"oof", []string{"r", "a", "b"}})
	d.AddDirective(Directive{"foo", []string{"1", "2", "a"}})

	result := d.String()
	expect := "foo 1 2 a b r; oof a b r"
	if result != expect {
		t.Errorf("expected %v, got %v", expect, result)
	}
}

func TestRemoveDirective(t *testing.T) {
	e := "default-src 'self'; script-src 'self'; object-src 'self' http://; base-uri 'none'; report-uri https://logs.templarbit.com/csp/xxkey/reports"
	directives, err := ParseDirectives(e)
	if err != nil {
		t.Fatal(err)
	}

	directives.RemoveDirectiveByName("script-src")
	directives.RemoveDirectiveByName("report-uri")
	directives.RemoveDirectiveByName("object-src")

	result := directives.String()
	expect := "default-src 'self'; base-uri 'none'"
	if result != expect {
		t.Errorf("expected %v, got %v", expect, result)
	}
}
