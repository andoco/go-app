package app

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadEnvConfig(t *testing.T) {
	c := &struct {
		Foo string
	}{}

	os.Setenv("MY_APP_FOO", "Foo")
	defer os.Unsetenv("MY_APP_FOO")

	err := ReadEnvConfig(c, "My", "App")
	assert.Nil(t, err)
	assert.Equal(t, "Foo", c.Foo)
}

func TestBuildEnvConfigName(t *testing.T) {
	assert.Equal(t, "MYAPP_FOO_BAR", BuildEnvConfigName("MyApp", "foo", "BAR"))
}

func TestValidateAppName(t *testing.T) {
	testCases := []struct {
		name     string
		inName   string
		outValid bool
	}{
		{name: "single camel", inName: "Foo", outValid: true},
		{name: "two camel", inName: "FooBar", outValid: true},
		{name: "three camel", inName: "FooBarBaz", outValid: true},
		{name: "adjacent uppercase", inName: "FBar", outValid: true},
		{name: "only uppercase", inName: "FOOBAR", outValid: true},
		{name: "only lowercase", inName: "foo", outValid: false},
		{name: "snake", inName: "foo_bar", outValid: false},
		{name: "kebab", inName: "foo-bar", outValid: false},
		{name: "spaces", inName: "foo bar", outValid: false},
		{name: "mixed", inName: "foo-bar_baz blah", outValid: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.outValid, validateAppName(tc.inName))
		})
	}
}

func TestSplitUpperCamelCase(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		out  []string
	}{
		{name: "empty string", in: "", out: []string{}},
		{name: "one letter", in: "F", out: []string{"F"}},
		{name: "one word", in: "Foo", out: []string{"Foo"}},
		{name: "two words", in: "FooBar", out: []string{"Foo", "Bar"}},
		{name: "lower two words", in: "fooBar", out: []string{"foo", "Bar"}},
		{name: "three words", in: "FooBarBaz", out: []string{"Foo", "Bar", "Baz"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.out, splitUpperCamelCase(tc.in))
		})
	}
}
