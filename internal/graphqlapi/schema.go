// Package graphqlapi expõe um endpoint GraphQL read-only (bônus da Sprint 3)
// que entrega marcas com seus perfumes aninhados (relacionamento 1:N).
package graphqlapi

import (
	"net/http"

	"github.com/graphql-go/graphql"
	gqlhandler "github.com/graphql-go/handler"
	"github.com/seunome/perfume-api/internal/db"
)

// NewHandler monta o schema GraphQL e devolve um handler HTTP com playground.
func NewHandler(q *db.Queries) (http.Handler, error) {
	perfumeType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Perfume",
		Fields: graphql.Fields{
			"id":       &graphql.Field{Type: graphql.Int, Resolve: campo(func(p db.Perfume) any { return p.ID })},
			"nome":     &graphql.Field{Type: graphql.String, Resolve: campo(func(p db.Perfume) any { return p.Nome })},
			"preco":    &graphql.Field{Type: graphql.Float, Resolve: campo(func(p db.Perfume) any { return p.Preco })},
			"tamanho":  &graphql.Field{Type: graphql.String, Resolve: campo(func(p db.Perfume) any { return p.Tamanho })},
			"genero":   &graphql.Field{Type: graphql.String, Resolve: campo(func(p db.Perfume) any { return p.Genero })},
			"estoque":  &graphql.Field{Type: graphql.Int, Resolve: campo(func(p db.Perfume) any { return p.Estoque })},
			"marca_id": &graphql.Field{Type: graphql.Int, Resolve: campo(func(p db.Perfume) any { return p.MarcaID })},
		},
	})

	marcaType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Marca",
		Fields: graphql.Fields{
			"id":          &graphql.Field{Type: graphql.Int, Resolve: campoMarca(func(m db.Marca) any { return m.ID })},
			"nome":        &graphql.Field{Type: graphql.String, Resolve: campoMarca(func(m db.Marca) any { return m.Nome })},
			"pais_origem": &graphql.Field{Type: graphql.String, Resolve: campoMarca(func(m db.Marca) any { return m.PaisOrigem })},
			// Campo aninhado 1:N: resolve os perfumes da marca sob demanda.
			"perfumes": &graphql.Field{
				Type: graphql.NewList(perfumeType),
				Resolve: func(p graphql.ResolveParams) (any, error) {
					m := p.Source.(db.Marca)
					return q.ListarPerfumesPorMarca(p.Context, m.ID)
				},
			},
		},
	})

	query := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"marcas": &graphql.Field{
				Type: graphql.NewList(marcaType),
				Resolve: func(p graphql.ResolveParams) (any, error) {
					return q.ListarMarcas(p.Context)
				},
			},
			"marca": &graphql.Field{
				Type: marcaType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					id := int64(p.Args["id"].(int))
					return q.BuscarMarca(p.Context, id)
				},
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{Query: query})
	if err != nil {
		return nil, err
	}

	return gqlhandler.New(&gqlhandler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true, // playground em GET /graphql
	}), nil
}

// campo/campoMarca constroem resolvers de campos escalares a partir do Source.
func campo(get func(db.Perfume) any) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (any, error) {
		return get(p.Source.(db.Perfume)), nil
	}
}

func campoMarca(get func(db.Marca) any) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (any, error) {
		return get(p.Source.(db.Marca)), nil
	}
}
