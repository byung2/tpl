package tpl

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/imdario/mergo"
	"github.com/spf13/viper"
	ini "github.com/vaughan0/go-ini"
	"gopkg.in/yaml.v2"
)

// TmplOpts holds options to execute template
type TmplOpts struct {
	DataFilesStr       string
	DataFiles          []string
	DataFormat         string
	TmplFiles          []string
	Output             string
	OutDir             string
	OutExt             string
	MissingKey         string
	Interactive        bool
	UseEnv             bool
	UseEnvFromPrefix   string
	DataOutFile        string
	DataOutFormat      string
	FoldContext        bool
	ShowProcessedFile  bool
	ShowOnlyMissingKey bool
	Overwrite          bool
}

// Tmpl contains metadata
type Tmpl struct {
	Re       *regexp.Regexp
	TmplOpts *TmplOpts
	Data     map[string]interface{}
	Files    []*TmplFileMeta
}

// TmplFileMeta holds information about template file
type TmplFileMeta struct {
	Name     string
	OrigPath string
	DestPath string
	Mode     os.FileMode
	Content  string
	Changed  bool
}

// LineMeta holds metadata specific to the line
type LineMeta struct {
	Line         string
	ReplacedLine string
	Space        int
	NoValue      bool
}

const (
	missingKeyError   = "missingkey=error"
	missingKeyDefault = "missingkey=default"
)

func getFileExt(file string) string {
	return strings.TrimPrefix(filepath.Ext(file), ".")
}

