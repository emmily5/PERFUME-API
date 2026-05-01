package models

// Perfume representa um perfume na loja
type Perfume struct {
	ID       int     `json:"id"`
	Nome     string  `json:"nome"`
	Marca    string  `json:"marca"`
	Preco    float64 `json:"preco"`
	Tamanho  string  `json:"tamanho"`  // ex: "50ml", "100ml"
	Genero   string  `json:"genero"`   // "masculino", "feminino", "unissex"
	Estoque  int     `json:"estoque"`
}
