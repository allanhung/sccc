package cmd

import (
  "fmt"
  "os"
  "io"
  "io/ioutil"
  "bytes"
  "net/http"
  "strings"
  "path/filepath"
  "gopkg.in/yaml.v2"
  "github.com/spf13/cobra"
)

type configFiles []string

func (v *configFiles) String() string {
  return fmt.Sprint(*v)
}

func (v *configFiles) Type() string {
  return "configFiles"
}

func (v *configFiles) Set(value string) error {
  for _, filePath := range strings.Split(value, ",") {
    *v = append(*v, filePath)
  }
  return nil
}

type getCmdFlags struct {
  uri          string
  application  string
  namespace    string
  version      string
  branch       string
  config       configFiles
  resource     configFiles
}

func fetchFileFromUrl(url string) ([]byte, error) {
  fmt.Println("fetch file config from url:", url)
  resp, err := http.Get(url)
  if err != nil {
    return []byte{}, err
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return []byte{}, err
  }
  return body, nil
}

func (cfg *getCmdFlags) fetchFile(debug bool) error {
  baseurl := fmt.Sprintf("%s/%s/%s/%s", cfg.uri, cfg.application, cfg.namespace, cfg.branch)
  sections := []string{
    "default",
    "namespace."+cfg.namespace,
    "version."+cfg.version,
  }
  fmt.Printf("combind config with %v\n", sections)
  // process config files
  for _, configFile := range cfg.config {
    body := []byte{}
    configList := strings.Split(configFile, "=")
    if strings.Contains(configList[0], ":") {
      templateConf := strings.Split(configList[0], ":")
      defaultBody, err := fetchFileFromUrl(baseurl+"/"+templateConf[0])
      if err != nil {
        return err
      }
      // process default body
      if filepath.Ext(templateConf[0]) == ".properties" {
        if string(defaultBody[len(defaultBody)-1:]) == "\n" {
          defaultBody = defaultBody[:len(defaultBody)-1]
        }
        defaultBody = append([]byte("default:\n"), defaultBody...)
        defaultBody = bytes.ReplaceAll(defaultBody, []byte("\n"), []byte("\n  "))
        if templateConf[1] == "" {
          cBody, err := fetchFileFromUrl(baseurl+"/"+templateConf[1])
          if err != nil {
            return err
          }
          body = append(defaultBody, []byte("\n")...)
          body = append(body, cBody...)
        } else {
          body = defaultBody
        }
      } else {
        template := yaml.MapSlice{}
        if err := yaml.Unmarshal(defaultBody, &template); err != nil {
          return err
        }
        template = setValue(template, "default", template)
        body, err = yaml.Marshal(template)
        if err != nil {
          return err
        }
        if templateConf[1] == "" {
          cBody, err := fetchFileFromUrl(baseurl+"/"+templateConf[1])
          if err != nil {
            return err
          }
          body = append(body, cBody...)
        }
      }
    } else {
      var err error
      body, err = fetchFileFromUrl(baseurl+"/"+configList[0])
      if err != nil {
        return err
      }
    }
    if filepath.Ext(configList[0]) == ".properties" {
      body = bytes.ReplaceAll(body, []byte("="), []byte(": "))
    }
    data, err := valsections(body, sections)
    if err != nil {
      return err
    }
    if filepath.Ext(configList[0]) == ".properties" {
      data = bytes.ReplaceAll(data, []byte("null\n"), []byte("\n"))
      data = bytes.ReplaceAll(data, []byte(": "), []byte("="))
    }
    fmt.Println("save config to:", configList[1])
    out, err := os.Create(configList[1])
    if err != nil {
      return err
    }
    defer out.Close()
    _, err = out.Write(data)
    if err != nil {
      return err
    }
  }
  // process resource files
  for _, res := range cfg.resource {
    resList := strings.Split(res, "=")
    fmt.Println("fetch resource from url:", baseurl+"/"+resList[0])
    resp, err := http.Get(baseurl+"/"+resList[0])
    if err != nil {
      return err
    }
    defer resp.Body.Close()
    if debug {
      buf := new(bytes.Buffer)
      buf.ReadFrom(resp.Body)
      fmt.Println(buf.String())
    } else {
      fmt.Println("save resource to:", resList[1])
      out, err := os.Create(resList[1])
      if err != nil {
        return err
      }
      defer out.Close()

      _, err = io.Copy(out, resp.Body)
      if err != nil {
        return err
      }
    }
  }
  return nil
}