// OptsToTmpl creates a Tmpl object from TmplOpts
func (opts *TmplOpts) OptsToTmpl() (Tmpl, error) {
	tmpl := Tmpl{TmplOpts: opts}

	// Check options
	missingKey := strings.ToLower(opts.MissingKey)
	if missingKey == "" {
		missingKey = "error"
	}
	if missingKey != "error" && missingKey != "zero" && missingKey != "default" && missingKey != "invalid" {
		return tmpl, fmt.Errorf("wrong missing key option")
	}
	opts.MissingKey = missingKey

	useEnv := viper.Get("tpl.env")
	foldContext := viper.Get("tpl.fold-context")
	interactive := viper.Get("tpl.interactive")
	overwrite := viper.Get("tpl.overwrite")
	showProcessedFile := viper.Get("tpl.show-processed-info")
	if useEnv != nil && !opts.UseEnv {
		opts.UseEnv = viper.GetBool("tpl.env")
	}
	if foldContext != nil && !opts.FoldContext {
		opts.FoldContext = viper.GetBool("tpl.fold-context")
	}
	if interactive != nil && !opts.Interactive {
		opts.Interactive = viper.GetBool("tpl.interactive")
	}
	if overwrite != nil && !opts.Overwrite {
		opts.Overwrite = viper.GetBool("tpl.overwrite")
	}
	if showProcessedFile != nil && !opts.ShowProcessedFile {
		opts.ShowProcessedFile = viper.GetBool("tpl.show-processed-info")
	}

	// data files separator: space vs colon
	//opts.DataFiles = strings.Fields(opts.DataFilesStr)
	dataFilesElem := make(map[string]bool)
	dataFiles := []string{}
	if opts.DataFilesStr != "" {
		dataFiles = strings.Split(opts.DataFilesStr, ":")
	}
	for _, v := range dataFiles {
		matches, err := filepath.Glob(v)
		if err != nil {
			return tmpl, fmt.Errorf("datafile glob error: %v", err)
		}
		for _, match := range matches {
			_, ok := dataFilesElem[match]
			if !ok {
				dataFilesElem[match] = true
			}
		}
	}
	for k := range dataFilesElem {
		opts.DataFiles = append(opts.DataFiles, k)
	}
	defaultDataFormat := "yaml"
	switch opts.DataFormat {
	case "json", "yml", "yaml", "ini", "kv":
		defaultDataFormat = opts.DataFormat
	}
	// DataFiles to Data object
	datakv := make(map[string]interface{})
	for _, file := range opts.DataFiles {
		dat, err := ioutil.ReadFile(file)
		if err != nil {
			return tmpl, err
		}
		//var kv map[string]interface{}
		kv := make(map[string]interface{})
		ext := getFileExt(file)
		switch ext {
		case "json", "yml", "yaml", "ini", "kv":
		default:
			ext = defaultDataFormat
		}
		switch ext {
		case "yml", "yaml":
			err = yaml.Unmarshal(dat, &kv)
			if err != nil {
				return tmpl, fmt.Errorf("failed to parse yaml data file '%s': %v", file, err)
			}
		case "json":
			err = json.Unmarshal(dat, &kv)
			if err != nil {
				return tmpl, fmt.Errorf("failed to parse json data file '%s': %v", file, err)
			}
		case "ini":
			inifile, err := ini.Load(bytes.NewReader(dat))
			if err != nil {
				return tmpl, fmt.Errorf("failed to parse ini data file '%s': %v", file, err)
			}
			for name, section := range inifile {
				kv[name] = section
			}
		case "kv":
			// parse key=value format
			tmpKv := make(map[string]interface{})
			scanner := bufio.NewScanner(strings.NewReader(string(dat)))
			for scanner.Scan() {
				line := scanner.Text()
				elems := strings.Split(line, "=")
				if len(elems) <= 1 {
					fmt.Printf("warn: datafile is neither yaml nor key=value format\n")
					continue
				}
				tmpKv[appendKeyPrefix(elems[0])] = strings.TrimSpace(elems[1])
			}
			kv = expand(tmpKv)
		}
		mergo.Merge(&datakv, kv)
	}
	tmpl.Data = datakv
	tmpl.TmplOpts = opts
	tmpl.Re = regexp.MustCompile(`{{[\s]*(\..*?)[\s]*}}`)
	if opts.UseEnv || opts.UseEnvFromPrefix != "" {
		tmpDataOutFormat := tmpl.TmplOpts.DataOutFormat
		tmpl.TmplOpts.DataOutFormat = "kv"
		keysStr, err := tmpl.ExtractKeys()
		tmpl.TmplOpts.DataOutFormat = tmpDataOutFormat
		if err != nil {
			return tmpl, err
		}
		keyScanner := bufio.NewScanner(strings.NewReader(keysStr))
		keys := make(map[string]bool)
		for keyScanner.Scan() {
			keys[strings.SplitN(keyScanner.Text(), "=", 2)[0]] = true
			//keys[appendKeyPrefix(strings.SplitN(keyScanner.Text(), "=", 2)[0])] = true
		}

		envPrefix := opts.UseEnvFromPrefix
		if opts.UseEnvFromPrefix != "" {
			envPrefix = opts.UseEnvFromPrefix + "."
		}
		envs := os.Environ()
		if envPrefix == "" {
			for _, e := range envs {
				elems := strings.Split(e, "=")
				if len(elems) >= 1 {
					_, ok := datakv[elems[0]]
					if !ok {
						_, ok = keys[elems[0]]
						if ok {
							datakv[elems[0]] = elems[1]
						}
					}
				}
			}
		} else {
			for _, e := range envs {
				elems := strings.Split(e, "=")
				if len(elems) >= 1 {
					_, ok := keys[envPrefix+elems[0]]
					if ok {
						if datakv[opts.UseEnvFromPrefix] == nil {
							datakv[opts.UseEnvFromPrefix] = make(map[interface{}]interface{})
						}
						i, ok := datakv[opts.UseEnvFromPrefix].(map[interface{}]interface{})
						if ok {
							_, ok = i[elems[0]]
							if !ok {
								i[elems[0]] = elems[1]
							}
						} else {
							i2, ok := datakv[opts.UseEnvFromPrefix].(map[string]interface{})
							if ok {
								_, ok = i2[elems[0]]
								if !ok {
									i2[elems[0]] = elems[1]
								}
							} else {
								return tmpl, fmt.Errorf("env prefix is already used for data key")
							}

						}
					}
				}
			}
		}
	}
	return tmpl, nil
}

func appendKeyPrefix(text string) string {
	if strings.HasPrefix(text, ".") {
		return text
	}
	return "." + text
}

func trimKeyPrefix(text string) string {
	return strings.TrimPrefix(text, ".")
}

// Ensure no missing keys in the template file
func (tmpl *Tmpl) Ensure(file string) error {
	data := tmpl.Data
	t, err := template.ParseFiles(file)
	if err != nil {
		return fmt.Errorf("error parsing template(s): %v", err)
	}

	// check missing keys
	t.Option(missingKeyError)
	buf := new(bytes.Buffer)
	err = t.Execute(buf, data)
	if err != nil {
		if !strings.Contains(err.Error(), "map has no entry for key") {
			return fmt.Errorf("failed to execute template: %v", err)
		}
		return fmt.Errorf("missing key found: %v", err)
	}
	return nil
}

