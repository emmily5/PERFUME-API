# 🎬 DEMO — roteiro de comandos para a gravação

Comandos prontos para colar durante o vídeo. Siga na ordem.
A base já está limpa (4 marcas + 4 perfumes de exemplo, sem duplicatas).

---

## Passo 0 — Subir o servidor (Terminal 1, deixe aberto)

```bash
cd /Users/emillymillermoreira/Desktop/Desktop/PERFUME-API
export DATABASE_URL="postgres://postgres:root@localhost:5432/perfumaria?sslmode=disable"
export JWT_SECRET="dev-secret-local-0123456789"
make run
```

Espere aparecer: `🌸 Servidor rodando em http://localhost:8080`.
Abra um **Terminal 2** para os comandos abaixo.

---

## 1 — Rota protegida SEM token → 401 (controle de acesso · OWASP A01)

```bash
curl -i -X POST http://localhost:8080/perfumes \
  -H "Content-Type: application/json" \
  -d '{"marca_id":1,"nome":"Hack","preco":1}'
```
➡️ `HTTP/1.1 401 Unauthorized`

---

## 2 — Registrar usuário e fazer login (JWT)

```bash
curl -s -X POST http://localhost:8080/auth/registrar \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@loja.com","senha":"senhaSegura123"}' | jq
```

```bash
LOGIN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@loja.com","senha":"senhaSegura123"}')
echo "$LOGIN" | jq
ACCESS=$(echo "$LOGIN"  | jq -r .access_token)
REFRESH=$(echo "$LOGIN" | jq -r .refresh_token)
```
➡️ Mostra o `access_token` (JWT) e o `refresh_token`.

---

## 3 — Rota protegida COM token → 201 (autorizado)

```bash
curl -s -X POST http://localhost:8080/marcas \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS" \
  -d '{"nome":"Paco Rabanne","pais_origem":"Espanha"}' | jq
```

```bash
curl -s -X POST http://localhost:8080/perfumes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS" \
  -d '{"marca_id":1,"nome":"1 Million","preco":390,"tamanho":"100ml","genero":"masculino","estoque":6}' | jq
```
➡️ `201 Created` nas duas.

---

## 4 — Relacionamento 1:N (dados aninhados)

```bash
curl -s http://localhost:8080/marcas/1 | jq
```
➡️ A marca **Dior** vem com o array `perfumes` aninhado dentro.

---

## 5 — Refresh com rotação + bloqueio de reuso

```bash
# rotaciona: devolve um refresh_token NOVO
curl -s -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH\"}" | jq
```

```bash
# reusar o refresh ANTIGO agora → 401
curl -i -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH\"}"
```
➡️ Primeiro `200` com token novo; segundo `401` (reuso detectado).

---

## 6 — Rate limiting no login (OWASP A07)

```bash
for i in $(seq 1 7); do
  curl -s -o /dev/null -w "tentativa $i -> %{http_code}\n" \
    -X POST http://localhost:8080/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"x@x.com","senha":"errada"}'
done
```
➡️ Começa em `401` e, ao passar de 5/min, vira `429 Too Many Requests`.

---

## 7 — Cabeçalhos de segurança (OWASP A05)

```bash
curl -s -D - -o /dev/null http://localhost:8080/ \
  | grep -iE "x-content-type|x-frame|content-security|referrer"
```

---

## 8 — GraphQL (bônus)

Terminal:
```bash
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{ marcas { nome pais_origem perfumes { nome preco } } }"}' | jq
```

Navegador (playground visual): **http://localhost:8080/graphql**
```graphql
{ marcas { nome pais_origem perfumes { nome preco } } }
```

---

## 9 — Testes automatizados + CI

```bash
go test ./... -v
```
➡️ 15 × `--- PASS`. Depois mostre o **GitHub Actions** verde (CI sobe PostgreSQL e roda as migrations).

---

## Encerrar
No Terminal 1: `Ctrl+C`.

> Os comandos usam `jq` para formatar o JSON. Se não tiver: `brew install jq`
> (ou remova o ` | jq` do fim de cada linha).
