# GoWebServer

## À propos

Ce projet a été réalisé dans le cadre d'un test technique pour l'entreprise [Scalingo](https://scalingo.com/).
Le but de ce projet est de montrer les capacités à réaliser un utilitaire utilisant des des ressources externes de type API HTTP et de traiter les résultats de ces ressources en parallèle. Le sujet complet peut être trouvé [ici](https://github.com/iAmoric/GoWebServer/blob/master/TechnicalTests_Backend_FR.pdf)

## Installation

Ce projet n'est pas installable via la command 'go get'. Il suffit de télécharger l'archive du projet [ici](https://github.com/iAmoric/GoWebServer/archive/master.zip)

## Utilisation

### Serveur

Le serveur se lance avec la commande `go run apiserver.go`. Le port d'écoute est le `8080`
(ex: `localhost:8080`)

### Requêtes à l'API

Les requêtes à l'API GitHub utilisent l'authentification OAuth. C'est pourquoi il est nécessaire de configurer le token pour avoir accès à l'API. Il s'agit de la variable globale `token`.

La requête pour lister les dépôts GitHub est : `https://api.github.com/repositories{?since}`.

La requête pour lister les languages d'un dépot GitHub est : `https://api.github.com/repos/{login}/{repository}/languages`

La requête pour lister les dépôts GitHub avec un langage particulier est :
`https://api.github.com/search/repositories?q=language:{language}`

Les requêtes permettant de récupérer les langages utilisés par un certain dépôts GitHub sont faites en parallèle. Le traitement est limité à 10 `goroutines` en parallèle.

### Affichage

La page d'accueil (`/`) affiche la liste des 100 dépôts GitHub, ainsi que la liste des langages utilisés et leur dépôts associés.

Un formulaire de recherche permet de rechercher les dépôts GitHub utilisant un langage particulier.

La page de recherche (`/search?language={language}`) affiche la liste résultant de la recherche ci-dessus.

Dans tous les cas, le nombre de lignes de code est affiché.

### Explications techniques

Le dossier `src/` contient un sous-dossier `tests/`. Ce dossier ne contient pas des tests pour le code, mais des essais (sur la création d'un serveur web ou l'utilisation de l'API GitHub). En effet, ce projet est mon premier projet utilisant le langage Go, et il m'a été nécessaire d'effectuer ces tests afin de comprendre le fonctionnement du langage.


## Auteur

* [**Lucas Pierrat**](https://github.com/iAmoric) - [contact](mailto:pierratlucas@gmail.com)

## License

This project is licensed under the MIT License - see the [LICENSE.md](https://github.com/iAmoric/GoWebServer/blob/master/LICENSE) file for details
