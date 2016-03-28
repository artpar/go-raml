package codegen

import (
	"strings"

	"github.com/Jumpscale/go-raml/raml"
)

type methodInterface interface {
	Verb() string
	Resource() *raml.Resource
}

// Method defines base Method struct
type Method struct {
	*raml.Method
	MethodName   string
	Endpoint     string
	verb         string
	ReqBody      string         // request body type
	RespBody     string         // response body type
	ResourcePath string         // normalized resource path
	resource     *raml.Resource // resource object of this method
	Params       string         // methods params
	FuncComments []string
	SecuredBy    []raml.DefinitionChoice
}

func (m Method) Verb() string {
	return m.verb
}

func (m Method) Resource() *raml.Resource {
	return m.resource
}

func newMethod(r *raml.Resource, rd *resourceDef, m *raml.Method, methodName, parentEndpoint, curEndpoint string) Method {
	method := Method{
		Method:   m,
		Endpoint: parentEndpoint + curEndpoint,
		verb:     strings.ToUpper(methodName),
		resource: r,
	}

	// set request body
	method.ReqBody = assignBodyName(m.Bodies, normalizeURITitle(method.Endpoint)+methodName, "ReqBody")

	//set response body
	for k, v := range m.Responses {
		if k >= 200 && k < 300 {
			method.RespBody = assignBodyName(v.Bodies, normalizeURITitle(method.Endpoint)+methodName, "RespBody")
		}
	}

	// set func comment
	if len(m.Description) > 0 {
		method.FuncComments = commentBuilder(m.Description)
	}

	return method
}

func newServerMethod(apiDef *raml.APIDefinition, r *raml.Resource, rd *resourceDef, m *raml.Method,
	methodName, parentEndpoint, curEndpoint, lang string) methodInterface {

	method := newMethod(r, rd, m, methodName, parentEndpoint, curEndpoint)

	// security scheme
	switch {
	case len(m.SecuredBy) > 0: // use secured by from this method
		method.SecuredBy = m.SecuredBy
	case len(r.SecuredBy) > 0: // use securedby from resource
		method.SecuredBy = r.SecuredBy
	default:
		method.SecuredBy = apiDef.SecuredBy // use secured by from root document
	}

	switch lang {
	case langGo:
		gm := goServerMethod{
			Method: &method,
		}
		gm.setup(apiDef, r, rd, methodName)
		return gm
	case langPython:
		pm := pythonServerMethod{
			Method: &method,
		}
		pm.setup(apiDef, r, rd)
		return pm
	default:
		panic("invalid language:" + lang)
	}
}

func newClientMethod(r *raml.Resource, rd *resourceDef, m *raml.Method, methodName, parentEndpoint, curEndpoint, lang string) (methodInterface, error) {
	method := newMethod(r, rd, m, methodName, parentEndpoint, curEndpoint)

	method.ResourcePath = paramizingURI(method.Endpoint)

	name := normalizeURITitle(method.Endpoint)

	method.ReqBody = assignBodyName(m.Bodies, name+methodName, "ReqBody")

	switch lang {
	case langGo:
		gcm := goClientMethod{Method: &method}
		err := gcm.setup(methodName)
		return gcm, err
	case langPython:
		pcm := pythonClientMethod{Method: method}
		pcm.setup()
		return pcm, nil
	default:
		panic("invalid language:" + lang)

	}
}

type goServerMethod struct {
	*Method
	Middlewares string
}

// setup go server method, initializes all needed variables
func (gm *goServerMethod) setup(apiDef *raml.APIDefinition, r *raml.Resource, rd *resourceDef, methodName string) error {
	// set method name
	name := normalizeURI(gm.Endpoint)
	if len(gm.DisplayName) > 0 {
		gm.MethodName = strings.Replace(gm.DisplayName, " ", "", -1)
	} else {
		gm.MethodName = name[len(rd.Name):] + methodName
	}

	// setting middlewares
	middlewares := []string{}

	// security middlewares
	for _, v := range gm.SecuredBy {
		if !validateSecurityScheme(v.Name, apiDef) {
			continue
		}
		// oauth2 middleware
		m, err := getOauth2MwrHandler(v)
		if err != nil {
			return err
		}
		middlewares = append(middlewares, m)
	}

	gm.Middlewares = strings.Join(middlewares, ", ")

	return nil
}

type goClientMethod struct {
	*Method
}

func (gcm *goClientMethod) setup(methodName string) error {
	// build func/method params
	buildParams := func(r *raml.Resource, bodyType string) (string, error) {
		paramsStr := strings.Join(getResourceParams(r), ",")
		if len(paramsStr) > 0 {
			paramsStr += " string"
		}

		// append request body type
		if len(bodyType) > 0 {
			if len(paramsStr) > 0 {
				paramsStr += ", "
			}
			paramsStr += strings.ToLower(bodyType) + " " + bodyType
		}

		// append header
		if len(paramsStr) > 0 {
			paramsStr += ","
		}
		paramsStr += "headers,queryParams map[string]interface{}"

		return paramsStr, nil
	}

	// method name
	name := normalizeURITitle(gcm.Endpoint)

	if len(gcm.DisplayName) > 0 {
		gcm.MethodName = strings.Replace(gcm.DisplayName, " ", "", -1)
	} else {
		gcm.MethodName = strings.Title(name + methodName)
	}

	// method param
	methodParam, err := buildParams(gcm.resource, gcm.ReqBody)
	if err != nil {
		return err
	}
	gcm.Params = methodParam

	return nil
}

// assignBodyName assign method's request body by bodies.Type or bodies.ApplicationJson
// if bodiesType generated from bodies.Type we dont need append prefix and suffix
// 		example : bodies.Type = City, so bodiesType = City
// if bodiesType generated from bodies.ApplicationJson, we get that value from prefix and suffix
//		suffix = [ReqBody | RespBody] and prefix should be uri + method name.
//		example prefix could be UsersUserIdDelete
func assignBodyName(bodies raml.Bodies, prefix, suffix string) string {
	var bodiesType string

	if len(bodies.Type) > 0 {
		bodiesType = convertToGoType(bodies.Type)
	} else if bodies.ApplicationJson != nil {
		if bodies.ApplicationJson.Type != "" {
			bodiesType = convertToGoType(bodies.ApplicationJson.Type)
		} else {
			bodiesType = prefix + suffix
		}
	}

	return bodiesType
}