// Keys store all missing keys to dataFlattenMap
func (tmpl *Tmpl) Keys(file string, dataFlattenMap map[string]interface{}) error {
	re := tmpl.Re
	// file to LineMeta
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer f.Close()
	var lines []string
	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSuffix(line, "\n")
		lines = append(lines, line)
	}
	for _, line := range lines {
		subMatched := re.FindAllStringSubmatch(line, -1)
		for _, x := range subMatched {
			_, ok := dataFlattenMap[x[1]]
			if !ok {
				dataFlattenMap[x[1]] = ""
			}
		}
	}
	return nil
}

// Execute gotemplate
func (tmpl *Tmpl) Execute(file string, dataFlattenMap map[string]interface{}) error {
	re := tmpl.Re
	data := tmpl.Data
	opts := tmpl.TmplOpts
	interactive := opts.Interactive
	//outDir := opts.OutDir
	t, err := template.ParseFiles(file)
	if err != nil {
		return fmt.Errorf("error parsing template(s): %v", err)
	}

	fileinfo, err := os.Stat(file)
	tfm := &TmplFileMeta{
		Name:     path.Base(file),
		OrigPath: file,
		Mode:     fileinfo.Mode(),
		//Changed: ooo,
	}
	t.Option(fmt.Sprintf("missingkey=%s", opts.MissingKey))
	buf := new(bytes.Buffer)
	err = t.Execute(buf, data)
	if err != nil {
		if !strings.Contains(err.Error(), "map has no entry for key") {
			return fmt.Errorf("failed to execute template: %v", err)
		}
		if !interactive {
			return fmt.Errorf("interactive flag is disabled, but the missing key is found: %v", err)
		}
	} else {
		tfm.Content = buf.String()
		tmpl.Files = append(tmpl.Files, tfm)
		//continue
		return nil
	}

	// file to LineMeta
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer f.Close()
	lineMetaList := []*LineMeta{}
	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSuffix(line, "\n")
		space := countLeadingSpace(line)
		lineMeta := LineMeta{Line: line, Space: space}
		lineMetaList = append(lineMetaList, &lineMeta)
	}

	// execute template with different missingkey option
	t.Option(missingKeyDefault)
	renderedOutputBuf := new(bytes.Buffer)
	err = t.Execute(renderedOutputBuf, data)
	if err != nil {
		return fmt.Errorf("failed to execute template with default option '%s': %v", missingKeyDefault, err)
	}
	reader = bufio.NewReader(strings.NewReader(renderedOutputBuf.String()))
	i := 0
	for {
		line, err := reader.ReadString('\n')
		// TODO EOF
		if err != nil {
			break
		}
		line = strings.TrimSuffix(line, "\n")
		if i >= len(lineMetaList) {
			fmt.Println("failed to interaction: i > len(lineMetaList)", i, len(lineMetaList))
		}
		if strings.Contains(line, "<no value>") {
			lineMetaList[i].ReplacedLine = line
			lineMetaList[i].NoValue = true
		}
		i = i + 1
	}

	c := InitializedNavColorMeta()
	c.NavFile.Printf("[%s]\n", file)
	for i, lineMeta := range lineMetaList {
		if lineMeta.NoValue {
			prevSpace := lineMetaList[i].Space
			tmpLineMetaList := []LineMeta{}
			for j := i - 1; j >= 0; j-- {
				space := lineMetaList[j].Space
				if lineMetaList[j].Line != "" && lineMetaList[j].Line != "\n" && space < prevSpace {
					tmpLineMeta := LineMeta{Line: lineMetaList[j].Line}
					tmpLineMetaList = append(tmpLineMetaList, tmpLineMeta)
					prevSpace = space
					if space == 0 {
						break
					}
				}
			}
			ttListLen := len(tmpLineMetaList)

			line := lineMetaList[i].Line
			subMatched := re.FindAllStringSubmatch(line, -1)

			// do not show context if a key already has value
			needToInput := false
			for _, x := range subMatched {
				_, ok := dataFlattenMap[x[1]]
				if !ok {
					needToInput = true
					break
				} else {
				}
			}
			if !needToInput {
				continue
			}

			if !opts.FoldContext {
				//fmt.Printf("\n")
				c.NavTitle.Printf("missing key found\n")
				for k := ttListLen - 1; k >= 0; k-- {
					tmpLineMeta := tmpLineMetaList[k]
					c.NavContext.Printf("%s\n", tmpLineMeta.Line)
				}
				c.NavContext.Printf("%s\n", lineMetaList[i].Line)
			}
			for _, x := range subMatched {
				_, ok := dataFlattenMap[x[1]]
				if !ok {
					c.NavInput.Printf("value for '%s': ", x[1])
					val := getValueStdin()
					dataFlattenMap[x[1]] = val
				}
			}
			if !opts.FoldContext {
				fmt.Printf("\n")
			}
		}
	}
	newNestedDataMap := expand(dataFlattenMap)
	t.Option(fmt.Sprintf("missingkey=%s", opts.MissingKey))

	// execute template with new data
	buf = new(bytes.Buffer)
	err = t.Execute(buf, newNestedDataMap)
	if err != nil {
		//TODO
		return fmt.Errorf("failed to execute template: %v", err)
	}
	tfm.Content = buf.String()
	tmpl.Files = append(tmpl.Files, tfm)
	return nil
}

