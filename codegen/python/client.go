package python

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Jumpscale/go-raml/codegen/commons"
	"github.com/Jumpscale/go-raml/codegen/resource"
	"github.com/Jumpscale/go-raml/codegen/security"
	"github.com/Jumpscale/go-raml/raml"
)

var (
	globAPIDef *raml.APIDefinition
)

// Client represents a python client
type Client struct {
	Name     string
	APIDef   *raml.APIDefinition
	BaseURI  string
	Services map[string]*service
}

// NewClient creates a python Client
func NewClient(apiDef *raml.APIDefinition) Client {
	services := map[string]*service{}
	for k, v := range apiDef.Resources {
		rd := resource.New(apiDef, commons.NormalizeURITitle(apiDef.Title), "")
		rd.GenerateMethods(&v, "python", newServerMethod, newClientMethod)
		services[k] = &service{
			rootEndpoint: k,
			Methods:      rd.Methods,
		}
	}
	c := Client{
		Name:     commons.NormalizeURI(apiDef.Title),
		APIDef:   apiDef,
		BaseURI:  apiDef.BaseURI,
		Services: services,
	}
	if strings.Index(c.BaseURI, "{version}") > 0 {
		c.BaseURI = strings.Replace(c.BaseURI, "{version}", apiDef.Version, -1)
	}
	return c
}

// generate empty __init__.py without overwrite it
func generateEmptyInitPy(dir string) error {
	return commons.GenerateFile(nil, "./templates/init_py.tmpl", "init_py", filepath.Join(dir, "__init__.py"), false)
}

// Generate generates python client library files
func (c Client) Generate(dir string) error {
	globAPIDef = c.APIDef

	// generate helper
	if err := commons.GenerateFile(nil, "./templates/client_utils_python.tmpl", "client_utils_python", filepath.Join(dir, "client_utils.py"), false); err != nil {
		return err
	}

	if err := c.generateServices(dir); err != nil {
		return err
	}

	if err := c.generateSecurity(dir); err != nil {
		return err
	}

	if err := c.generateInitPy(dir); err != nil {
		return err
	}
	// generate main client lib file
	return commons.GenerateFile(c, "./templates/client_python.tmpl", "client_python", filepath.Join(dir, "client.py"), true)
}

func (c Client) generateServices(dir string) error {
	for _, s := range c.Services {
		sort.Sort(resource.ByEndpoint(s.Methods))
		if err := commons.GenerateFile(s, "./templates/client_service_python.tmpl", "client_service_python", s.filename(dir), false); err != nil {
			return err
		}
	}
	return nil
}

func (c Client) generateSecurity(dir string) error {
	for name, ss := range c.APIDef.SecuritySchemes {
		if !security.Supported(ss) {
			continue
		}
		ctx := map[string]string{
			"Name":           oauth2ClientName(name),
			"AccessTokenURI": fmt.Sprintf("%v", ss.Settings["accessTokenUri"]),
		}
		filename := filepath.Join(dir, oauth2ClientFilename(name))
		if err := commons.GenerateFile(ctx, "./templates/oauth2_client_python.tmpl", "oauth2_client_python", filename, true); err != nil {
			return err
		}
	}
	return nil
}

func (c Client) generateInitPy(dir string) error {
	type oauth2Client struct {
		Name       string
		ModuleName string
		Filename   string
	}

	var securities []oauth2Client

	for name, ss := range c.APIDef.SecuritySchemes {
		if !security.Supported(ss) {
			continue
		}
		s := oauth2Client{
			Name:     oauth2ClientName(name),
			Filename: oauth2ClientFilename(name),
		}
		s.ModuleName = strings.TrimSuffix(s.Filename, ".py")
		securities = append(securities, s)
	}
	ctx := map[string]interface{}{
		"BaseURI":    c.APIDef.BaseURI,
		"Securities": securities,
	}
	filename := filepath.Join(dir, "__init__.py")
	return commons.GenerateFile(ctx, "./templates/client_initpy_python.tmpl", "client_initpy_python", filename, false)
}
