package tpl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/spf13/viper"
)

func getValueStdin() string {
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return ""
	}
	text := scanner.Text()
	return strings.TrimSpace(text)
}

func countLeadingSpace(line string) int {
	i := 0
	for _, runeValue := range line {
		if runeValue == ' ' {
			i++
		} else if runeValue == '\t' {
			i++
		} else {
			break
		}
	}
	return i
}

func nestedToFlattenMap(value interface{}, list map[string]interface{}, path string, delegate bool) string {
	switch value.(type) {
	//case reflect.Interface:
	//	fmt.Println("value is interface", reflect.TypeOf(value), value)
	case map[interface{}]interface{}:
		path = path + "."
		xxx := value.(map[interface{}]interface{})
		for key, val := range xxx {
			tpath := path + key.(string)
			nestedToFlattenMap(val, list, tpath, false)
		}
	case map[string]interface{}:
		path = path + "."
		xxx := value.(map[string]interface{})
		for key, val := range xxx {
			tpath := path + key
			nestedToFlattenMap(val, list, tpath, false)
		}
	case []interface{}:
		path = path + "."
		xxx := value.([]interface{})
		for idx, val := range xxx {
			tpath := path + "[" + strconv.Itoa(idx) + "]"
			nestedToFlattenMap(val, list, tpath, true)
		}
	default:
		list[path] = value
	}
	return path
}

/*
func flattenToNestedMap(value interface{}, dataMap map[string]interface{}, path string, delegate bool) string {
	switch value.(type) {
	case map[interface{}]interface{}:
		path = strings.TrimPrefix(path, ".")
		xxx := value.(map[interface{}]interface{})
		for key, val := range xxx {
			tpath := path + key.(string)
			nestedToFlattenMap(val, dataMap, tpath, false)
		}
	case map[string]interface{}:
		path = path + "."
		xxx := value.(map[string]interface{})
		for key, val := range xxx {
			tpath := path + key
			nestedToFlattenMap(val, dataMap, tpath, false)
		}
	case []interface{}:
		path = path + "."
		xxx := value.([]interface{})
		for idx, val := range xxx {
			tpath := path + "[" + strconv.Itoa(idx) + "]"
			nestedToFlattenMap(val, dataMap, tpath, true)
		}
	default:
		dataMap[path] = value
	}
	return path
}
*/

func expand(value map[string]interface{}) map[string]interface{} {
	return expandPrefixed(value, "")
}

func expandPrefixed(value map[string]interface{}, prefix string) map[string]interface{} {
	m := make(map[string]interface{})
	expandPrefixedToResult(value, prefix, m)
	return m
}

func expandPrefixedToResult(value map[string]interface{}, prefix string, result map[string]interface{}) {
	//if prefix != "" {
	prefix += "."
	//}
	for k, val := range value {
		if !strings.HasPrefix(k, prefix) {
			continue
		}

		key := k[len(prefix):]
		idx := strings.Index(key, ".")
		if idx != -1 {
			key = key[:idx]
		}
		if _, ok := result[key]; ok {
			continue
		}
		if idx == -1 {
			result[key] = val
			continue
		}

		// It contains a period, so it is a more complex structure
		result[key] = expandPrefixed(value, k[:len(prefix)+len(key)])
	}
}

// CreateDirectoryIfNotExists creates a directory only if it doesn't already exist.
func CreateDirectoryIfNotExists(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			//MkdirAll
			return os.Mkdir(path, os.ModePerm)
		}
	}
	return nil
}

// ErrFileExists contains error message
type ErrFileExists struct {
	message string
}

// NewErrFileExists creates the ErrFileExists
func NewErrFileExists(message string) *ErrFileExists {
	return &ErrFileExists{
		message: message,
	}
}

// Error returns error message
func (e *ErrFileExists) Error() string {
	return e.message
}

