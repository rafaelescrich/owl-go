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

package owl

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"git-ce.rwth-aachen.de/acs/private/research/agent/owl2go.git/pkg/rdf"
)

// ExtractOntologyLink extracts all classes, properties, individuals and imports
func ExtractOntologyLink(link string) (on Ontology, err error) {
	var resp *http.Response
	resp, err = requestOntology(link)
	if err != nil {
		return
	}
	on, err = ExtractOntology(resp.Body)
	resp.Body.Close()
	return
}

// ExtractOntology extracts all classes, properties, individuals and imports
func ExtractOntology(input io.Reader) (on Ontology, err error) {
	iri := ""
	description := ""
	on.Class = make(map[string]*Class)
	on.Property = make(map[string]*Property)
	on.Individual = make(map[string]*Individual)
	on.Imports = make(map[string][]string)
	on.Description = make(map[string]string)
	on.Content = make(map[string][]byte)

	var g rdf.Graph
	g, iri, description, _, err = parseOntology(input)
	if err != nil {
		return
	}

	on.graph = &g
	on.Description[iri] = description
	on.Imports[iri] = []string{}

	err = on.parseImports(on.graph)
	if err != nil {
		return
	}

	on.Class, err = extractClasses(on.graph)
	if err != nil {
		return
	}

	on.Property, err = extractProperties(on.graph)
	if err != nil {
		return
	}

	on.Individual, err = extractIndividuals(on.graph, on.Class)
	if err != nil {
		return
	}

	err = on.postProcessProperties()
	if err != nil {
		return
	}
	err = on.postProcessClasses()
	if err != nil {
		return
	}
	err = on.addPropertyDomain()
	if err != nil {
		return
	}

	return
}

// parseOntology parses the specified ontology
func parseOntology(input io.Reader) (g rdf.Graph, iri string, description string, content []byte,
	err error) {
	fmt.Println("Read TTL input")
	g, err = readTTL(input)
	if err != nil {
		err = errors.New("Cannot parse ontology: " + err.Error())
		return
	}

	// get ontology iri
	for i := range g.Edges {
		if g.Edges[i].Pred.String() == "http://www.w3.org/1999/02/22-rdf-syntax-ns#type" &&
			g.Edges[i].Object.Term.String() == "http://www.w3.org/2002/07/owl#Ontology" {
			iri = g.Edges[i].Subject.Term.String()
		} else if g.Edges[i].Pred.String() == "http://purl.org/dc/terms/description" {
			description = g.Edges[i].Object.Term.String()
		}
	}
	return
}

// readTTL reads a ttl file and returns a graph
func readTTL(input io.Reader) (g rdf.Graph, err error) {
	var triples []rdf.Triple
	triples, err = rdf.DecodeTTL(input)
	if err != nil {
		return
	}

	g, err = rdf.NewGraph(triples)
	return
}

// parseImports parses all imports and adds imports to ontologies
func (on *Ontology) parseImports(gIn *rdf.Graph) (err error) {
	var gTemp rdf.Graph
	gTemp.Nodes = make(map[string]*rdf.Node)
	hasImport := false
	for i := range gIn.Edges {
		if gIn.Edges[i].Pred.String() == "http://www.w3.org/2002/07/owl#imports" {
			hasImport = true
			iri := gIn.Edges[i].Subject.Term.String()
			impIRI := gIn.Edges[i].Object.Term.String()

			on.Imports[iri] = append(on.Imports[iri], impIRI)

			var resp *http.Response
			resp, err = requestOntology(gIn.Edges[i].Object.Term.String())
			if err != nil {
				return
			}
			fmt.Println("Parse Imported Ontology " + gIn.Edges[i].Object.Term.String())
			var g rdf.Graph
			var desc string
			g, impIRI, desc, _, err = parseOntology(resp.Body)
			if err != nil {
				err = errors.New("Error parsing import " + gIn.Edges[i].Object.Term.String())
				return
			}
			resp.Body.Close()

			on.Description[impIRI] = desc
			on.Imports[impIRI] = []string{}

			gTemp.Merge(&g)
		}
	}
	if hasImport {
		on.parseImports(&gTemp)
		on.graph.Merge(&gTemp)
	}
	return
}

// getComment returns a comment if it exists (rdf:comment)
func getComment(node *rdf.Node) (ret string) {
	// find comment
	for j := range node.Edge {
		if node.Edge[j].Pred.String() == "http://www.w3.org/2000/01/rdf-schema#comment" {
			regex := regexp.MustCompile(`\r?\n`)
			ret = regex.ReplaceAllString(node.Edge[j].Object.Term.String(), " ")
			regex = regexp.MustCompile(`\n`)
			ret = regex.ReplaceAllString(ret, " ")
			// ret = strings.Replace(node.Predicates[j].Object.Name, "\n", " ", -1)

			break
		}
	}
	if ret == "" {
		ret = "no comment"
	}
	return
}

// getUnionValues returns all values of a union (rdfs:first and rdfs:rest)
func getUnionValues(node *rdf.Node) (ret []*rdf.Node) {
	for i := range node.Edge {
		if node.Edge[i].Pred.String() == "http://www.w3.org/1999/02/22-rdf-syntax-ns#first" {
			ret = append(ret, node.Edge[i].Object)
		} else if node.Edge[i].Pred.String() == "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest" &&
			node.Edge[i].Object.Term.String() != "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil" {
			ret = append(ret, getUnionValues(node.Edge[i].Object)...)
		}
	}
	return
}

// requestOntology requests a ontology via http
func requestOntology(path string) (resp *http.Response, err error) {
	client := &http.Client{
		Timeout: time.Second * 2,
	}
	resp, err = client.Get(path)
	return
}