// ExtractKeys get all missing keys and processed key:value pairs with given format (by default, yaml)
func (tmpl *Tmpl) ExtractKeys() (string, error) {
	tmpMissingKeyOption := tmpl.TmplOpts.MissingKey
	str, err := tmpl.extractKeys()
	tmpl.TmplOpts.MissingKey = tmpMissingKeyOption
	return str, err
}

func (tmpl Tmpl) extractKeys() (string, error) {
	tmpl.TmplOpts.MissingKey = "default"
	tmplFiles := tmpl.TmplOpts.TmplFiles
	dataFlattenMap := make(map[string]interface{})
	givenDataFlattenMap := make(map[string]interface{})
	data := tmpl.Data
	//fmt.Println("DEBUG: tmpl.Data:", data)
	nestedToFlattenMap(data, givenDataFlattenMap, "", false)
	for _, file := range tmplFiles {
		err := tmpl.Execute(file, givenDataFlattenMap)
		if err != nil {
			return "", err
		}
	}
	for _, file := range tmplFiles {
		err := tmpl.Keys(file, dataFlattenMap)
		if err != nil {
			return "", err
		}
	}
	// fill values with the given data
	for key := range dataFlattenMap {
		value, ok := givenDataFlattenMap[key]
		if ok {
			dataFlattenMap[key] = value
		}
	}

	if tmpl.TmplOpts.ShowOnlyMissingKey {
		for key := range givenDataFlattenMap {
			_, ok := dataFlattenMap[key]
			if ok {
				delete(dataFlattenMap, key)
			}
		}
	}
	dataOut, err := tmpl.marshalData(dataFlattenMap)
	return dataOut, err
}

func (tmpl Tmpl) marshalData(dataFlattenMap map[string]interface{}) (string, error) {
	expandedKeys := expand(dataFlattenMap)
	dataOut := ""
	if len(dataFlattenMap) > 0 {
		if tmpl.TmplOpts.DataOutFormat == "" {
			ext := getFileExt(tmpl.TmplOpts.DataOutFile)
			switch ext {
			case "json", "yml", "yaml", "ini", "kv":
				tmpl.TmplOpts.DataOutFormat = ext
			}
		}

		switch tmpl.TmplOpts.DataOutFormat {
		case "kv":
			for key, value := range dataFlattenMap {
				dataOut = fmt.Sprintf("%s%s=%v\n", dataOut, trimKeyPrefix(key), value)
			}
		case "json":
			dd, err := json.MarshalIndent(&expandedKeys, "", "  ")
			if err != nil {
				return "", fmt.Errorf("failed to marshal keys map to json: %v", err)
			}
			dataOut = fmt.Sprintf("%s\n", string(dd))
		default:
			dd, err := yaml.Marshal(&expandedKeys)
			if err != nil {
				return "", fmt.Errorf("failed to marshal keys map to yaml: %v", err)
			}
			dataOut = fmt.Sprintf("---\n\n%s", string(dd))
			//dataOut = fmt.Sprintf("%s", string(dd))
		}
	}
	return dataOut, nil
}

