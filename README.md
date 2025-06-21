# Fantastic coffee (decaffeinated)

Questo repository contiene la struttura base per il progetto d'esame di [Web and Software Architecture](http://gamificationlab.uniroma1.it/en/wasa/).
È stato descritto a lezione.

"Fantastic coffee (decaffeinated)" è una versione semplificata per il corso WASA, non adatta a un ambiente di produzione.
La versione completa si trova nel repository "Fantastic Coffee".

## Struttura del progetto

* `cmd/` contiene tutti gli eseguibili; i programmi Go qui devono solo fare "cose da eseguibile", come leggere opzioni da CLI/env, ecc.
	* `cmd/healthcheck` è un esempio di demone per controllare la salute dei server; utile quando l'hypervisor non fornisce probe HTTP readiness/liveness (es. Docker engine)
	* `cmd/webapi` contiene un esempio di demo server API web
* `demo/` contiene un file di configurazione demo
* `doc/` contiene la documentazione (di solito, per le API, un file OpenAPI)
* `service/` contiene tutti i package per le funzionalità specifiche del progetto
	* `service/api` contiene un esempio di server API
	* `service/globaltime` contiene un package wrapper per `time.Time` (utile nei test unitari)
* `vendor/` è gestita da Go e contiene una copia di tutte le dipendenze
* `webui/` è un esempio di frontend web in Vue.js; include:
	* Bootstrap JavaScript framework
	* una versione personalizzata del template "Bootstrap dashboard"
	* feather icons come SVG
	* codice Go per l'embedding in release

Altri file del progetto includono:
* `open-node.sh` avvia un nuovo container (temporaneo) usando l'immagine `node:20` per uno sviluppo frontend sicuro (non vuoi usare node nel tuo sistema, vero?).

## Go vendoring

Questo progetto usa [Go Vendoring](https://go.dev/ref/mod#vendoring). Devi usare `go mod vendor` dopo aver cambiato una dipendenza (`go get` o `go mod tidy`) e aggiungere tutti i file nella cartella `vendor/` al commit.

Per maggiori informazioni sul vendoring:

* https://go.dev/ref/mod#vendoring
* https://www.ardanlabs.com/blog/2020/04/modules-06-vendoring.html

## Node/YARN vendoring

Questo repository usa `yarn` e una tecnica di vendoring che sfrutta lo ["Offline mirror"](https://yarnpkg.com/features/caching). Come per il vendoring Go, le dipendenze sono dentro il repository.

Devi fare commit dei file dentro la cartella `.yarn`.

## Come impostare un nuovo progetto da questo template

Devi:

* Cambiare il path del modulo Go in `go.mod`, `go.sum` e nei file `*.go` nel progetto
* Riscrivere la documentazione API in `doc/api.yaml`
* Se non serve il frontend web, rimuovi `webui` e `cmd/webapi/register-webui.go`
* Aggiorna il commento top/package dentro `cmd/webapi/main.go` per riflettere l'uso reale del progetto
* Aggiorna il codice nella funzione `run()` (`cmd/webapi/main.go`) per connetterti a database o risorse esterne
* Scrivi il codice API dentro `service/api` e crea altri package dentro `service/` (o sottocartelle)

## Come compilare

Se non usi la WebUI, o non vuoi includerla nell'eseguibile finale:

```shell
go build ./cmd/webapi/
```

Se usi la WebUI e vuoi includerla nell'eseguibile finale:

```shell
./open-node.sh
# (qui sei dentro il container)
yarn run build-embed
exit
# (fuori dal container)
go build -tags webui ./cmd/webapi/
```

## Come eseguire (in modalità sviluppo)

Puoi avviare solo il backend usando:

```shell
go run ./cmd/webapi/
```

Se vuoi avviare la WebUI, apri una nuova tab e lancia:

```shell
./open-node.sh
# (qui sei dentro il container)
yarn run dev
```

## Come compilare per la produzione / consegna

```shell
./open-node.sh
# (qui sei dentro il container)
yarn run build-prod
```

Per gli studenti di "Web and Software Architecture": prima di fare commit e push per la valutazione, leggi la sezione sotto chiamata "My build works when I use `yarn run dev`, however there is a Javascript crash in production/grading"

## Problemi noti

### My build works when I use `yarn run dev`, however there is a Javascript crash in production/grading

Alcuni errori nel codice non vengono mostrati in modalità sviluppo `vite`. Per vedere il codice che sarà usato in produzione/valutazione, usa questi comandi:

```shell
./open-node.sh
# (qui sei dentro il container)
yarn run build-prod
yarn run preview
```

# WASAText
Progetto esame 2025 sessione estiva

# parte backend
go run ./cmd/webapi/ 

# parte frontend
docker run -it --rm -v "$(pwd):/src" -u "$(id -u):$(id -g)" --network host --workdir /src/webui node:20 /bin/bash

# run in dev mode
yarn run dev
# non modificabile esecuzione finale 
yarn run build-prod

# sul frontend per chiudere
CTRL + C exit

# sul backend per chiudere
CTRL + C

# Docker deve essere aperto
# Docker build
docker build -f Dockerfile.backend -t wasa-backend .
docker build -f Dockerfile.frontend -t wasa-frontend .

# Docker run
docker run -p 3000:3000 wasa-backend
docker run -p 8080:80 wasa-frontend

