package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"strings"
)

type targetMethod struct {
	name         string
	requestType  string
	responseType string
}

type target struct {
	serviceName    string
	clientIntfName string
	methods        []targetMethod

	fullServiceName string

	importAlias string
	importPath  string
}

func parseMethod(field *ast.Field) targetMethod {
	typedMethod := field.Type.(*ast.FuncType)
	methodName := field.Names[0].Name

	defer func() {
		err := recover()
		if err != nil {
			log.Fatalf(
				"failed to parse interface method %s, is the input file a connectrpc generation?\nerr: %v\n",
				methodName,
				err,
			)
		}
	}()

	req := typedMethod.Params.List[1].Type.(*ast.StarExpr).X.(*ast.IndexExpr).Index.(*ast.SelectorExpr)
	res := typedMethod.Results.List[0].Type.(*ast.StarExpr).X.(*ast.IndexExpr).Index.(*ast.SelectorExpr)

	return targetMethod{
		name:         methodName,
		requestType:  fmt.Sprintf("%s.%s", req.X.(*ast.Ident).Name, req.Sel.Name),
		responseType: fmt.Sprintf("%s.%s", res.X.(*ast.Ident).Name, res.Sel.Name),
	}
}

func parseInterface(spec *ast.TypeSpec) []targetMethod {
	typedType := spec.Type.(*ast.InterfaceType)

	methods := make([]targetMethod, len(typedType.Methods.List))
	for i, field := range typedType.Methods.List {
		methods[i] = parseMethod(field)
	}
	return methods
}

func parseTargets(file *ast.File) []*target {
	var targetList []*target

	for _, decl := range file.Decls {
		switch typedDecl := decl.(type) {
		case *ast.GenDecl:
			if typedDecl.Tok != token.TYPE {
				break
			}
			for _, spec := range typedDecl.Specs {
				typedSpec := spec.(*ast.TypeSpec)
				name := typedSpec.Name.String()

				firstChar := name[0]
				if firstChar < 'A' || firstChar > 'Z' || !strings.HasSuffix(name, "Client") {
					continue
				}
				targetList = append(targetList, &target{
					serviceName:    name[:len(name)-6],
					clientIntfName: name,
					methods:        parseInterface(typedSpec),
				})
			}
		}
	}

	for _, decl := range file.Decls {
		switch typedDecl := decl.(type) {
		case *ast.GenDecl:
			if typedDecl.Tok != token.CONST {
				break
			}
			for _, spec := range typedDecl.Specs {
				typedSpec := spec.(*ast.ValueSpec)

				for i, ident := range typedSpec.Names {
					for _, target := range targetList {
						if ident.Name == target.serviceName+"Name" {
							value := typedSpec.Values[i].(*ast.BasicLit)
							// remove quotes from string
							target.fullServiceName = value.Value[1 : len(value.Value)-1]
							break
						}
					}
				}
			}
		}
	}

	for _, target := range targetList {
		segments := strings.Split(target.fullServiceName, ".")
		mustContain := strings.Join(segments[:len(segments)-1], "/")
		for _, imp := range file.Imports {
			if strings.Contains(imp.Path.Value, mustContain) {
				target.importPath = imp.Path.Value
				target.importAlias = imp.Name.Name
				break
			}
		}
	}

	return targetList
}