// WriteStringToFileAndCreateDir writes string to the file at path `dst`, creating it if necessary.
func WriteStringToFileAndCreateDir(dst string, content string, overwrite bool) error {
	if !overwrite {
		if _, err := os.Stat(dst); !os.IsNotExist(err) {
			fmt.Printf("tpl: overwrite '%s'? ", dst)
			val := strings.ToLower(getValueStdin())
			if !strings.HasPrefix(val, "y") {
				return NewErrFileExists("file exists")
			}
		}
	}
	path := filepath.Dir(dst)
	if path != "." {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, strings.NewReader(content))
	if err != nil {
		return err
	}
	// err = f.Chmod(mode)
	// if err != nil {
	// 	return err
	// }
	return nil
}

// NavColorMeta holds metadata for git style color
type NavColorMeta struct {
	ExecInfo   *color.Color
	NavFile    *color.Color
	NavTitle   *color.Color
	NavContext *color.Color
	NavInput   *color.Color
}

var navColorMeta *NavColorMeta
var once sync.Once

// InitializedNavColorMeta return NavColorMeta that it has been initialized once
func InitializedNavColorMeta() *NavColorMeta {
	once.Do(func() {
		navColorMeta = &NavColorMeta{}
		navColorMeta.init()
	})
	return navColorMeta
}

func (c *NavColorMeta) init() error {
	execInfo := viper.GetString("color.exec.file")
	navFile := viper.GetString("color.nav.file")
	navTitle := viper.GetString("color.nav.title")
	navContext := viper.GetString("color.nav.context")
	navInput := viper.GetString("color.nav.input")

	c.ExecInfo = newColor(execInfo, "")
	c.NavFile = newColor(navFile, "cyan")
	c.NavTitle = newColor(navTitle, "yellow")
	c.NavContext = newColor(navContext, "")
	c.NavInput = newColor(navInput, "normal normal bold")

	nav := viper.GetString("color.nav.ui")
	if nav == "true" || nav == "auto" {
		c.ExecInfo.EnableColor()
		c.NavFile.EnableColor()
		c.NavTitle.EnableColor()
		c.NavContext.EnableColor()
		c.NavInput.EnableColor()
	} else if nav == "false" || nav == "never" {
		c.ExecInfo.DisableColor()
		c.NavFile.DisableColor()
		c.NavTitle.DisableColor()
		c.NavContext.DisableColor()
		c.NavInput.DisableColor()
	}

	return nil
}

func newColor(colorsStr string, defaultColorsStr string) *color.Color {
	c := color.New()
	if colorsStr == "" {
		if defaultColorsStr == "" {
			return c
		}
		colorsStr = defaultColorsStr
	}
	colors := strings.Fields(colorsStr)
	if len(colors) >= 1 {
		fg := strings.ToLower(colors[0])
		if fg == "normal" {
			fg = defaultColorsStr
		}
		setFgColor(c, fg)
	}
	if len(colors) >= 2 {
		bg := strings.ToLower(colors[1])
		if bg == "normal" {
			bg = ""
		}
		setBgColor(c, bg)
	}
	if len(colors) >= 3 {
		bold := strings.ToLower(colors[2])
		if bold == "normal" {
			bold = ""
		}
		setBoldColor(c, bold)
	}
	return c
}

func setFgColor(c *color.Color, colorStr string) {
	switch strings.ToLower(colorStr) {
	case "red":
		c.Add(color.FgRed)
	case "green":
		c.Add(color.FgGreen)
	case "yellow":
		c.Add(color.FgYellow)
	case "blue":
		c.Add(color.FgBlue)
	case "magenta":
		c.Add(color.FgMagenta)
	case "cyan":
		c.Add(color.FgCyan)
	case "white":
		c.Add(color.FgWhite)
	case "black":
		c.Add(color.FgBlack)
	}
}

func setBgColor(c *color.Color, colorStr string) {
	switch strings.ToLower(colorStr) {
	case "red":
		c.Add(color.BgRed)
	case "green":
		c.Add(color.BgGreen)
	case "yellow":
		c.Add(color.BgYellow)
	case "blue":
		c.Add(color.BgBlue)
	case "magenta":
		c.Add(color.BgMagenta)
	case "cyan":
		c.Add(color.BgCyan)
	case "white":
		c.Add(color.BgWhite)
	case "black":
		c.Add(color.BgBlack)
	}
}

func setBoldColor(c *color.Color, colorStr string) {
	switch strings.ToLower(colorStr) {
	case "bold":
		c.Add(color.Bold)
	}
}
