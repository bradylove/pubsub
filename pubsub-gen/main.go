package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/apoydence/pubsub/pubsub-gen/internal/generator"
	"github.com/apoydence/pubsub/pubsub-gen/internal/inspector"
)

func main() {
	structPath := flag.String("struct-name", "", "The name of the struct create a traverser for")
	packageName := flag.String("package", "", "The package name of the generated code")
	traverserName := flag.String("traverser", "", "The name of the generated traverser")
	output := flag.String("output", "", "The path to output the generated file")
	isPtr := flag.Bool("pointer", false, "Will the struct be a pointer when being published?")
	includePkgName := flag.Bool("include-pkg-name", false, "Prefix the struct type with the package name?")
	interfaces := flag.String("interfaces", "{}", "A map (map[string][]string encoded in JSON) mapping interface types to implementing structs")
	imports := flag.String("imports", "", "A comma separated list of imports required in the generated file")
	blacklist := flag.String("blacklist-fields", "", `A comma separated list of struct name and field
	combos to not include (e.g., mystruct.myfield,otherthing.otherfield).
	A wildcard (*) can be provided for the struct name (e.g., *.fieldname).`)

	flag.Parse()
	gopath := os.Getenv("GOPATH")

	if gopath == "" {
		log.Fatal("GOPATH is empty")
	}

	if *structPath == "" {
		log.Fatal("struct-name is required")
	}

	if *packageName == "" {
		log.Fatal("package is required")
	}

	if *traverserName == "" {
		log.Fatal("traverser is required")
	}

	if *output == "" {
		log.Fatal("output is required")
	}

	idx := strings.LastIndex(filepath.ToSlash(*structPath), ".")
	if idx < 0 {
		log.Fatalf("Invalid struct name: %s", *structPath)
	}

	mi := make(map[string][]string)
	if err := json.Unmarshal([]byte(*interfaces), &mi); err != nil {
		log.Fatalf("Invalid interfaces (%s): %s", *interfaces, err)
	}

	importList := strings.Split(*imports, ",")

	structName := (*structPath)[idx+1:]

	var pkgName string
	if *includePkgName {
		idx2 := strings.LastIndex(filepath.ToSlash(*structPath), "/")
		pkgName = filepath.ToSlash(*structPath)[idx2+1:idx] + "."
	}

	fieldBlacklist := buildBlacklist(*blacklist)

	sf := inspector.NewStructFetcher(fieldBlacklist)
	pp := inspector.NewPackageParser(sf)
	m, err := pp.Parse((*structPath)[:idx], gopath)
	if err != nil {
		log.Fatal(err)
	}

	linker := inspector.NewLinker()
	linker.Link(m, mi)

	g := generator.NewTraverserGenerator(generator.CodeWriter{})
	src, err := g.Generate(
		m,
		*packageName,
		*traverserName,
		structName,
		*isPtr,
		pkgName,
		importList,
	)
	if err != nil {
		log.Fatal(err)
	}

	pg := generator.NewPathGenerator()
	src, err = pg.Generate(src, m, *traverserName, structName)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(*output, []byte(src), 420)
	if err != nil {
		log.Fatal(err)
	}
}

func buildBlacklist(bl string) map[string][]string {
	if len(bl) == 0 {
		return nil
	}

	m := make(map[string][]string)
	for _, s := range strings.Split(bl, ",") {
		x := strings.Split(s, ".")
		if len(x) != 2 {
			log.Fatalf("'%s' is not in the proper format (structname.fieldname)", x)
		}

		m[x[0]] = append(m[x[0]], x[1])
	}
	return m
}
