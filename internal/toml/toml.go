package toml

import (
	"os"
	"reflect"
	"strings"

	"drift/internal/errs"

	tm "github.com/BurntSushi/toml"
)

func Init(path string, str any, defaultConfig map[string]any, overwrite bool) error {

	t := reflect.TypeOf(str).Elem()
	v := reflect.ValueOf(str).Elem()

	_, err := os.Stat(path)

	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if os.IsNotExist(err) || overwrite {

		fillStructFromMap(t, v, "", defaultConfig)

		err = Write(path, str)
		if err != nil {
			return errs.NewError(errs.ErrWritingToml, err.Error())
		}

		return nil
	}

	tomlMeta, err := tm.DecodeFile(path, str)
	if err != nil {
		return errs.NewError(errs.ErrDecodingToml, err.Error())
	}

	var recreateToml bool

	checkToml(tomlMeta, t, v, "", defaultConfig, &recreateToml)

	if recreateToml {
		err := Write(path, str)
		if err != nil {
			return errs.NewError(errs.ErrFixingToml, err.Error())
		}
	}

	return nil
}

func Write(path string, str any) error {

	tmpPath := path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	defer os.Remove(tmpPath)

	err = tm.NewEncoder(f).Encode(str)
	if err != nil {
		f.Close()
		return err
	}

	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	err = os.Rename(tmpPath, path)
	if err != nil {
		return err
	}

	return nil
}

func checkToml(tomlMeta tm.MetaData, t reflect.Type, v reflect.Value, prefix string, defaultConfig map[string]any, recreateToml *bool) {

	for i := 0; i < t.NumField(); i++ {

		tt := t.Field(i)
		tv := v.Field(i)

		name := tt.Tag.Get("toml")
		if name == "" {
			continue
		}

		key := ""
		if prefix != "" {
			key = strings.Join([]string{prefix, name}, ".")
		} else {
			key = name
		}

		if tv.Kind() == reflect.Struct {
			checkToml(tomlMeta, tt.Type, tv, key, defaultConfig, recreateToml)
			continue
		}

		if !tomlMeta.IsDefined(key) && tv.CanSet() {

			switch tv.Kind() {

			case reflect.String:
				if v, ok := defaultConfig[key].(string); ok {
					*recreateToml = true
					tv.SetString(v)
				}

			case reflect.Int:
				if v, ok := defaultConfig[key].(int); ok {
					*recreateToml = true
					tv.SetInt(int64(v))
				}

			case reflect.Bool:
				if v, ok := defaultConfig[key].(bool); ok {
					*recreateToml = true
					tv.SetBool(v)
				}
			}
		}

	}

}

func fillStructFromMap(t reflect.Type, v reflect.Value, prefix string, values map[string]any) {

	for i := 0; i < t.NumField(); i++ {

		tt := t.Field(i)
		tv := v.Field(i)

		name := tt.Tag.Get("toml")
		if name == "" {
			continue
		}

		key := ""
		if prefix != "" {
			key = strings.Join([]string{prefix, name}, ".")
		} else {
			key = name
		}

		if tv.Kind() == reflect.Struct {
			fillStructFromMap(tt.Type, tv, key, values)
			continue
		}

		if tv.CanSet() {

			switch tv.Kind() {

			case reflect.String:
				if v, ok := values[key].(string); ok {
					tv.SetString(v)
				}

			case reflect.Int:
				if v, ok := values[key].(int); ok {
					tv.SetInt(int64(v))
				}

			case reflect.Bool:
				if v, ok := values[key].(bool); ok {
					tv.SetBool(v)
				}
			}
		}

	}
}
