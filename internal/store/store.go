package store

import (
	"errors"
	"sync"

	"github.com/seunome/perfume-api/internal/models"
)

// ErrNaoEncontrado é retornado quando o perfume não existe
var ErrNaoEncontrado = errors.New("perfume não encontrado")

// PerfumeStore guarda os perfumes em memória (sem banco de dados)
type PerfumeStore struct {
	mu       sync.RWMutex
	perfumes map[int]models.Perfume
	nextID   int
}

// New cria um novo store com dados iniciais de exemplo
func New() *PerfumeStore {
	s := &PerfumeStore{
		perfumes: make(map[int]models.Perfume),
		nextID:   1,
	}

	// Dados iniciais para demonstração
	iniciais := []models.Perfume{
		{Nome: "Sauvage", Marca: "Dior", Preco: 650.00, Tamanho: "100ml", Genero: "masculino", Estoque: 10},
		{Nome: "Chanel N°5", Marca: "Chanel", Preco: 780.00, Tamanho: "50ml", Genero: "feminino", Estoque: 5},
		{Nome: "Black Opium", Marca: "YSL", Preco: 520.00, Tamanho: "90ml", Genero: "feminino", Estoque: 8},
		{Nome: "Acqua di Gio", Marca: "Giorgio Armani", Preco: 480.00, Tamanho: "100ml", Genero: "masculino", Estoque: 12},
	}

	for _, p := range iniciais {
		p.ID = s.nextID
		s.perfumes[s.nextID] = p
		s.nextID++
	}

	return s
}

// Listar retorna todos os perfumes
func (s *PerfumeStore) Listar() []models.Perfume {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lista := make([]models.Perfume, 0, len(s.perfumes))
	for _, p := range s.perfumes {
		lista = append(lista, p)
	}
	return lista
}

// BuscarPorID retorna um perfume pelo ID
func (s *PerfumeStore) BuscarPorID(id int) (models.Perfume, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.perfumes[id]
	if !ok {
		return models.Perfume{}, ErrNaoEncontrado
	}
	return p, nil
}

// Criar adiciona um novo perfume e retorna ele com o ID gerado
func (s *PerfumeStore) Criar(p models.Perfume) models.Perfume {
	s.mu.Lock()
	defer s.mu.Unlock()

	p.ID = s.nextID
	s.perfumes[s.nextID] = p
	s.nextID++
	return p
}

// Atualizar substitui os dados de um perfume existente
func (s *PerfumeStore) Atualizar(id int, p models.Perfume) (models.Perfume, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.perfumes[id]; !ok {
		return models.Perfume{}, ErrNaoEncontrado
	}
	p.ID = id
	s.perfumes[id] = p
	return p, nil
}

// Deletar remove um perfume pelo ID
func (s *PerfumeStore) Deletar(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.perfumes[id]; !ok {
		return ErrNaoEncontrado
	}
	delete(s.perfumes, id)
	return nil
}
