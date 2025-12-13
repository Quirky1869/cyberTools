# cyberTools

![](./_images/cyberTools.png)  

![Static Badge](https://img.shields.io/badge/TUI-cyberTools-cyan?style=plastic)
![Static Badge](https://img.shields.io/badge/License-MIT-500aff?style=plastic)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=plastic&logo=go)](https://golang.org/)
[![Go Report Card](https://goreportcard.com/badge/github.com/Quirky1869/cyberTools?color=00F0FF)](https://goreportcard.com/report/github.com/Quirky1869/cyberTools)
[![Latest Version](https://img.shields.io/github/v/release/Quirky1869/cyberTools?color=FF2A6D)](https://github.com/Quirky1869/cyberTools/releases)
[![GitHub Actions](https://github.com/Quirky1869/cyberTools/actions/workflows/build-and-tests.yml/badge.svg)](https://github.com/Quirky1869/cyberTools/actions/workflows/build-and-tests.yml)
[![GitHub Actions Release](https://github.com/Quirky1869/cyberTools/actions/workflows/build-test-release.yml/badge.svg)](https://github.com/Quirky1869/cyberTools/actions/workflows/build-test-release.yml)

## Résumé
<p align="center">  
<a href="https://golang.org" target="_blank" rel="noreferrer">  
  <img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/go/go-original.svg" alt="go" width="40" height="40"/>  
</a>  
</p>  

CyberTools est une interface utilisateur textuelle (TUI) immersive et modulaire développée en Go. Elle agit comme un hub centralisé ("Tools Box") permettant d'organiser, de naviguer et d'exécuter rapidement divers scripts et outils de cybersécurité et d'administration système.  

Construite avec l'écosystème Charm (Bubble Tea, Lipgloss), cette TUI met l'accent sur l'ergonomie :

- Navigation Fluide : Interface entièrement pilotable au clavier (flèches, tab et des touches Vim h/j/k/l).  
- Organisation par Onglets : Système de catégories ([BDD], [Utilitaire], [DATA], etc.) pour classer proprement les outils.  
- Design Cyberpunk : Une identité visuelle forte avec titre ASCII Art, effets de focus lumineux et bordures stylisées.  
- Thèmes Dynamiques : Possibilité de changer l'ambiance visuelle à la volée (ex: thème Neon vs Cyberpunk) sans redémarrer le programme (t).  
- Aide Contextuelle : Barre d'aide interactive qui s'adapte selon la section active.  

## Fonctionnalités

La TUI `cyberTools` faite en [Go](https://go.dev) permet de lancés plusieurs outils utilitaires eux aussi fait en Go  

## Structure du projet  

```bash
.
├── cmd
│   └── app
│       └── main.go # Point d'entrée du programme  
├── tools
│   ├── definition.go # Définition de la liste des outils disponibles (catégories et noms)
│   ├── logv
│   │   └── model.go  # Visualiseur de Logs avec coloration syntaxique
│   ├── sqltui
│   │   └── model.go  # Explorateur de bases de données SQLite en TUI
│   ├── structViewer
│   │   └── model.go  # Explorateur YAML/JSON en vue arborescente
│   └── aed
│       └── model.go  # Analyseur d'Espace Disque
├── ui
│   ├── keys.go   # Définition et gestion des combinaisons de touches globales
│   ├── root.go   # Logique principale du menu. Gère les différents états de l'application
│   └── styles.go # Définition des styles Lipgloss
├── README.md  
├── _images     # Dossier des assets
├── go.mod
├── go.sum
└── build.sh    # Script Bash pour la compilation du binaire cyberTools.
```

## Lancement de la TUI

>[!TIP]  
Vous pouvez lancer la TUI de 3 manières différentes  
### Via les releases
Vous pouvez exécuter le binaire en téléchargeant les [releases](https://github.com/Quirky1869/cyberTools/releases)  
### En buildant le projet
>[!CAUTION]
Go doit être [installé](https://go.dev/doc/install) sur votre PC  
Après avoir fait un `git clone https://github.com/Quirky1869/cyberTools.git` et `cd cyberTools`  
Vous pouvez compiler le projet en exécutant le fichier `./build.sh` puis lancer le projet compiler via`./bin/cyberTools` (Go doit être installé sur votre PC)  
### En exécutant directement le projet
>[!CAUTION]
Go doit être [installé](https://go.dev/doc/install) sur votre PC  
Vous pouvez aussi lancer la commande `go run cmd/app/main.go`    

![gif](_images/gif/cyberTools.gif)

## Releases

Les [releases](https://github.com/Quirky1869/cyberTools/releases) sont disponibles [ici](https://github.com/Quirky1869/cyberTools/releases)  

## Technologies utilisées

| Librairie                                                    | Utilisation                          |
| ------------------------------------------------------------ | ------------------------------------ |
| [Bubble Tea](https://github.com/charmbracelet/bubbletea)     | Gestion de l’interface TUI           |
| [Lipgloss](https://github.com/charmbracelet/lipgloss)        | Stylisation du texte et des bordures |
| ![Go](https://img.shields.io/badge/Go-%2300ADD8.svg?style=flat&logo=go&logoColor=white) [Golang](https://go.dev)   | Scripts d’installation des paquets   |


## Auteur

Projet développé par Quirky  
<a href="https://github.com/Quirky1869" target="_blank">  
  <img src="./_images/white-github.png" alt="GitHub" width="30" height="30" style="vertical-align:middle;"> GitHub  
</a>  

## Licence

Ce projet est distribué sous licence MIT  
