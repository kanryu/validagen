package generator

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"text/template"

	"github.com/BurntSushi/toml"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

const OZZO_VALIDATOR_TEMPLATE string = "templates/ozzo-validator.tmpl"

// Project information to generate validators. Generated from Validator Toml
type ValidateProject struct {
	Type       string
	MethodName string
	Template   string
	Validators map[string]ValidateStruct
}

// Information on the struct that generates the validator and the validator go codes
type ValidateStruct struct {
	Package    string
	Name       string
	Dir        string
	FileName   string
	FileMode   int
	Import     []string
	MethodName string
	Properties map[string]ValidateProperty
}

// Validator settings to apply to a field
type ValidateProperty struct {
	Name string
	Type string
	ValidateRule
}

type TypedList struct {
	Int    []int
	Float  []float32
	String []string
}

type ValidateRule struct {
	In            TypedList
	NotIn         TypedList
	Length        []int
	RuneLength    []int
	Match         string
	Date          string
	Required      bool
	NotNil        bool
	Nil           bool
	NilOrNotEmpty bool
	Empty         bool
	MultipleOf    []TypedList
	Each          []ValidateRule
	//	When TODO
	Else []ValidateRule
	ValidateIsRule
}

type ValidateIsRule struct {
	Email            bool
	EmailFormat      bool
	URL              bool
	RequestURL       bool
	RequestURI       bool
	Alpha            bool
	Digit            bool
	Alphanumeric     bool
	UTFLetter        bool
	UTFDigit         bool
	UTFLetterNumeric bool
	UTFNumeric       bool
	LowerCase        bool
	UpperCase        bool
	Hexadecimal      bool
	HexColor         bool
	RGBColor         bool
	Int              bool
	Float            bool
	UUIDv3           bool
	UUIDv4           bool
	UUIDv5           bool
	UUID             bool
	CreditCard       bool
	ISBN10           bool
	ISBN13           bool
	ISBN             bool
	JSON             bool
	ASCII            bool
	PrintableASCII   bool
	Multibyte        bool
	FullWidth        bool
	HalfWidth        bool
	VariableWidth    bool
	Base64           bool
	DataURI          bool
	E164             bool
	CountryCode2     bool
	CountryCode3     bool
	DialString       bool
	MAC              bool
	IP               bool
	IPv4             bool
	IPv6             bool
	Subdomain        bool
	Domain           bool
	DNSName          bool
	Host             bool
	Port             bool
	MongoID          bool
	Latitude         bool
	Longitude        bool
	SSN              bool
	Semver           bool
}

// ParseToml parses the Toml file and generate ValidatorProject
func ParseToml(tomlPath string) (*ValidateProject, error) {
	var project ValidateProject
	_, err := toml.DecodeFile(tomlPath, &project)
	if err != nil {
		return nil, err
	}
	err = project.Validate()
	if err != nil {
		return nil, err
	}

	return &project, nil
}