// WriteDataObject writes the filled data to the output
func (tmpl *Tmpl) WriteDataObject(dataOut string) {
	if tmpl.TmplOpts.DataOutFile != "" {
		WriteStringToFileAndCreateDir(tmpl.TmplOpts.DataOutFile, dataOut, tmpl.TmplOpts.Overwrite)
	} else {
		fmt.Printf("%s", dataOut)
	}
}

// ExecuteFiles executes template files with datafile
// and writes the filled data to the file if the file is specified
func (tmpl *Tmpl) ExecuteFiles() error {
	data := tmpl.Data
	dataFlattenMap := make(map[string]interface{})
	nestedToFlattenMap(data, dataFlattenMap, "", false)

	tmplFiles := tmpl.TmplOpts.TmplFiles
	for _, file := range tmplFiles {
		err := tmpl.Execute(file, dataFlattenMap)
		if err != nil {
			return err
		}
	}
	if tmpl.TmplOpts.DataOutFile != "" {
		dataOut, err := tmpl.marshalData(dataFlattenMap)
		if err != nil {
			return err
		}
		WriteStringToFileAndCreateDir(tmpl.TmplOpts.DataOutFile, dataOut, tmpl.TmplOpts.Overwrite)
	}
	return nil
}

// EnsureFiles check for missing keys in the template files
func (tmpl *Tmpl) EnsureFiles() error {
	tmplFiles := tmpl.TmplOpts.TmplFiles
	for _, file := range tmplFiles {
		err := tmpl.Ensure(file)
		if err != nil {
			return err
		}
	}
	return nil
}

// FillDestPath fill in the values for destination path
func (tmpl *Tmpl) FillDestPath(trimPrefixForDestPath string) {
	output := tmpl.TmplOpts.Output
	outdir := tmpl.TmplOpts.OutDir
	//ignoreDirOfOrigPath := tmpl.TmplOpts.IgnoreDirOfOrigPath
	if outdir == "" && output == "" {
		return
	}
	isMultipleTemplates := false
	if len(tmpl.Files) > 1 {
		isMultipleTemplates = true
	}
	for _, tmplMeta := range tmpl.Files {
		if outdir != "" {
			origPath := tmplMeta.OrigPath
			if trimPrefixForDestPath != "" {
				origPath = strings.TrimPrefix(origPath, trimPrefixForDestPath)
			}
			//if ignoreDirOfOrigPath {
			//	origPath = tmplMeta.Name
			//}
			if strings.HasSuffix(origPath, ".tpl") {
				origPath = strings.TrimSuffix(origPath, ".tpl")
			} else if strings.HasSuffix(origPath, ".tmpl") {
				origPath = strings.TrimSuffix(origPath, ".tmpl")
			}
			if output != "" && !isMultipleTemplates {
				origPath = output
			}
			tmplMeta.DestPath = outdir + string(os.PathSeparator) + origPath
		} else {
			tmplMeta.DestPath = output
		}
	}
}

// WriteProcessedTmpl writes processed template
func (tmpl *Tmpl) WriteProcessedTmpl() error {
	output := tmpl.TmplOpts.Output
	outdir := tmpl.TmplOpts.OutDir
	lenTmplFiles := len(tmpl.Files)
	if lenTmplFiles > 1 && output != "" && outdir == "" {
		return fmt.Errorf("multiple template files given with empty 'outdir' flag and non-empty 'out' flag")
	}
	c := InitializedNavColorMeta()
	for idx, tmplMeta := range tmpl.Files {
		if outdir == "" && output == "" {
			if idx > 0 {
				fmt.Printf("\n")
			}
			if tmpl.TmplOpts.ShowProcessedFile { // lenTmplFiles > 1 &&
				c.ExecInfo.Printf("'%s' is processed\n", tmplMeta.OrigPath)
			}
			fmt.Printf("%s", tmplMeta.Content)
			continue
		}
		err := WriteStringToFileAndCreateDir(tmplMeta.DestPath, tmplMeta.Content, tmpl.TmplOpts.Overwrite)
		//err := WriteStringToFileWithModeAndCreateDir(tmplMeta.DestPath, tmplMeta.Content, tmplMeta.Mode)
		if err != nil {
			switch err.(type) {
			case *ErrFileExists:
				continue
			default:
				return err
			}
		}
		if tmpl.TmplOpts.ShowProcessedFile {
			c.ExecInfo.Printf("'%s' is processed and stored in '%s'\n", tmplMeta.OrigPath, tmplMeta.DestPath)
		}
	}
	return nil
}
