/*
Copyright 2020 Institute for Automation of Complex Power Systems,
E.ON Energy Research Center, RWTH Aachen University

This project is licensed under either of
- Apache License, Version 2.0
- MIT License
at your option.

Apache License, Version 2.0:

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

MIT License:

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

// Package codegen provides a function for generating a Go package based on a previously created
// owl.GoModel.
package codegen

import (
	"fmt"
	"os"
	"strings"

	"git.rwth-aachen.de/acs/public/ontology/owl/owl2go/pkg/codegen/template"
	"git.rwth-aachen.de/acs/public/ontology/owl/owl2go/pkg/owl"
)

// GenerateGoCode generates the go package for a model
func GenerateGoCode(mod []owl.GoModel, path string) (err error) {
	fmt.Println("Generate Go Code")
	if len(mod) < 1 {
		return
	}

	// make dirs
	for i := range mod {
		err = os.MkdirAll(path+"/pkg/"+mod[i].Name, os.ModePerm)
		if err != nil {
			return
		}
	}
	// err = os.MkdirAll(path+"/docs", os.ModePerm)
	// if err != nil {
	// 	return
	// }
	err = os.MkdirAll(path+"/internal/helper", os.ModePerm)
	if err != nil {
		return
	}

	var file *os.File

	// README
	// file, err = os.Create(path + "/README.md")
	// if err != nil {
	// 	return
	// }
	// readme := generateReadme(mod[0].Name, mod[0].Description, mod[0].IRI)
	// fmt.Fprintln(file, readme)
	// file.Close()

	// module
	file, err = os.Create(path + "/go.mod")
	if err != nil {
		return
	}
	gomod := generateModule(mod[0].Module)
	fmt.Fprintln(file, gomod)
	file.Close()

	// internal
	file, err = os.Create(path + "/internal/helper/helper.go")
	if err != nil {
		return
	}
	help := generateHelper()
	fmt.Fprintln(file, template.OSSHeader+help)
	file.Close()

	for i := range mod {
		fmt.Println("\tGenerate package " + mod[i].Name)

		// model
		file, err = os.Create(path + "/pkg/" + mod[i].Name + "/model.go")
		if err != nil {
			return
		}
		model := generateModel(&mod[i])
		fmt.Fprintln(file, template.OSSHeader+model)
		file.Close()

		// imports
		if len(mod[i].Import) > 0 {
			file, err = os.Create(path + "/pkg/" + mod[i].Name + "/imports.go")
			if err != nil {
				return
			}
			imp := generateImport(&mod[i])
			fmt.Fprintln(file, template.OSSHeader+imp)
			file.Close()
		}

		// individuals
		file, err = os.Create(path + "/pkg/" + mod[i].Name + "/individuals.go")
		if err != nil {
			return
		}
		ind := generateIndividuals(&mod[i])
		fmt.Fprintln(file, template.OSSHeader+ind)
		file.Close()

		// Properties struct
		file, err = os.Create(path + "/pkg/" + mod[i].Name + "/propstruct.go")
		if err != nil {
			return
		}
		str, man, ser, ifc := generateProperties(&mod[i])
		fmt.Fprintln(file, template.OSSHeader+str)
		file.Close()

		// Properties interface
		file, err = os.Create(path + "/pkg/" + mod[i].Name + "/propinterface.go")
		if err != nil {
			return
		}
		fmt.Fprintln(file, template.OSSHeader+ifc)
		file.Close()

		// Properties manipulator
		file, err = os.Create(path + "/pkg/" + mod[i].Name + "/propmanipulator.go")
		if err != nil {
			return
		}
		fmt.Fprintln(file, template.OSSHeader+man)
		file.Close()

		// Properties serializer
		file, err = os.Create(path + "/pkg/" + mod[i].Name + "/propserializer.go")
		if err != nil {
			return
		}
		fmt.Fprintln(file, template.OSSHeader+ser)
		file.Close()

		// Classes
		for j := range mod[i].Class {
			file, err = os.Create(path + "/pkg/" + mod[i].Name + "/" + mod[i].Class[j].Name + ".go")
			if err != nil {
				return
			}
			class := generateClass(mod[i].Class[j], &mod[i])
			fmt.Fprintln(file, template.OSSHeader+class)
			file.Close()
		}

	}
	return
}

// generateReadme generates the readme file of the package
func generateReadme(ontName string, description string, iri string) (ret string) {
	ret = "# " + strings.ToTitle(ontName) + "\n\n"
	ret += "## [Link](" + iri + ")\n\n"
	ret += "## Description\n\n"
	ret += description + "\n\n"
	ret += "## Package\n\n"
	ret += "This package has been autogenerated using the ontology file located in folder `docs`."
	return
}

// generateModule writes the go mod file
func generateModule(name string) (ret string) {
	ret = "module " + name + "\n\n"
	ret += "require (\n"
	ret += "\tgit.rwth-aachen.de/acs/public/ontology/owl/owl2go v0.0.0-20200302081207-a47ddaf40a3c\n"
	ret += "\tgithub.com/piprate/json-gold v0.3.0\n"
	ret += ")\n\n"
	ret += "go 1.13"
	return
}

// generateHelper generates the helper functions in internal dir
func generateHelper() (ret string) {
	ret = template.HelperHeader + template.HelperAddToGraph
	ret += template.HelperAddClassPropertyToGraph + template.HelperAddIntPropertyToGraph +
		template.HelperAddFloatPropertyToGraph + template.HelperAddStringPropertyToGraph +
		template.HelperAddBoolPropertyToGraph + template.HelperAddInterfacePropertyToGraph +
		template.HelperAddDurationPropertyToGraph
	ret += strings.Replace(strings.Replace(template.HelperAddTimePropertyToGraph,
		"###timeType###", "DateTime", -1),
		"###timeLiteral###", template.LiteralDateTime, -1)
	ret += strings.Replace(strings.Replace(template.HelperAddTimePropertyToGraph,
		"###timeType###", "Date", -1),
		"###timeLiteral###", template.LiteralDate, -1)
	ret += strings.Replace(strings.Replace(template.HelperAddTimePropertyToGraph,
		"###timeType###", "DateTimeStamp", -1),
		"###timeLiteral###", template.LiteralDateTimeStamp, -1)
	ret += strings.Replace(strings.Replace(template.HelperAddTimePropertyToGraph,
		"###timeType###", "GYear", -1),
		"###timeLiteral###", template.LiteralGYear, -1)
	ret += strings.Replace(strings.Replace(template.HelperAddTimePropertyToGraph,
		"###timeType###", "GDay", -1),
		"###timeLiteral###", template.LiteralGDay, -1)
	ret += strings.Replace(strings.Replace(template.HelperAddTimePropertyToGraph,
		"###timeType###", "GYearMonth", -1),
		"###timeLiteral###", template.LiteralGYearMonth, -1)
	ret += strings.Replace(strings.Replace(template.HelperAddTimePropertyToGraph,
		"###timeType###", "GMonth", -1),
		"###timeLiteral###", template.LiteralGMonth, -1)
	ret += template.HelperParseXsdDuration
	ret += template.HelperIsIRI
	return
}

// generateModel generates model.go
func generateModel(mod *owl.GoModel) (ret string) {
	// Header
	ret = strings.Replace(template.ModelHeader, "###pkgName###", mod.Name, -1)
	imports := ""
	for i := range mod.Import {
		imports += "\tim" + mod.Import[i].Name + " \"" + mod.Module + "/pkg/" + mod.Import[i].Name +
			"\"\n"
	}
	ret = strings.Replace(ret, "###imports###", imports, -1)

	// Struct
	objectMaps := ""
	importModels := ""
	for i := range mod.Class {
		objectMaps += strings.Replace(template.StructMap, "###className###", mod.Class[i].Name, -1)
	}
	for i := range mod.Import {
		importModels += strings.Replace(template.StructImport, "###importName###",
			mod.Import[i].Name, -1)
	}
	ret += strings.Replace(strings.Replace(template.ModelStruct, "###objectMaps###",
		objectMaps, -1), "###importModels###", importModels, -1)

	// New Model
	makeMaps := ""
	newImportModels := ""
	for i := range mod.Class {
		makeMaps += strings.Replace(template.NewObjectMap, "###className###", mod.Class[i].Name, -1)
	}
	for i := range mod.Import {
		makeMaps += strings.Replace(template.NewImport, "###importName###", mod.Import[i].Name, -1)
	}
	ret += strings.Replace(strings.Replace(template.ModelNew, "###makeMaps###", makeMaps, -1),
		"###newImportModels###", newImportModels, -1)

	// model exist
	ret += template.ModelExists

	// ttl to model
	ret += template.ModelNewFromTTL

	// jsonld to model
	ret += template.ModelNewFromJSONLD

	// graph to model
	impRecv := getImportRecursive(mod)
	newObjects := ""
	for i := range mod.Class {
		newObjects += strings.Replace(strings.Replace(strings.Replace(template.NewObject,
			"###className###", mod.Class[i].Name, -1),
			"###capImportName###", "", -1),
			"###classIRI###", mod.Class[i].IRI, -1)
	}
	imps := make(map[string]interface{})
	for i := range impRecv {
		if _, ok := imps[impRecv[i].Name]; !ok {
			impPath := strings.Split(getImportPath(mod, impRecv[i].Name), ".")
			if len(impPath) > 1 {
				for j := range impRecv[i].Class {
					newObjects += strings.Replace(strings.Replace(strings.Replace(
						template.NewObject, "###className###", impRecv[i].Class[j].Name, -1),
						"###capImportName###", strings.Title(impPath[len(impPath)-1]), -1),
						"###classIRI###", impRecv[i].Class[j].IRI, -1)
				}
			}
		}
	}
	ret += strings.Replace(template.ModelNewFromGraph, "###newObjects###", newObjects, -1)

	// model to graph
	ret += template.ModelToGraph

	// model to ttl
	ret += template.ModelToTTL

	// model to jsonld
	ret += template.ModelToJSONLD

	// delete object
	deleteFromImports := ""
	deleteFromMaps := ""
	for i := range mod.Import {
		deleteFromImports += strings.Replace(template.DeleteFromImport, "###importName###",
			mod.Import[i].Name, -1)
	}
	for i := range mod.Class {
		deleteFromMaps += strings.Replace(template.DeleteFromMap, "###className###",
			mod.Class[i].Name, -1)
	}
	ret += strings.Replace(strings.Replace(template.ModelDeleteObject,
		"###deleteFromImports###", deleteFromImports, -1),
		"###deleteFromMaps###", deleteFromMaps, -1)

	// string
	ret += template.ModelString

	// ToDot
	replaceImports := strings.Replace(strings.Replace(template.ImportReplace,
		"###importName###", mod.Name, -1),
		"###importIRI###", mod.IRI+"#", -1)
	shapeImports := strings.Replace(template.ImportShape, "###importIRI###", mod.IRI+"#", -1)
	for i := range impRecv {
		replaceImports += strings.Replace(strings.Replace(template.ImportReplace,
			"###importName###", impRecv[i].Name, -1),
			"###importIRI###", impRecv[i].IRI+"#", -1)
		shapeImports += strings.Replace(template.ImportShape, "###importIRI###", impRecv[i].IRI+"#",
			-1)
	}
	ret += strings.Replace(strings.Replace(template.ModelToDot,
		"###importReplace###", replaceImports, -1),
		"###importShape###", shapeImports, -1)
	return
}

// generateImport generates import.go
func generateImport(mod *owl.GoModel) (ret string) {
	impRecv := getImportRecursive(mod)
	// Header
	ret = strings.Replace(template.ImportsHeader, "###pkgName###", mod.Name, -1)
	imports := ""
	for i := range impRecv {
		imports += "\tim" + impRecv[i].Name + " \"" + mod.Module + "/pkg/" + impRecv[i].Name +
			"\"\n"
	}
	ret = strings.Replace(ret, "###imports###", imports, -1)

	// New and Get methods
	imps := make(map[string]interface{})
	for i := range impRecv {
		if _, ok := imps[impRecv[i].Name]; !ok {
			impPath := strings.Split(getImportPath(mod, impRecv[i].Name), ".")
			if len(impPath) == 2 {
				for j := range impRecv[i].Class {
					ret += strings.Replace(strings.Replace(strings.Replace(strings.Replace(
						strings.Replace(template.ImportsNewGetMethods,
							"###className###", impRecv[i].Class[j].Name, -1),
						"###capImportName###", strings.Title(impPath[1]), -1),
						"###importName###", impPath[1], -1),
						"###importModelName###", impPath[1], -1),
						"###importCapImportName###", "", -1)
				}
			} else if len(impPath) > 2 {
				for j := range impRecv[i].Class {
					ret += strings.Replace(strings.Replace(strings.Replace(strings.Replace(
						strings.Replace(template.ImportsNewGetMethods,
							"###className###", impRecv[i].Class[j].Name, -1),
						"###capImportName###", strings.Title(impPath[len(impPath)-1]), -1),
						"###importName###", impPath[len(impPath)-1], -1),
						"###importModelName###", impPath[1], -1),
						"###importCapImportName###", strings.Title(impPath[len(impPath)-1]), -1)
				}
			}
		}
	}

	return
}

// generateIndividuals generates individuals.go
func generateIndividuals(mod *owl.GoModel) (ret string) {
	// Header
	ret = strings.Replace(template.Individual, "###pkgName###", mod.Name, -1)

	// individuals
	createIndividuals := ""
	for i := range mod.Individual {
		createIndividuals += strings.Replace(strings.Replace(template.CreateIndividual,
			"###individualType###", mod.Individual[i].Typ, -1),
			"###individualIRI###", mod.Individual[i].IRI, -1)
	}
	for i := range mod.Import {
		for j := range mod.Import[i].Individual {
			indType := mod.Import[i].Individual[j].Typ
			if mod.Import[i].Individual[j].ImportName != mod.Name {
				indType = strings.Title(mod.Import[i].Individual[j].ImportName) + indType
			}
			createIndividuals += strings.Replace(strings.Replace(template.CreateIndividual,
				"###individualType###", indType, -1),
				"###individualIRI###", mod.Import[i].Individual[j].IRI, -1)
		}
	}
	ret = strings.Replace(ret, "###createIndividuals###", createIndividuals, -1)
	return
}

// generateProperties generates propinterface.go, propmanipulator.go, propserializer.go and
// propstruct.go
func generateProperties(mod *owl.GoModel) (str, man, ser, ifc string) {
	strImport := make(map[string]string)
	manImport := make(map[string]string)
	serImport := make(map[string]string)
	ifcImport := make(map[string]string)
	for i := range mod.Class {
		for j := range mod.Class[i].Property {
			if mod.Class[i].Property[j].Multi ||
				mod.Class[i].Property[j].AllowedTyp[0][0] != mod.Class[i].Property[j].BaseTyp[0] {
				manImport["errors"] = ""
			}
			if mod.Class[i].Property[j].Typ[0] == "time.Time" ||
				mod.Class[i].Property[j].Typ[0] == "time.Duration" {
				manImport["time"] = ""
				strImport["time"] = ""
				ifcImport["time"] = ""
			}
			if mod.Class[i].Property[j].BaseTyp[0] == "owl.Thing" {
				strImport["git.rwth-aachen.de/acs/public/ontology/owl/owl2go/pkg/owl"] =
					""
				ifcImport["git.rwth-aachen.de/acs/public/ontology/owl/owl2go/pkg/owl"] =
					""
			}
			if mod.Class[i].Property[j].BaseTyp[0] == "float64" ||
				mod.Class[i].Property[j].BaseTyp[0] == "int" ||
				mod.Class[i].Property[j].BaseTyp[0] == "bool" {
				manImport["strconv"] = ""
			}
			if mod.Class[i].Property[j].Typ[0] == "time.Duration" {
				manImport[mod.Module+"/internal/helper"] = ""
			}
			temp := strings.Split(mod.Class[i].Property[j].BaseTyp[0], ".")
			if len(temp) == 2 && mod.Class[i].Property[j].BaseTyp[0] != "time.Time" &&
				mod.Class[i].Property[j].BaseTyp[0] != "time.Duration" &&
				mod.Class[i].Property[j].BaseTyp[0] != "owl.Thing" {
				ifcImport[mod.Module+"/pkg/"+strings.TrimPrefix(temp[0], "im")] = temp[0] + " "
			}
		}
		for j := range mod.Class[i].Imports {
			if mod.Class[i].Imports[j] != "" {
				manImport[j] = mod.Class[i].Imports[j]
				strImport[j] = mod.Class[i].Imports[j]
				ifcImport[j] = mod.Class[i].Imports[j]
			}
		}
	}
	manImport["git.rwth-aachen.de/acs/public/ontology/owl/owl2go/pkg/owl"] = ""
	serImport["git.rwth-aachen.de/acs/public/ontology/owl/owl2go/pkg/rdf"] = ""
	serImport[mod.Module+"/internal/helper"] = ""
	serImport["fmt"] = ""

	// Struct Header
	str = strings.Replace(template.PropertyHeader, "###pkgName###", mod.Name, -1)
	propImports := ""
	if len(strImport) > 0 {
		propImports += "import (\n"
		for i := range strImport {
			propImports += "\t" + strImport[i] + "\"" + i + "\"\n"
		}
		propImports += ")\n\n"
	}
	str = strings.Replace(str, "###propImports###", propImports, -1)
	str += template.PropertyStructCommon

	// Manipulator Header
	man = strings.Replace(template.PropertyHeader, "###pkgName###", mod.Name, -1)
	propImports = ""
	if len(manImport) > 0 {
		propImports += "import (\n"
		for i := range manImport {
			propImports += "\t" + manImport[i] + "\"" + i + "\"\n"
		}
		propImports += ")\n\n"
	}
	man = strings.Replace(man, "###propImports###", propImports, -1)
	man += template.PropertyIRI

	// Serializer Header
	ser = strings.Replace(template.PropertyHeader, "###pkgName###", mod.Name, -1)
	propImports = ""
	if len(serImport) > 0 {
		propImports += "import (\n"
		for i := range serImport {
			propImports += "\t" + serImport[i] + "\"" + i + "\"\n"
		}
		propImports += ")\n\n"
	}
	ser = strings.Replace(ser, "###propImports###", propImports, -1)

	// Interface Header
	ifc = strings.Replace(template.PropertyHeader, "###pkgName###", mod.Name, -1)
	propImports = ""
	if len(ifcImport) > 0 {
		propImports += "import (\n"
		for i := range ifcImport {
			propImports += "\t" + ifcImport[i] + "\"" + i + "\"\n"
		}
		propImports += ")\n\n"
	}
	ifc = strings.Replace(ifc, "###propImports###", propImports, -1)

	stor := make(map[string]interface{})
	ifcstor := make(map[string]interface{})
	for i := range mod.Class {
		class := mod.Class[i]
		for j := range class.Property {
			prop := class.Property[j]
			propName := generatePropertyName(prop)
			if _, ok := stor[propName]; !ok {
				stor[propName] = nil
				// Struct
				str += generatePropertyStruct(prop)
				// Manipulator
				man += generatePropertyManipulator(prop)
				// Serializer
				ser += generatePropertySerializer(prop)
			}
			mult := ""
			if prop.Multi {
				mult = "Multi"
			} else {
				mult = "Single"
			}
			if _, ok := ifcstor[prop.Capital+mult+prop.BaseTyp[0]]; !ok {
				ifcstor[prop.Capital+mult+prop.BaseTyp[0]] = nil
				ifc += generatePropertyInterface(prop)
			}
		}
	}
	return
}

// generatePropertyName generates the name of a property based on the type, basetype and allowed
// types
func generatePropertyName(prop owl.GoProperty) (ret string) {
	ret = "prop" + prop.Capital + "Base"
	if prop.BaseTyp[0] == "time.Time" {
		ret += "GoTime"
	} else if prop.BaseTyp[0] == "time.Duration" {
		ret += "GoDuration"
	} else if prop.BaseTyp[0] == "interface{}" {
		ret += "interface"
	} else {
		temp := strings.Split(prop.BaseTyp[0], ".")
		ret += temp[len(temp)-1]
	}
	ret += "Type"
	if prop.Typ[0] == "time.Time" {
		ret += "GoTime"
	} else if prop.Typ[0] == "time.Duration" {
		ret += "GoDuration"
	} else if prop.Typ[0] == "interface{}" {
		ret += "interface"
	} else {
		temp := strings.Split(prop.Typ[0], ".")
		ret += temp[len(temp)-1]
	}
	if prop.Multi {
		ret += "Multiple"
	} else {
		ret += "Single"
	}
	for i := range prop.AllowedTyp {
		if prop.AllowedTyp[i][0] == "time.Time" {
			ret += "GoTime"
		} else if prop.AllowedTyp[i][0] == "time.Duration" {
			ret += "GoDuration"
		} else if prop.AllowedTyp[i][0] == "interface{}" {
			ret += "interface"
		} else {
			temp := strings.Split(prop.AllowedTyp[i][0], ".")
			ret += temp[len(temp)-1]
		}
	}
	return
}

// generatePropertyStruct generates the struct that belongs to a property
func generatePropertyStruct(prop owl.GoProperty) (ret string) {
	propName := generatePropertyName(prop)
	if prop.Multi {
		if prop.Typ[0] == "time.Time" || prop.Typ[0] == "time.Duration" ||
			prop.Typ[0] == "float64" || prop.Typ[0] == "string" || prop.Typ[0] == "int" ||
			prop.Typ[0] == "interface{}" {
			ret += strings.Replace(strings.Replace(strings.Replace(strings.Replace(
				template.PropertyStructMultipleLiteral, "###propName###", prop.Name, -1),
				"###propType###", prop.Typ[0], -1),
				"###propLongName###", propName, -1),
				"###comment###", prop.Comment, -1)
		} else {
			ret += strings.Replace(strings.Replace(strings.Replace(strings.Replace(
				template.PropertyStructMultipleClass, "###propName###", prop.Name, -1),
				"###propType###", prop.Typ[0], -1),
				"###propLongName###", propName, -1),
				"###comment###", prop.Comment, -1)
		}
	} else {
		ret += strings.Replace(strings.Replace(strings.Replace(strings.Replace(
			template.PropertyStructSingle, "###propName###", prop.Name, -1),
			"###propType###", prop.Typ[0], -1),
			"###propLongName###", propName, -1),
			"###comment###", prop.Comment, -1)
	}
	return
}

// generatePropertyInterface generates the interface that belongs to a property
func generatePropertyInterface(prop owl.GoProperty) (ret string) {
	baseTypeNoImp := prop.BaseTyp[0]
	temp := strings.Split(prop.BaseTyp[0], ".")
	if len(temp) > 0 {
		baseTypeNoImp = temp[len(temp)-1]
	}
	if baseTypeNoImp == "interface{}" {
		baseTypeNoImp = "interface"
	}
	if prop.Multi {
		ret = strings.Replace(strings.Replace(strings.Replace(strings.Replace(strings.Replace(
			template.PropertyInterfaceMultiple, "###propCapital###", prop.Capital, -1),
			"###propName###", prop.Name, -1),
			"###propBaseTypeNoImp###", baseTypeNoImp, -1),
			"###propBaseType###", prop.BaseTyp[0], -1),
			"###comment###", prop.Comment, -1)
	} else {
		ret = strings.Replace(strings.Replace(strings.Replace(strings.Replace(strings.Replace(
			template.PropertyInterfaceSingle, "###propCapital###", prop.Capital, -1),
			"###propName###", prop.Name, -1),
			"###propBaseTypeNoImp###", baseTypeNoImp, -1),
			"###propBaseType###", prop.BaseTyp[0], -1),
			"###comment###", prop.Comment, -1)
	}
	return
}

// generatePropertyManipulator generates the manipulator functions for a property
func generatePropertyManipulator(prop owl.GoProperty) (ret string) {
	propName := generatePropertyName(prop)
	if prop.Multi {
		if prop.Typ[0] == "time.Time" || prop.Typ[0] == "time.Duration" ||
			prop.Typ[0] == "float64" || prop.Typ[0] == "string" || prop.Typ[0] == "int" ||
			prop.Typ[0] == "interface{}" {
			if prop.Inverse != "" {
				ret = template.PropertyGetMultipleLiteral
			} else {
				ret = template.PropertyGetMultipleLiteral + template.PropertySetMultipleLiteral +
					template.PropertyAddLiteral + template.PropertyDelLiteral
			}
		} else {
			if prop.Inverse != "" {
				ret = template.PropertyGetMultipleClass
			} else {
				ret = template.PropertyGetMultipleClass + template.PropertySetMultipleClass
				if len(prop.AllowedTyp) == 1 && prop.AllowedTyp[0] == prop.BaseTyp {
					ret += template.PropertyAddClassSingle + template.PropertyDelClassSingle
				} else {
					addClassMultiple := ""
					delClassMultiple := ""
					allTypes := ""
					for i := range prop.AllowedTyp {
						addClassMultiple += strings.Replace(template.AddClassMultiple,
							"###propAllowedType###", prop.AllowedTyp[i][0], -1)
						delClassMultiple += strings.Replace(template.DelClassMultiple,
							"###propAllowedType###", prop.AllowedTyp[i][0], -1)
						allTypes += prop.AllowedTyp[i][0] + ", "
					}
					ret += strings.Replace(template.PropertyAddClassMultiple,
						"###addClassMultiple###", addClassMultiple, -1)
					ret += strings.Replace(template.PropertyDelClassMultiple,
						"###delClassMultiple###", delClassMultiple, -1)
					ret = strings.Replace(ret, "###propAllowedTypes###", allTypes, -1)
				}
			}
		}
	} else {
		ret += template.PropertyGetSingle
		if prop.Typ[0] == "time.Time" || prop.Typ[0] == "time.Duration" ||
			prop.Typ[0] == "float64" || prop.Typ[0] == "string" || prop.Typ[0] == "int" ||
			prop.Typ[0] == "interface{}" && prop.Inverse == "" {
			if len(prop.AllowedTyp) == 1 && prop.AllowedTyp[0] == prop.BaseTyp {
				ret += template.PropertySetSingleLiteral
			} else {
				setSingleLiteralMultiple := ""
				allTypes := ""
				for i := range prop.AllowedTyp {
					setSingleLiteralMultiple += strings.Replace(template.SetSingleLiteralMultiple,
						"###propAllowedType###", prop.AllowedTyp[i][0], -1)
					allTypes += prop.AllowedTyp[i][0] + ", "
				}
				ret += strings.Replace(template.PropertySetSingleLiteralMultiple,
					"###setSingleLiteralMultiple###", setSingleLiteralMultiple, -1)
				ret = strings.Replace(ret, "###propAllowedTypes###", allTypes, -1)
			}
		} else if prop.Inverse == "" {
			if len(prop.AllowedTyp) == 1 && prop.AllowedTyp[0] == prop.BaseTyp {
				ret += template.PropertySetSingleClassSingle
			} else {
				setSingleClassMultiple := ""
				allTypes := ""
				for i := range prop.AllowedTyp {
					setSingleClassMultiple += strings.Replace(template.SetSingleClassMultiple,
						"###propAllowedType###", prop.AllowedTyp[i][0], -1)
					allTypes += prop.AllowedTyp[i][0] + ", "
				}
				ret += strings.Replace(template.PropertySetSingleClassMultiple,
					"###setSingleClassMultiple###", setSingleClassMultiple, -1)
				ret = strings.Replace(ret, "###propAllowedTypes###", allTypes, -1)
			}
		}
	}
	ret = strings.Replace(strings.Replace(strings.Replace(strings.Replace(strings.Replace(
		strings.Replace(ret, "###propName###", prop.Name, -1),
		"###propType###", prop.Typ[0], -1),
		"###propBaseType###", prop.BaseTyp[0], -1),
		"###propLongName###", propName, -1),
		"###comment###", prop.Comment, -1),
		"###propCapital###", prop.Capital, -1)

	if prop.Inverse == "" {
		baseType := prop.BaseTyp[0]
		initProp := ""
		mult := ""
		if prop.Multi {
			mult = template.MultiplicityMultiple
		} else {
			mult = template.MultiplicitySingle
		}
		switch prop.Typ[0] {
		case "time.Time":
			switch prop.XSDTyp {
			case "http://www.w3.org/2001/XMLSchema#dateTime":
				initProp = strings.Replace(template.PropertyInitLiteral, "###PropInit###",
					template.PropInitDateTime, -1)
			case "http://www.w3.org/2001/XMLSchema#date":
				initProp = strings.Replace(template.PropertyInitLiteral, "###PropInit###",
					template.PropInitDate, -1)
			case "http://www.w3.org/2001/XMLSchema#dateTimeStamp":
				initProp = strings.Replace(template.PropertyInitLiteral, "###PropInit###",
					template.PropInitDateTimeStamp, -1)
			case "http://www.w3.org/2001/XMLSchema#gYear":
				initProp = strings.Replace(template.PropertyInitLiteral, "###PropInit###",
					template.PropInitGYear, -1)
			case "http://www.w3.org/2001/XMLSchema#gDay":
				initProp = strings.Replace(template.PropertyInitLiteral, "###PropInit###",
					template.PropInitGDay, -1)
			case "http://www.w3.org/2001/XMLSchema#gYearMonth":
				initProp = strings.Replace(template.PropertyInitLiteral, "###PropInit###",
					template.PropInitGYearMonth, -1)
			case "http://www.w3.org/2001/XMLSchema#gMonth":
				initProp = strings.Replace(template.PropertyInitLiteral, "###PropInit###",
					template.PropInitGMonth, -1)
			}
		case "time.Duration":
			initProp = strings.Replace(template.PropertyInitLiteral, "###PropInit###",
				template.PropInitDuration, -1)
		case "int":
			initProp = strings.Replace(template.PropertyInitLiteral, "###PropInit###",
				template.PropInitInt, -1)
		case "float64":
			initProp = strings.Replace(template.PropertyInitLiteral, "###PropInit###",
				template.PropInitFloat, -1)
		case "bool":
			initProp = strings.Replace(template.PropertyInitLiteral, "###PropInit###",
				template.PropInitBool, -1)
		case "string":
			initProp = strings.Replace(template.PropertyInitLiteral, "###PropInit###",
				template.PropInitString, -1)
		case "interface{}":
			initProp = strings.Replace(template.PropertyInitLiteral, "###PropInit###",
				template.PropInitInterface, -1)
		default:
			tempSp := strings.Split(prop.BaseTyp[0], ".")
			if prop.BaseTyp[0] == "owl.Thing" {
				initProp = strings.Replace(template.PropertyInitClass, "###PropInit###",
					template.PropInitClassBaseThing, -1)
			} else if len(tempSp) > 1 {
				imName := strings.TrimPrefix(tempSp[0], "im")
				baseType = tempSp[1]
				initProp = strings.Replace(template.PropertyInitClass, "###PropInit###",
					template.PropInitClassImport, -1)
				initProp = strings.Replace(initProp, "###capImportName###", strings.Title(imName),
					-1)
			} else {
				initProp = strings.Replace(template.PropertyInitClass, "###PropInit###",
					template.PropInitClassDefault, -1)
			}
		}
		ret += strings.Replace(strings.Replace(strings.Replace(strings.Replace(initProp,
			"###Multiplicity###", mult, -1),
			"###propLongName###", propName, -1),
			"###propBaseType###", baseType, -1),
			"###propCapital###", prop.Capital, -1)

		if prop.Typ[0] == "time.Time" || prop.Typ[0] == "time.Duration" ||
			prop.Typ[0] == "float64" || prop.Typ[0] == "string" || prop.Typ[0] == "int" &&
			prop.Inverse == "" || prop.Typ[0] == "bool" || prop.Typ[0] == "interface{}" {
			return
		}
		if prop.Multi {
			ret += strings.Replace(strings.Replace(strings.Replace(template.PropertyMultipleRemove,
				"###propLongName###", propName, -1),
				"###propBaseType###", prop.BaseTyp[0], -1),
				"###propCapital###", prop.Capital, -1)
		} else {
			ret += strings.Replace(strings.Replace(strings.Replace(strings.Replace(
				template.PropertySingleRemove, "###propLongName###", propName, -1),
				"###propBaseType###", prop.BaseTyp[0], -1),
				"###propName###", prop.Name, -1),
				"###propCapital###", prop.Capital, -1)
		}
	}
	return
}

// generatePropertySerializer generates the serializer functions that belong to a property
func generatePropertySerializer(prop owl.GoProperty) (ret string) {
	propName := generatePropertyName(prop)
	graphProp := ""
	stringProp := ""
	switch prop.Typ[0] {
	case "string":
		graphProp = template.GraphPropString
		stringProp = template.StringPropString
	case "float64":
		graphProp = template.GraphPropFloat
		stringProp = template.StringPropFloat
	case "int":
		graphProp = template.GraphPropInt
		stringProp = template.StringPropInt
	case "bool":
		graphProp = template.GraphPropBool
		stringProp = template.StringPropBool
	case "interface{}":
		graphProp = template.GraphPropInterface
		stringProp = template.StringPropInterface
	case "time.Time", "time.Duration":
		stringProp = template.StringPropTime
		switch prop.XSDTyp {
		case "http://www.w3.org/2001/XMLSchema#dateTime":
			graphProp = template.GraphPropSDateTime
		case "http://www.w3.org/2001/XMLSchema#date":
			graphProp = template.GraphPropSDate
		case "http://www.w3.org/2001/XMLSchema#duration":
			graphProp = template.GraphPropSDuration
		case "http://www.w3.org/2001/XMLSchema#dateTimeStamp":
			graphProp = template.GraphPropSDateTimeStamp
		case "http://www.w3.org/2001/XMLSchema#gYear":
			graphProp = template.GraphPropSGYear
		case "http://www.w3.org/2001/XMLSchema#gDay":
			graphProp = template.GraphPropSGDay
		case "http://www.w3.org/2001/XMLSchema#gYearMonth":
			graphProp = template.GraphPropSGYearMonth
		case "http://www.w3.org/2001/XMLSchema#gMonth":
			graphProp = template.GraphPropSGMonth
		}
	default:
		graphProp = template.GraphPropClass
		if prop.Multi {
			stringProp = template.StringPropClassMultiple
		} else {
			stringProp = template.StringPropClassSingle
		}
	}
	indent := ""
	array := ""
	if prop.Multi {
		ret = strings.Replace(template.PropertyGraphMultiple, "###graphProp###", graphProp, -1)
		ret += strings.Replace(template.PropertyStringMultiple, "###stringProp###", stringProp, -1)
		indent = template.IndentMultiple
		array = template.ArrayMultiple
	} else {
		ret = strings.Replace(template.PropertyGraphSingle, "###graphProp###", graphProp, -1)
		ret += strings.Replace(template.PropertyStringSingle, "###stringProp###", stringProp, -1)
	}
	ret = strings.Replace(strings.Replace(strings.Replace(strings.Replace(strings.Replace(
		strings.Replace(strings.Replace(strings.Replace(ret, "###indent###", indent, -1),
			"###array###", array, -1),
			"###propName###", prop.Name, -1),
		"###propIRI###", prop.IRI, -1),
		"###propType###", prop.Typ[0], -1),
		"###propBaseType###", prop.BaseTyp[0], -1),
		"###propLongName###", propName, -1),
		"###propCapital###", prop.Capital, -1)
	return
}

// generateClass generates the <class>.go file
func generateClass(class owl.GoClass, mod *owl.GoModel) (ret string) {
	singleParent := false
	if len(class.DirectParent) == 1 {
		if _, ok := mod.Class[class.DirectParent[0]]; ok {
			singleParent = true
		}
	}
	isExactChild := false
	if singleParent {
		for i := range mod.Class[class.DirectParent[0]].Property {
			for j := range class.Property {
				if mod.Class[class.DirectParent[0]].Property[i].Name == class.Property[j].Name {
					if mod.Class[class.DirectParent[0]].Property[i].BaseTyp[0] ==
						class.Property[j].BaseTyp[0] &&
						mod.Class[class.DirectParent[0]].Property[i].Typ[0] ==
							class.Property[j].Typ[0] &&
						len(mod.Class[class.DirectParent[0]].Property[i].AllowedTyp) ==
							len(class.Property[j].AllowedTyp) {
						equalAllowedTypes := true
						for k := range class.Property[j].AllowedTyp {
							if class.Property[j].AllowedTyp[k][0] !=
								mod.Class[class.DirectParent[0]].Property[i].AllowedTyp[k][0] {
								equalAllowedTypes = false
								break
							}
						}
						if equalAllowedTypes {
							isExactChild = true
						} else {
							isExactChild = false
						}
					} else {
						isExactChild = false
					}
					break
				}
			}
			if !isExactChild {
				break
			}
		}
	}
	equalParentProps := false
	if isExactChild {
		if len(class.Property) == len(mod.Class[class.DirectParent[0]].Property) {
			equalParentProps = true
		}
	}

	// Header
	ret = strings.Replace(template.ClassHeader, "###pkgName###", mod.Name, -1)
	imports := ""
	for i := range class.Imports {
		if equalParentProps && class.Imports[i] != "" {
			continue
		}
		imports += "\t" + class.Imports[i] + "\"" + i + "\"\n"
	}
	imports += "\t\"errors\"\n"
	imports += "\t\"git.rwth-aachen.de/acs/public/ontology/owl/owl2go/pkg/rdf\"\n"
	if !equalParentProps {
		imports += "\t\"git.rwth-aachen.de/acs/public/ontology/owl/owl2go/pkg/owl\"\n"
	}
	imports += "\t\"" + mod.Module + "/internal/helper\"\n"
	imports += "\t\"strings\"\n"
	ret = strings.Replace(ret, "###imports###", imports, -1)

	// interface
	interfaceMethods := ""
	if !singleParent {
		interfaceMethods += "\towl.Thing\n"
	} else {
		interfaceMethods += "\t" + class.DirectParent[0] + "\n"
	}
	for i := range class.Property {
		prop := class.Property[i]
		isParentProp := false
		if singleParent {
			for j := range mod.Class[class.DirectParent[0]].Property {
				if prop.Name == mod.Class[class.DirectParent[0]].Property[j].Name {
					isParentProp = true
					break
				}
			}
		}
		if !isParentProp {
			multi := "Single"
			if class.Property[i].Multi {
				multi = "Multiple"
			}
			baseTypeNoImp := prop.BaseTyp[0]
			temp := strings.Split(prop.BaseTyp[0], ".")
			if len(temp) > 0 {
				baseTypeNoImp = temp[len(temp)-1]
			}
			if baseTypeNoImp == "interface{}" {
				baseTypeNoImp = "interface"
			}
			interfaceMethods += strings.Replace(strings.Replace(strings.Replace(
				template.InterfaceInterface, "###propName###", prop.Name, -1),
				"###propBaseTypeNoImp###", baseTypeNoImp, -1),
				"###multi###", multi, -1)
		}
	}
	interfaceInheritance := strings.Replace(template.InterfaceInheritance, "###parentName###",
		class.Name, -1)
	parents := make(map[string]interface{})
	if !singleParent {
		for i := range class.Parent {
			tempSp := strings.Split(class.Parent[i], ".")
			if len(tempSp) > 1 {
				if _, ok := parents[tempSp[1]]; !ok {
					parents[tempSp[1]] = nil
					interfaceInheritance += strings.Replace(template.InterfaceInheritance,
						"###parentName###", tempSp[1], -1)
				}
			} else {
				if _, ok := parents[class.Parent[i]]; !ok {
					parents[class.Parent[i]] = nil
					interfaceInheritance += strings.Replace(template.InterfaceInheritance,
						"###parentName###", class.Parent[i], -1)
				}
			}
		}
	}
	ret += strings.Replace(strings.Replace(strings.Replace(template.ClassInterface,
		"###comment###", class.Comment, -1),
		"###interfaceMethods###", interfaceMethods, -1),
		"###interfaceInheritance###", interfaceInheritance, -1)

	// Struct
	structProperties := ""
	if !isExactChild {
		structProperties += "\tpropCommon\n"
	} else {
		structProperties += "\ts" + class.DirectParent[0] + "\n"
	}
	for i := range class.Property {
		isParentProp := false
		if isExactChild {
			for j := range mod.Class[class.DirectParent[0]].Property {
				if class.Property[i].Name == mod.Class[class.DirectParent[0]].Property[j].Name {
					isParentProp = true
					break
				}
			}
		}
		if !isParentProp {
			structProperties += strings.Replace(template.StructProperty, "###propLongName###",
				generatePropertyName(class.Property[i]), -1)
		}
	}
	ret += strings.Replace(strings.Replace(template.ClassStruct,
		"###comment###", class.Comment, -1),
		"###structProperties###", structProperties, -1)

	// New
	newMakeMaps := ""
	newInitProps := ""
	if isExactChild {
		newMakeMaps += "\tres.s" + class.DirectParent[0] + ".makeMaps()\n"
	}
	for i := range class.Property {
		prop := class.Property[i]
		isParentProp := false
		if isExactChild {
			for j := range mod.Class[class.DirectParent[0]].Property {
				if class.Property[i].Name == mod.Class[class.DirectParent[0]].Property[j].Name {
					isParentProp = true
					break
				}
			}
		}
		if isParentProp {
			continue
		}
		if prop.Multi && !(prop.Typ[0] == "time.Time" || prop.Typ[0] == "time.Duration" ||
			prop.Typ[0] == "float64" || prop.Typ[0] == "string" || prop.Typ[0] == "int" ||
			prop.Typ[0] == "interface{}") {
			newMakeMaps += strings.Replace(strings.Replace(template.NewMakeMap,
				"###propName###", prop.Name, -1),
				"###propType###", prop.Typ[0], -1)
		}
		if len(prop.Individual) > 0 {
			for j := range prop.Individual {
				switch prop.Typ[0] {
				case "string":
					if prop.Multi {
						newInitProps += strings.Replace(template.NewInitPropLiteralMultiple,
							"###value###", "\""+prop.Individual[j]+"\"", -1)
					} else {
						newInitProps += strings.Replace(template.NewInitPropLiteralSingle,
							"###value###", "\""+prop.Individual[j]+"\"", -1)
					}
				case "time.Time", "time.Duration":
				case "float64", "int":
					if prop.Multi {
						newInitProps += template.NewInitPropLiteralMultiple
					} else {
						newInitProps += template.NewInitPropLiteralSingle
					}
				default:
					if prop.Multi {
						newInitProps += template.NewInitPropClassMultiple
					} else {
						newInitProps += template.NewInitPropClassSingle
					}
				}
				newInitProps = strings.Replace(strings.Replace(strings.Replace(newInitProps,
					"###propCapital###", prop.Capital, -1),
					"###propType###", prop.Typ[0], -1),
					"###value###", prop.Individual[j], -1)
			}
		}
	}
	ret += template.ClassNew
	if !equalParentProps {
		ret += strings.Replace(strings.Replace(template.ClassMakeMaps,
			"###newMakeMaps###", newMakeMaps, -1),
			"###newInitProps###", newInitProps, -1)
	}

	// Add
	newAddToMaps := strings.Replace(template.AddToMap, "###parentName###", class.Name, -1)
	if singleParent {
		newAddToMaps += "\tmod.add" + class.DirectParent[0] + "(res)\n"
	} else {
		for i := range class.Parent {
			tempSp := strings.Split(class.Parent[i], ".")
			if len(tempSp) > 1 {

			} else {
				newAddToMaps += strings.Replace(template.AddToMap, "###parentName###",
					class.Parent[i], -1)
			}
		}
		newAddToMaps += strings.Replace(template.AddToMap, "###parentName###", "Thing", -1)
	}
	ret += strings.Replace(template.ClassAdd, "###newAddToMaps###", newAddToMaps, -1)

	// Get
	ret += template.ClassGet

	// Remove
	removeProps := ""
	parentRemove := ""
	if isExactChild {
		parentRemove += "\t\tres.s" + class.DirectParent[0] + ".RemoveObject(obj, prop)\n"
	}
	for i := range class.Property {
		prop := class.Property[i]
		isParentProp := false
		if isExactChild {
			for j := range mod.Class[class.DirectParent[0]].Property {
				if class.Property[i].Name == mod.Class[class.DirectParent[0]].Property[j].Name {
					isParentProp = true
					break
				}
			}
		}
		if isParentProp {
			continue
		}
		if prop.BaseTyp[0] == "string" || prop.BaseTyp[0] == "float64" ||
			prop.BaseTyp[0] == "int" || prop.BaseTyp[0] == "time.Time" ||
			prop.BaseTyp[0] == "time.Duration" || prop.BaseTyp[0] == "bool" ||
			prop.BaseTyp[0] == "interface{}" {
			continue
		}
		if prop.Inverse == "" {
			removeProps += strings.Replace(strings.Replace(template.RemoveNonInverse,
				"###propLongName", generatePropertyName(prop), -1),
				"###propIRI###", prop.IRI, -1)
			continue
		}
		temp := ""
		if prop.Multi {
			temp = template.RemovePropMultiple
		} else {
			temp = template.RemovePropSingle
		}
		removeProps += strings.Replace(strings.Replace(strings.Replace(strings.Replace(temp,
			"###propIRI###", prop.IRI, -1),
			"###propBaseType###", prop.BaseTyp[0], -1),
			"###propName###", prop.Name, -1),
			"###propCapital###", prop.Capital, -1)
	}
	if !equalParentProps {
		ret += strings.Replace(strings.Replace(template.ClassRemove,
			"###removeProps###", removeProps, -1),
			"###parentRemove###", parentRemove, -1)
	}

	// inheritance
	ret += strings.Replace(template.ClassInheritance, "###parentName###", class.Name, -1)
	if !isExactChild {
		parents = make(map[string]interface{})
		for i := range class.Parent {
			tempSp := strings.Split(class.Parent[i], ".")
			if len(tempSp) > 1 {
				if _, ok := parents[tempSp[1]]; !ok {
					parents[tempSp[1]] = nil
					ret += strings.Replace(template.ClassInheritance, "###parentName###",
						tempSp[1], -1)
				}
			} else {
				if _, ok := parents[class.Parent[i]]; !ok {
					parents[class.Parent[i]] = nil
					ret += strings.Replace(template.ClassInheritance, "###parentName###",
						class.Parent[i], -1)
				}
			}
		}
	}

	// inverse
	for i := range class.Property {
		prop := class.Property[i]
		isParentProp := false
		if isExactChild {
			for j := range mod.Class[class.DirectParent[0]].Property {
				if class.Property[i].Name == mod.Class[class.DirectParent[0]].Property[j].Name {
					isParentProp = true
					break
				}
			}
		}
		if prop.Inverse != "" && !isParentProp {
			propName := generatePropertyName(prop)
			temp := ""
			if class.Property[i].Multi {
				if prop.Typ[0] == "time.Time" || prop.Typ[0] == "time.Duration" ||
					prop.Typ[0] == "float64" || prop.Typ[0] == "string" {
					ret = ""
					continue
				}
				if len(prop.AllowedTyp) == 1 && prop.AllowedTyp[0] == prop.BaseTyp {
					temp = template.ClassInverseMultipleSingle
				} else {
					inverseAddMultipleMultipleAllowed := ""
					inverseDelMultipleMultipleAllowed := ""
					allowedString := "["
					for j := range prop.AllowedTyp {
						inverseAddMultipleMultipleAllowed += strings.Replace(
							template.InverseAddMultipleMultiple, "###propAllowedType###",
							prop.AllowedTyp[j][0], -1)
						inverseDelMultipleMultipleAllowed += strings.Replace(
							template.InverseDelMultipleMultiple, "###propAllowedType###",
							prop.AllowedTyp[j][0], -1)
						allowedString += prop.AllowedTyp[j][0] + ", "
					}
					allowedString += "]"
					temp = strings.Replace(strings.Replace(strings.Replace(
						template.ClassInverseMultipleMultiple,
						"###inverseAddMultipleMultipleAllowed###",
						inverseAddMultipleMultipleAllowed, -1),
						"###inverseDelMultipleMultipleAllowed###",
						inverseDelMultipleMultipleAllowed, -1),
						"###allowedTypes###", allowedString, -1)
				}
			} else {
				if len(prop.AllowedTyp) == 1 && prop.AllowedTyp[0] == prop.BaseTyp {
					temp = template.ClassInverseSingleSingle
				} else {
					inverseSetSingleMultipleOne := ""
					inverseSetSingleMultipleTwo := ""
					inverseSetSingleMultipleThree := ""
					for j := range prop.AllowedTyp {
						inverseSetSingleMultipleOne += strings.Replace(
							template.InverseSetSingleMultipleOne, "###propAllowedType###",
							prop.AllowedTyp[j][0], -1)
						inverseSetSingleMultipleTwo += strings.Replace(
							template.InverseSetSingleMultipleTwo, "###propAllowedType###",
							prop.AllowedTyp[j][0], -1)
						inverseSetSingleMultipleThree += strings.Replace(
							template.InverseSetSingleMultipleThree, "###propAllowedType###",
							prop.AllowedTyp[j][0], -1)
					}
					temp = strings.Replace(strings.Replace(strings.Replace(
						template.ClassInverseSingleMultiple,
						"###inverseSetSingleMultipleOne###", inverseSetSingleMultipleOne, -1),
						"###inverseSetSingleMultipleTwo###", inverseSetSingleMultipleTwo, -1),
						"###inverseSetSingleMultipleThree###", inverseSetSingleMultipleThree, -1)
				}
			}
			ret += strings.Replace(strings.Replace(strings.Replace(strings.Replace(strings.Replace(
				strings.Replace(strings.Replace(temp, "###propCapital###", prop.Capital, -1),
					"###comment###", prop.Comment, -1),
				"###propName###", prop.Name, -1),
				"###propInverse###", prop.Inverse, -1),
				"###propLongName###", propName, -1),
				"###propBaseType###", prop.BaseTyp[0], -1),
				"###propType###", prop.Typ[0], -1)
		}
	}

	// Init
	initSwitchProps := ""
	parentInit := ""
	if isExactChild {
		parentInit += "\t\tres.s" + class.DirectParent[0] + ".propsInit(pred)\n"
	}
	for i := range class.Property {
		prop := class.Property[i]
		isParentProp := false
		if isExactChild {
			for j := range mod.Class[class.DirectParent[0]].Property {
				if class.Property[i].Name == mod.Class[class.DirectParent[0]].Property[j].Name {
					isParentProp = true
					break
				}
			}
		}
		if isParentProp {
			continue
		}
		temp := strings.Replace(template.InitSwitchProp, "###propIRI###", prop.IRI, -1)
		if prop.Inverse == "" {
			propName := generatePropertyName(prop)
			if prop.Typ[0] == "time.Time" || prop.Typ[0] == "int" || prop.Typ[0] == "string" ||
				prop.Typ[0] == "float64" || prop.Typ[0] == "bool" ||
				prop.Typ[0] == "time.Duration" || prop.Typ[0] == "interface{}" {
				initSwitchProps += strings.Replace(strings.Replace(temp,
					"###PropInit###", template.PropInitLiteralNonInverse, -1),
					"###propLongName###", propName, -1)
			} else {
				initSwitchProps += strings.Replace(strings.Replace(temp,
					"###PropInit###", template.PropInitClassNonInverse, -1),
					"###propLongName###", propName, -1)
			}
		} else {
			mult := ""
			if prop.Multi {
				mult = template.MultiplicityMultiple
			} else {
				mult = template.MultiplicitySingle
			}
			switch prop.Typ[0] {
			case "time.Time":
				switch prop.XSDTyp {
				case "http://www.w3.org/2001/XMLSchema#dateTime":
					temp = strings.Replace(temp, "###PropInit###", template.PropDateTime, -1)
				case "http://www.w3.org/2001/XMLSchema#date":
					temp = strings.Replace(temp, "###PropInit###", template.PropDate, -1)
				case "http://www.w3.org/2001/XMLSchema#dateTimeStamp":
					temp = strings.Replace(temp, "###PropInit###", template.PropDateTimeStamp, -1)
				case "http://www.w3.org/2001/XMLSchema#gYear":
					temp = strings.Replace(temp, "###PropInit###", template.PropGYear, -1)
				case "http://www.w3.org/2001/XMLSchema#gDay":
					temp = strings.Replace(temp, "###PropInit###", template.PropGDay, -1)
				case "http://www.w3.org/2001/XMLSchema#gYearMonth":
					temp = strings.Replace(temp, "###PropInit###", template.PropGYearMonth, -1)
				case "http://www.w3.org/2001/XMLSchema#gMonth":
					temp = strings.Replace(temp, "###PropInit###", template.PropGMonth, -1)
				}
			case "time.Duration":
				temp = strings.Replace(temp, "###PropInit###", template.PropDuration, -1)
			case "int":
				temp = strings.Replace(temp, "###PropInit###", template.PropInt, -1)
			case "float64":
				temp = strings.Replace(temp, "###PropInit###", template.PropFloat, -1)
			case "bool":
				temp = strings.Replace(temp, "###PropInit###", template.PropBool, -1)
			case "string":
				temp = strings.Replace(temp, "###PropInit###", template.PropString, -1)
			default:
				tempSp := strings.Split(prop.BaseTyp[0], ".")
				if prop.BaseTyp[0] == "owl.Thing" {
					temp = strings.Replace(temp, "###PropInit###", template.PropClassBaseThing, -1)
				} else if len(tempSp) > 1 {
					imName := strings.TrimPrefix(tempSp[0], "im")
					prop.BaseTyp[0] = tempSp[1]
					temp = strings.Replace(temp, "###PropInit###", template.PropClassImport, -1)
					temp = strings.Replace(temp, "###capImportName###", strings.Title(imName), -1)
				} else {
					temp = strings.Replace(temp, "###PropInit###", template.PropClassDefault, -1)
				}
			}
			initSwitchProps += strings.Replace(strings.Replace(strings.Replace(temp,
				"###Multiplicity###", mult, -1),
				"###propBaseType###", prop.BaseTyp[0], -1),
				"###propCapital###", prop.Capital, -1)
		}
	}
	if !equalParentProps {
		ret += strings.Replace(strings.Replace(template.ClassInit+template.PropsInit,
			"###initSwitchProps###", initSwitchProps, -1),
			"###parentPropInit###", parentInit, -1)
	}

	// To Graph
	toGraphProps := ""
	if isExactChild {
		toGraphProps += "\tres.s" + class.DirectParent[0] + ".propsToGraph(node, g)\n"
	}
	for i := range class.Property {
		isParentProp := false
		if isExactChild {
			for j := range mod.Class[class.DirectParent[0]].Property {
				if class.Property[i].Name == mod.Class[class.DirectParent[0]].Property[j].Name {
					isParentProp = true
					break
				}
			}
		}
		if !isParentProp {
			toGraphProps += strings.Replace(template.ToGraphProp, "###propLongName###",
				generatePropertyName(class.Property[i]), -1)
		}
	}
	if toGraphProps == "" {
		ret += strings.Replace(template.ClassToGraphNoProp, "###classIRI###", class.IRI, -1)
	} else {
		ret += strings.Replace(template.ClassToGraph,
			"###classIRI###", class.IRI, -1)
		if !equalParentProps {
			ret += strings.Replace(template.PropsToGraph,
				"###toGraphProps###", toGraphProps, -1)
		}
	}

	// String
	stringProps := ""
	if isExactChild {
		stringProps += "\tret += res.s" + class.DirectParent[0] + ".propsString()\n"
	}
	for i := range class.Property {
		isParentProp := false
		if isExactChild {
			for j := range mod.Class[class.DirectParent[0]].Property {
				if class.Property[i].Name == mod.Class[class.DirectParent[0]].Property[j].Name {
					isParentProp = true
					break
				}
			}
		}
		if !isParentProp {
			stringProps += strings.Replace(template.StringProp, "###propLongName###",
				generatePropertyName(class.Property[i]), -1)
		}
	}
	ret += template.ClassString

	if !equalParentProps {
		ret += strings.Replace(template.PropsString, "###stringProps###", stringProps, -1)
	}

	ret = strings.Replace(ret, "###className###", class.Name, -1)
	return
}

// getImportRecursive
func getImportRecursive(mod *owl.GoModel) (ret []*owl.GoModel) {
	for i := range mod.Import {
		ret = append(ret, mod.Import[i])
		ret = append(ret, getImportRecursive(mod.Import[i])...)
	}
	return
}

// getImportPath
func getImportPath(mod *owl.GoModel, pkg string) (ret string) {
	for i := range mod.Import {
		ret = getImportPath(mod.Import[i], pkg)
	}
	if mod.Name == pkg {
		ret = pkg
	} else if ret != "" {
		ret = mod.Name + "." + ret
	} else {
		ret = ""
	}
	return
}
