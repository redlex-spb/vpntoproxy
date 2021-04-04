// application configuration processing package
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

const (
	descriptionTag = "desc"
	defaultTag     = "default"
	configsDir     = "./configs"
)

// global variable for program configuration
var conf *Config

// structure containing pointers to grouped parameters
type Config struct {
	Basic  *Basic
	Server *Server
	Docker *Docker
	Proxy  *Proxy
	Log    *Log
}

// structure of basic parameters
type Basic struct {
	Debug bool `json:"debug" default:"false" desc:"Debug mode"`
}

// structure of server parameters
type Server struct {
	Port int `json:"port" default:"8080"`
}

// structure of parameters for working with containers
type Docker struct {
	ImageName     string   `json:"image_name" default:"vpnwithproxy"`
	MaxAttempts   int      `json:"max_attempts" default:"10"`
	ServicePrefix string   `json:"service_prefix" default:"vpn_"`
	DNS           []string `json:"dns" default:"[\"8.8.8.8\", \"8.8.4.4\"]"`
	ProxyPort     int      `json:"proxy_port" default:"1080"`
	ProxyUser     string   `json:"proxy_user" default:"user"`
	ProxyPassword string   `json:"proxy_password" default:"password"`
}

// structure of parameters for proxying traffic through a container
type Proxy struct {
	StartingPort int    `json:"starting_port" default:"9001"`
	TestURL      string `json:"test_url" default:"http://httpbin.org/ip"`
}

// structure of log parameters
type Log struct {
	Mode       string `json:"mode" default:"file"`
	MaxSize    int    `json:"max_size" default:"10"` // megabytes
	MaxBackups int    `json:"max_backups" default:"10"`
	MaxAge     int    `json:"max_age" default:"28"` // days
	Compress   bool   `json:"compress" default:"true"`
}

// Creation of a configuration object.
// The configuration structure is iterated over, filling nested structures with data.
// Value setting priority:
// - data from the file (if file not exist, a configuration file is created, filling it with default parameters from the structure)
// - data from environment variables
// - transmitted data from flags
func make() *Config {
	// instantiating configuration
	conf = &Config{}

	// filling the structure
	err := iterateTemplate(conf, false)

	checkError(err)

	// parsing set flags
	flag.Parse()

	if conf.Basic.Debug {
		printConfToLog(conf)
	}

	return conf
}

// global method for getting an instance of application configuration
func Get() *Config {
	if conf != nil {
		return conf
	} else {
		return make()
	}
}

// panic output on error
func checkError(err error) {
	if err != nil {
		logrus.Panicln(err)
	}
}

// template processing
func iterateTemplate(template interface{}, setDefault bool) (err error) {
	// checking if a template is a pointer
	if reflect.ValueOf(template).Kind() == reflect.Ptr {

		// receiving reflex value template instance
		reflectTemplate := reflect.ValueOf(template).Elem()
		// getting type/information about an object
		typeOfT := reflectTemplate.Type()

		// iterating template fields
		for i := 0; i < reflectTemplate.NumField(); i++ {
			// receiving reflex value template instance
			field := reflectTemplate.Field(i)
			// getting type/information about an field
			fieldType := typeOfT.Field(i)
			// getting default value from tag
			defaultValue := fieldType.Tag.Get(defaultTag)
			// generate key from name type and name field to a lower case key, for setting the flag
			keyLow := upperToUnderline(fmt.Sprintf("%s%s", typeOfT.Name(), fieldType.Name))
			// generating an uppercase key to check the value in environment variables
			keyUp := strings.ToUpper(keyLow)
			// formation of a lower case key, for setting the flag
			//keyLow := strings.ToLower(keyUp)
			if field.IsValid() {
				switch field.Kind() {
				case reflect.Ptr:
					// creating an instance of the reflect value of the field type
					fieldValue := reflect.New(fieldType.Type.Elem())

					// checking if the interface is obtainable and that the field type is a pointer
					if fieldValue.CanInterface() && fieldValue.Type().Kind() == reflect.Ptr {
						// getting a filled structure with data from a file
						fileV, err := readFile(strings.ToLower(fieldType.Name), fieldValue.Interface())
						if err != nil {
							return err
						}

						// setting the field to the resulting filled structure
						field.Set(reflect.ValueOf(fileV))

						// recursively process the field
						err = iterateTemplate(field.Interface(), setDefault)
						if err != nil {
							return err
						}
					}

				case reflect.String:
					if setDefault {
						// setting the field value to the default
						field.SetString(defaultValue)
					} else {
						if field.CanAddr() && field.Addr().CanInterface() {
							// creating a pointer-to-field interface
							valueInterface := field.Addr().Interface()
							// casting an interface to a value pointer
							ptrValStr := valueInterface.(*string)
							// flag declaration passing a pointer to the value to set, also the pointer value as default
							flag.StringVar(ptrValStr, keyLow, *ptrValStr, fieldType.Tag.Get(descriptionTag))
						}

						// checking value in environment variables
						if v := os.Getenv(keyUp); v != "" {
							// setting the field value from environment variables
							field.SetString(v)
						}
					}
				case reflect.Int:
					if setDefault {
						// setting the field value to the default
						if vInt, err := strconv.Atoi(defaultValue); err != nil {
							logrus.Error(err)
						} else {
							field.SetInt(int64(vInt))
						}
					} else {
						if field.CanAddr() && field.Addr().CanInterface() {
							// creating a pointer-to-field interface
							valueInterface := field.Addr().Interface()
							// casting an interface to a value pointer
							ptrValInt := valueInterface.(*int)
							// flag declaration passing a pointer to the value to set, also the pointer value as default
							flag.IntVar(ptrValInt, keyLow, *ptrValInt, fieldType.Tag.Get(descriptionTag))
						}

						// checking value in environment variables
						if v := os.Getenv(keyUp); v != "" {
							// setting the field value from environment variables
							if vInt, err := strconv.Atoi(v); err != nil {
								logrus.Error(err)
							} else {
								field.SetInt(int64(vInt))
							}
						}
					}
				case reflect.Bool:
					if setDefault {
						// setting the field value to the default
						field.SetBool(defaultValue == "true")
					} else {
						if field.CanAddr() && field.Addr().CanInterface() {
							// creating a pointer-to-field interface
							valueInterface := field.Addr().Interface()
							// casting an interface to a value pointer
							ptrValBool := valueInterface.(*bool)
							// flag declaration passing a pointer to the value to set, also the pointer value as default
							flag.BoolVar(ptrValBool, keyLow, *ptrValBool, fieldType.Tag.Get(descriptionTag))
						}

						// checking value in environment variables
						if v := os.Getenv(keyUp); v != "" {
							// setting the field value from environment variables
							field.SetBool(v == "true")
						}
					}
				case reflect.Slice, reflect.Map, reflect.Array:
					fieldValue := reflect.New(fieldType.Type)
					if setDefault {
						// setting the field value to the default
						if fieldValue.CanInterface() {
							// filling the field value with default value
							if err := json.Unmarshal([]byte(defaultValue), fieldValue.Interface()); err != nil {
								return err
							}
							field.Set(fieldValue.Elem())
						}
					} else {
						// checking value in environment variables
						if v := os.Getenv(keyUp); v != "" {
							if fieldValue.CanInterface() {
								// filling the field value from environment variables
								if err := json.Unmarshal([]byte(v), fieldValue.Interface()); err != nil {
									return err
								}
								// setting the field value from environment variables
								field.Set(fieldValue.Elem())
							}
						}
					}
				}
			}
		}

	}

	return err
}