func getSection(template yaml.MapSlice, tag []string) (yaml.MapSlice, error) {
  if len(tag) > 1 {
    for _, item := range template {
      if item.Key == tag[0] {
        return getSection(item.Value.(yaml.MapSlice), tag[1: len(tag)])
      }
    }
  } else {
    for _, item := range template {
      if item.Key == tag[0] {
        return item.Value.(yaml.MapSlice), nil
      }
    }
  }
  return yaml.MapSlice{}, nil
}

func valsections(configContext []byte, sections []string) ([]byte, error) {
  template := yaml.MapSlice{}
  base := yaml.MapSlice{}

  if err := yaml.Unmarshal(configContext, &template); err != nil {
    return []byte{}, fmt.Errorf("failed to parse config: %s", err)
  }
  // section version.1.0.5
  for _, section := range sections {
    fmt.Println("process section:", section)
    currentMap, err := getSection(template, strings.SplitN(section,".",2))
    if err != nil {
      return []byte{}, err
    }
    // Merge with the previous map
    base = mergeValues(base, currentMap)
  }

  return yaml.Marshal(base)
}

func containsKey(slice yaml.MapSlice, key interface{}) bool {
  for _, item := range slice {
    if item.Key == key {
      return true
    }
  }
  return false
}

func setValue(slice yaml.MapSlice, key, newValue interface{}) yaml.MapSlice {
  for i := 0; i < len(slice); i++ {
    if slice[i].Key == key { // if key exist in slice, replace it
      slice[i].Value = newValue
      return slice
    }
  }
  // If we got to this point, it is a new key in slice, so just add at the end of slice
  return append(slice, yaml.MapItem{Key: key, Value: newValue})
}

func getValue(slice yaml.MapSlice, key interface{}) (value yaml.MapSlice, ok bool) {
	for _, item := range slice {
		if item.Key == key {
			value, ok = item.Value.(yaml.MapSlice)
			return
		}
	}
	return
}

func mergeValues(dest yaml.MapSlice, src yaml.MapSlice) yaml.MapSlice {
  for _, item := range src {
    // If the key doesn't exist already, then just set the key to that value
    if exists := containsKey(dest, item.Key); !exists {
      dest = setValue(dest, item.Key, item.Value)
      continue
    }
    nextMap, ok := item.Value.(yaml.MapSlice)
    // If it isn't another map, overwrite the value
    if !ok {
      dest = setValue(dest, item.Key, item.Value)
      continue
    }
    // Edge case: If the key exists in the destination, but isn't a map
    destMap, isMap := getValue(dest, item.Key)
    // If the source map has a map for this key, prefer it
    if !isMap {
      dest = setValue(dest, item.Key, item.Value)
      continue
    }
    // If we got to this point, it is a map in both, so merge them
    merged := mergeValues(destMap, nextMap)
    dest = setValue(dest, item.Key, merged)
  }
  return dest
}

var CmdFlags = getCmdFlags{}

// getCmd represents the get command
var getCmd = &cobra.Command{
  Use:   "get",
  Short: "get config from server",
  Long: `get config from spring cloud config server
For example:

sccc get -u http://localhost:8888 -a app -n dev -v 1.0.6 -b master \
         -c conf/app1.properties=/app/application1.properties \
         -c conf/app1.yaml=/app/application1.yaml \
         -r resources/myres1=/app/myres1.res \ 
         -r resources/myres2=/app/myres2.res
or
sccc get -u http://localhost:8888 -a app -n dev -v 1.0.6 -b master \
         -c conf/app2.properties:conf/app3.properties=/app/application1.properties \
         -c conf/app2.yaml:conf/app3.yaml=/app/application1.yaml \
         -r resources/myres1=/app/myres1.res \ 
         -r resources/myres2=/app/myres2.res`, 
    Run: func(cmd *cobra.Command, args []string) {
    err := CmdFlags.fetchFile(false)
    if err != nil {
      panic(err)
    }
  },
}

func init() {
  rootCmd.AddCommand(getCmd)
  f := getCmd.Flags()
  f.StringVarP(&CmdFlags.uri, "uri", "u", "http://localhost:8888", "spring cloud config server uri")
  f.StringVarP(&CmdFlags.application, "application", "a", "application", "application default: application")
  f.StringVarP(&CmdFlags.namespace, "namespace", "n", "", "kubernetes namespace")
  f.StringVarP(&CmdFlags.version, "version", "v", "", "application version")
  f.StringVarP(&CmdFlags.branch, "branch", "b", "master", "git branch default: master")
  f.VarP(&CmdFlags.config, "configfile", "c", "config file example: conf/app.conf=/etc/application.propertiess (can specify multiple)")
  f.VarP(&CmdFlags.resource, "resourcefile", "r", "resource file example: resources/myres=/app/app.res (can specify multiple)")
}