// Generate validators defined in ValidatorProject.Validators in the specified directory
func (vp ValidateProject) Generate() error {
	tmpl, err := initTemplate(vp.Template)
	if err != nil {
		return err
	}
	for name, vs := range vp.Validators {
		validator_path, filemode := vp.prepare(&vs, name)
		vp.prepareIsValidators(&vs)
		// generate validator as a file
		slog.Info(fmt.Sprintf("generate %s", validator_path))
		f, err := os.Create(validator_path)
		f.Chmod(filemode)
		if err != nil {
			return err
		}
		err = tmpl.Execute(f, vs)
		f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// initTemplate Initialize the template for validator generation
func initTemplate(template_path string) (*template.Template, error) {
	if template_path == "" {
		template_path = OZZO_VALIDATOR_TEMPLATE
	}
	tmpl := template.New(template_path)
	funcs := template.FuncMap{
		"isValidNumericList": isValidNumericList,
		"isValidIntList":     isValidIntList,
		"isValidFloatList":   isValidFloatList,
		"isValidStringList":  isValidStringList,
	}
	tmpl.Funcs(funcs)
	if template_path == OZZO_VALIDATOR_TEMPLATE {
		data, err := AssetString(OZZO_VALIDATOR_TEMPLATE)
		if err != nil {
			return nil, err
		}
		_, err = tmpl.Parse(data)
		if err != nil {
			return nil, err
		}
	} else {
		slog.Debug("Custom Template", "FilePath", template_path)
		_, err := tmpl.ParseFiles(template_path)
		if err != nil {
			return nil, err
		}
	}
	return tmpl, nil
}

// prepare Specify the initial value of ValidateStruct
func (vp ValidateProject) prepare(vs *ValidateStruct, name string) (string, fs.FileMode) {
	if vs.Name == "" {
		vs.Name = name
	}
	if vs.MethodName == "" {
		vs.MethodName = vp.MethodName
	}
	if vs.MethodName == "" {
		vs.MethodName = "Validate"
	}
	validator_dir := vs.Package
	if vs.Dir != "" {
		validator_dir = vs.Dir
	}
	validator_filename := fmt.Sprintf("%s_validator.go", vs.Package)
	if vs.FileName != "" {
		validator_filename = vs.FileName
	}
	validator_path := fmt.Sprintf("%s/%s", validator_dir, validator_filename)
	filemode := fs.FileMode(0644)
	if vs.FileMode != 0 {
		filemode = fs.FileMode(vs.FileMode)
	}
	slog.Debug("ValidateProject",
		"Name", vs.Name,
		"MethodName", vs.MethodName,
		"ValidatorDir", validator_dir,
		"ValidatorFilename", validator_filename,
		"FileMode", filemode,
	)

	return validator_path, filemode
}

func (vp ValidateProject) prepareIsValidators(vs *ValidateStruct) {
	hasRegexp := false
	hasIsValidator := false
	for _, vp := range vs.Properties {
		if vp.Match != "" {
			hasRegexp = true
		}
		if vp.Email ||
			vp.EmailFormat ||
			vp.URL ||
			vp.RequestURL ||
			vp.RequestURI ||
			vp.Alpha ||
			vp.Digit ||
			vp.Alphanumeric ||
			vp.UTFLetter ||
			vp.UTFDigit ||
			vp.UTFLetterNumeric ||
			vp.UTFNumeric ||
			vp.LowerCase ||
			vp.UpperCase ||
			vp.Hexadecimal ||
			vp.HexColor ||
			vp.RGBColor ||
			vp.Int ||
			vp.Float ||
			vp.UUIDv3 ||
			vp.UUIDv4 ||
			vp.UUIDv5 ||
			vp.UUID ||
			vp.CreditCard ||
			vp.ISBN10 ||
			vp.ISBN13 ||
			vp.ISBN ||
			vp.JSON ||
			vp.ASCII ||
			vp.PrintableASCII ||
			vp.Multibyte ||
			vp.FullWidth ||
			vp.HalfWidth ||
			vp.VariableWidth ||
			vp.Base64 ||
			vp.DataURI ||
			vp.E164 ||
			vp.CountryCode2 ||
			vp.CountryCode3 ||
			vp.DialString ||
			vp.MAC ||
			vp.IP ||
			vp.IPv4 ||
			vp.IPv6 ||
			vp.Subdomain ||
			vp.Domain ||
			vp.DNSName ||
			vp.Host ||
			vp.Port ||
			vp.MongoID ||
			vp.Latitude ||
			vp.Longitude ||
			vp.SSN ||
			vp.Semver {
			hasIsValidator = true
		}
	}
	if hasRegexp {
		vs.Import = append(vs.Import, "regexp")
	}
	if hasIsValidator {
		vs.Import = append(vs.Import, "github.com/go-ozzo/ozzo-validation/v4/is")
	}
}

func (vp ValidateProject) Validate() error {
	return validation.ValidateStruct(&vp,
		validation.Field(&vp.Type, validation.Required, validation.In("struct", "map")),
		validation.Field(&vp.MethodName, is.Alphanumeric),
		validation.Field(&vp.Validators, validation.Required),
	)
}

func (vs ValidateStruct) Validate() error {
	return validation.ValidateStruct(&vs,
		validation.Field(&vs.Name, is.Alphanumeric), // can be nil, key name is substituted
		validation.Field(&vs.FileMode, validation.Min(0), validation.Max(0777)),
		validation.Field(&vs.Properties, validation.Required),
	)
}

func (vp ValidateProperty) Validate() error {
	return validation.ValidateStruct(&vp,
		validation.Field(&vp.Name, is.Alphanumeric), // can be nil, key name is substituted
		validation.Field(&vp.Type, validation.Required, validation.In("int", "float", "string", "object", "array")),
		validation.Field(&vp.Length, validation.By(checkLength)),
		validation.Field(&vp.RuneLength, validation.By(checkLength)),
	)
}

func checkLength(value interface{}) error {
	s, _ := value.([]int)
	switch len(s) {
	case 0:
		return nil
	case 2:
		if s[0] < 0 || s[1] < 0 {
			return errors.New("must be positive int(>=0)")
		}
		if s[0] > s[1] {
			return errors.New("must be [min, max]")
		}
	default:
		return errors.New("must be empty or 2-ints")
	}
	return nil
}

func isValidNumericList(value TypedList) bool {
	if len(value.Int) > 0 {
		return true
	}
	return len(value.Float) > 0
}
func isValidIntList(value TypedList) bool {
	return len(value.Int) > 0
}
func isValidFloatList(value TypedList) bool {
	return len(value.Float) > 0
}

func isValidStringList(value TypedList) bool {
	return len(value.String) > 0
}