// read params from file
func readFile(name string, template interface{}) (interface{}, error) {
	// Open json file
	jsonFile, err := os.Open(fmt.Sprintf("%s/%s.json", configsDir, name))
	// if os.Open returns an error then handle it
	if err != nil {
		logrus.Println(err)

		// file is not exist, create file
		if _, ok := err.(*os.PathError); ok {
			return createFile(name, template)
		}

		return nil, err
	}
	//logrus.Printf("Successfully Opened %s.json\n", name)
	// defer the closing of json file
	defer func() {
		err := jsonFile.Close()
		if err != nil {
			logrus.Println(err)
		}
	}()

	// read opened jsonFile as a byte array.
	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		logrus.Println(err)
		return nil, err
	}

	// filling the template with data from a file
	err = json.Unmarshal(bytes, &template)
	if err != nil {
		logrus.Println(err)
		return nil, err
	}

	return template, nil
}

// create file with parameters structure
func createFile(name string, template interface{}) (interface{}, error) {
	// filling the template with a default value
	if err := iterateTemplate(template, true); err != nil {
		return nil, err
	}

	// casting a template to an array of bytes
	jsonBytes, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return nil, err
	}

	// write template to json file
	if err = ioutil.WriteFile(fmt.Sprintf("%s/%s.json", configsDir, name), jsonBytes, 0644); err != nil {
		return nil, err
	}

	return template, err
}

func upperToUnderline(sentence string) string {
	var (
		resRune       []rune
		underlineRune int32 = 95
		prevUpper     bool
	)

	for i, r := range sentence {
		if unicode.IsUpper(r) && unicode.IsLetter(r) {
			if i != 0 && !prevUpper {
				resRune = append(resRune, underlineRune)
			}
			r = unicode.ToLower(r)
			prevUpper = true
		} else {
			prevUpper = false
		}
		resRune = append(resRune, r)
	}

	return string(resRune)
}

func printConfToLog(config *Config) {
	// checking if a template is a pointer
	if reflect.ValueOf(config).Kind() == reflect.Ptr {

		// receiving reflex value template instance
		reflectTemplate := reflect.ValueOf(config).Elem()
		// getting type/information about an object
		typeOfT := reflectTemplate.Type()

		// iterating template fields
		for i := 0; i < reflectTemplate.NumField(); i++ {
			// receiving reflex value template instance
			field := reflectTemplate.Field(i)
			// getting type/information about an field
			fieldType := typeOfT.Field(i)
			if field.CanInterface() {
				// Print to log
				logrus.Printf("%s.%s: %+v", typeOfT.Name(), fieldType.Name, field.Interface())
			}
		}
	}
}
