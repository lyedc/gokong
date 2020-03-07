package gokong

import (
	"encoding/json"
	"fmt"
	"github.com/parnurzeal/gorequest"
)

type PluginClient struct {
	config *Config
}

type PluginRequest struct {
	Name       string                 `json:"name"`
	ApiId      string                 `json:"api_id,omitempty"`
	ConsumerId string                 `json:"consumer_id,omitempty"`
	Config     map[string]interface{} `json:"config,omitempty"`
}

type Plugin struct {
	Id         string                 `json:"id"`
	Name       string                 `json:"name"`
	ApiId      string                 `json:"api_id,omitempty"`
	ConsumerId string                 `json:"consumer_id,omitempty"`
	Config     map[string]interface{} `json:"config,omitempty"`
	Enabled    bool                   `json:"enabled,omitempty"`
}

type Plugins struct {
	Results []*Plugin `json:"data,omitempty"`
	Total   int       `json:"total,omitempty"`
	Next    string    `json:"next,omitempty"`
}

type PluginFilter struct {
	Id         string `url:"id,omitempty"`
	Name       string `url:"name,omitempty"`
	ApiId      string `url:"api_id,omitempty"`
	ConsumerId string `url:"consumer_id,omitempty"`
	Size       int    `url:"size,omitempty"`
	Offset     int    `url:"offset,omitempty"`
}

type PluginSchema struct {
	Fields       map[string]PluginSchemaField `json:"fields,omitempty"`
	ErrorMessage string                       `json:"message,omitempty"`
}

type PluginSchemaField struct {
	//Validates the type of a property.
	Type string `json:"type"`
	//Default: false. If true, the property must be present in the configuration.
	IsRequired bool `json:"required,omitempty"`
	//Default: false. If true, the value must be unique (see remark below).
	IsUnique bool `json:"unique,omitempty"`
	//If the property is not specified in the configuration, will set the property to the given value.
	DefaultValue interface{} `json:"default,omitempty"`
	//A function to perform any custom validation on a property. See later examples for its parameters and return values.
	Func string `json:"func,omitempty"`
}

func (psf *PluginSchemaField) HasDefaultValue() bool {
	return psf.DefaultValue != nil
}

const PluginsPath = "/plugins/"

func (pluginClient *PluginClient) GetById(id string) (*Plugin, error) {

	_, body, errs := gorequest.New().Get(pluginClient.config.HostAddress + PluginsPath + id).End()
	if errs != nil {
		return nil, fmt.Errorf("could not get plugin, error: %v", errs)
	}

	plugin := &Plugin{}
	err := json.Unmarshal([]byte(body), plugin)
	if err != nil {
		return nil, fmt.Errorf("could not parse plugin plugin response, error: %v", err)
	}

	if plugin.Id == "" {
		return nil, nil
	}

	return plugin, nil
}

func (pluginClient *PluginClient) List() (*Plugins, error) {
	return pluginClient.ListFiltered(nil)
}

//ListEnabled retrieves a list of enabled plugin names.
//Invited URL: /plugins/enabled
func (pluginClient *PluginClient) ListEnabled() ([]string, error) {
	_, body, errs := gorequest.New().Get(pluginClient.config.HostAddress + PluginsPath + "enabled").End()
	if errs != nil {
		return nil, fmt.Errorf("could not get plugins, error: %v", errs)
	}

	type enabledPlugins struct {
		Plugins []string `json:"enabled_plugins"`
	}
	rsp := enabledPlugins{}
	err := json.Unmarshal([]byte(body), &rsp)
	if err != nil {
		return nil, fmt.Errorf("could not parse enabled plugins list response, error: %v", err)
	}
	return rsp.Plugins, nil
}

func (pluginClient *PluginClient) ListFiltered(filter *PluginFilter) (*Plugins, error) {

	address, err := addQueryString(pluginClient.config.HostAddress+PluginsPath, filter)

	if err != nil {
		return nil, fmt.Errorf("could not build query string for plugins filter, error: %v", err)
	}

	_, body, errs := gorequest.New().Get(address).End()
	if errs != nil {
		return nil, fmt.Errorf("could not get plugins, error: %v", errs)
	}

	plugins := &Plugins{}
	err = json.Unmarshal([]byte(body), plugins)
	if err != nil {
		return nil, fmt.Errorf("could not parse plugins list response, error: %v", err)
	}

	return plugins, nil
}

func (pluginClient *PluginClient) Create(pluginRequest *PluginRequest) (*Plugin, error) {
	//BUG fixes: According to the given API ID to compose the final HTTP request address.
	httpPath := ""
	if pluginRequest.ApiId == "" {
		httpPath = pluginClient.config.HostAddress + PluginsPath
	} else {
		httpPath = fmt.Sprintf("%s/apis/%s%s", pluginClient.config.HostAddress, pluginRequest.ApiId, PluginsPath)
	}
	_, body, errs := gorequest.New().Put(httpPath).Send(pluginRequest).End()
	if errs != nil {
		return nil, fmt.Errorf("could not create new plugin, error: %v", errs)
	}

	createdPlugin := &Plugin{}
	err := json.Unmarshal([]byte(body), createdPlugin)
	if err != nil {
		return nil, fmt.Errorf("could not parse plugin creation response, error: %v kong response: %s", err, body)
	}

	if createdPlugin.Id == "" {
		return nil, fmt.Errorf("could not create plugin, err: %v", body)
	}

	return createdPlugin, nil
}

func (pluginClient *PluginClient) UpdateById(id string, pluginRequest *PluginRequest) (*Plugin, error) {

	_, body, errs := gorequest.New().Patch(pluginClient.config.HostAddress + PluginsPath + id).Send(pluginRequest).End()
	if errs != nil {
		return nil, fmt.Errorf("could not update plugin, error: %v", errs)
	}

	updatedPlugin := &Plugin{}
	err := json.Unmarshal([]byte(body), updatedPlugin)
	if err != nil {
		return nil, fmt.Errorf("could not parse plugin update response, error: %v kong response: %s", err, body)
	}

	if updatedPlugin.Id == "" {
		return nil, fmt.Errorf("could not update plugin, error: %v", body)
	}

	return updatedPlugin, nil
}

func (pluginClient *PluginClient) DeleteById(id string) error {

	res, _, errs := gorequest.New().Delete(pluginClient.config.HostAddress + PluginsPath + id).End()
	if errs != nil {
		return fmt.Errorf("could not delete plugin, result: %v error: %v", res, errs)
	}

	return nil
}

func (pluginClient *PluginClient) GetSchema(pluginName string) (*PluginSchema, error) {
	rsp, body, errs := gorequest.New().Get(fmt.Sprintf("%s/plugins/schema/%s", pluginClient.config.HostAddress, pluginName)).End()
	if errs != nil {
		return nil, fmt.Errorf("could not get plugins, error: %v", errs)
	}
	ps := &PluginSchema{}
	err := json.Unmarshal([]byte(body), ps)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode != 200 {
		return nil, fmt.Errorf("Error occured during retrieving a plugin's schema data: %s", ps.ErrorMessage)
	}
	return ps, nil
}
