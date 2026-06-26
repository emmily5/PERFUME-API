-- Dados de exemplo para desenvolvimento (Sprint 2).
-- Aplique depois do schema. Idempotente o suficiente para rodar em banco limpo.

INSERT INTO marcas (nome, pais_origem) VALUES
    ('Dior', 'França'),
    ('Chanel', 'França'),
    ('YSL', 'França'),
    ('Giorgio Armani', 'Itália')
ON CONFLICT (nome) DO NOTHING;

INSERT INTO perfumes (marca_id, nome, preco, tamanho, genero, estoque)
SELECT m.id, v.nome, v.preco, v.tamanho, v.genero, v.estoque
FROM (VALUES
    ('Dior',           'Sauvage',      650.00, '100ml', 'masculino', 10),
    ('Chanel',         'Chanel N°5',   780.00, '50ml',  'feminino',   5),
    ('YSL',            'Black Opium',  520.00, '90ml',  'feminino',   8),
    ('Giorgio Armani', 'Acqua di Gio', 480.00, '100ml', 'masculino', 12)
) AS v(marca, nome, preco, tamanho, genero, estoque)
JOIN marcas m ON m.nome = v.marca;
