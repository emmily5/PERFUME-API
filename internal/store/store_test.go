package store_test

import (
	"testing"

	"github.com/seunome/perfume-api/internal/models"
	"github.com/seunome/perfume-api/internal/store"
)

func TestCriarEBuscarPerfume(t *testing.T) {
	s := store.New()

	novo := models.Perfume{
		Nome:    "Bleu de Chanel",
		Marca:   "Chanel",
		Preco:   700.00,
		Tamanho: "100ml",
		Genero:  "masculino",
		Estoque: 3,
	}

	criado := s.Criar(novo)

	if criado.ID == 0 {
		t.Fatal("esperava ID gerado, mas ficou 0")
	}
	if criado.Nome != novo.Nome {
		t.Errorf("nome esperado %q, recebeu %q", novo.Nome, criado.Nome)
	}

	// Busca pelo ID gerado
	encontrado, err := s.BuscarPorID(criado.ID)
	if err != nil {
		t.Fatalf("erro ao buscar perfume: %v", err)
	}
	if encontrado.Marca != novo.Marca {
		t.Errorf("marca esperada %q, recebeu %q", novo.Marca, encontrado.Marca)
	}
}

func TestListarPerfumes(t *testing.T) {
	s := store.New()

	// O store já começa com 4 perfumes de exemplo
	lista := s.Listar()
	if len(lista) < 4 {
		t.Errorf("esperava pelo menos 4 perfumes, recebeu %d", len(lista))
	}
}

func TestAtualizarPerfume(t *testing.T) {
	s := store.New()

	novo := s.Criar(models.Perfume{Nome: "Teste", Marca: "MarcaTeste", Preco: 100})

	atualizado, err := s.Atualizar(novo.ID, models.Perfume{
		Nome:  "Teste Atualizado",
		Marca: "MarcaNova",
		Preco: 200,
	})
	if err != nil {
		t.Fatalf("erro ao atualizar: %v", err)
	}
	if atualizado.Nome != "Teste Atualizado" {
		t.Errorf("esperava nome atualizado, recebeu %q", atualizado.Nome)
	}
	if atualizado.ID != novo.ID {
		t.Errorf("ID mudou após atualização: era %d, ficou %d", novo.ID, atualizado.ID)
	}
}

func TestDeletarPerfume(t *testing.T) {
	s := store.New()

	novo := s.Criar(models.Perfume{Nome: "ParaDeletar", Marca: "X", Preco: 50})

	if err := s.Deletar(novo.ID); err != nil {
		t.Fatalf("erro ao deletar: %v", err)
	}

	_, err := s.BuscarPorID(novo.ID)
	if err == nil {
		t.Error("perfume deveria ter sido deletado, mas ainda foi encontrado")
	}
}

func TestBuscarIDInexistente(t *testing.T) {
	s := store.New()

	_, err := s.BuscarPorID(9999)
	if err == nil {
		t.Error("esperava erro ao buscar ID inexistente")
	}
}

func TestDeletarIDInexistente(t *testing.T) {
	s := store.New()

	err := s.Deletar(9999)
	if err == nil {
		t.Error("esperava erro ao deletar ID inexistente")
	}
}
