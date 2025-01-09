# User Management

* This is a dummy school login dashboard built with HTMX, Golang and TEMPL Golang templates.
This is also built with Docker (separately for running the database locally), Postgres, TailwindCSS and FrankenUI.

![Video](./assets/usermgmt.mp4)

## Table of Contents
- [User Management](#user-management)
  - [Table of Contents](#table-of-contents)
  - [Getting started](#getting-started)
  - [Running dev environment](#running-dev-environment)
  - [Directory structure](#directory-structure)
    - [Views](#views)
    - [Pkg](#pkg)
  - [Versioning](#versioning)

## Getting started 

Ensure the following dev-dependencies are installed:

1. Ensure Go version 1.22.3 or later is installed. You can find installation/upgrade instructions [here](https://go.dev/doc/install)
2. Install [templ](https://templ.guide/quick-start/installation), it generates the typed go templates, which is what the UI is. Make sure templ version matches the one installed in go.mod
3. Install [tailwindcss](https://tailwindcss.com/docs/installation), **NOTE**: this step can be skipped by creating a new make file instruction to use `npx` instead.
4. Install [air](https://github.com/air-verse/air?tab=readme-ov-file#installation), this ensures hot reloading of `.go` file changes, useful when running dev env. **NOTE**: Air is already configured (check `./air.toml`). You only need to install air


## Running dev environment

1. Run `npm install` for the latest dependencies
2. Run `templ generate`, this will genertae the templ files to Go files
3. Run `air`, this will run all the project dependencies in watch mode
4. You will need to set up a database in Postgres for this to work and have Docker and the database ports aligned. Look at the pkg/Model for the scheme. NOTE - Schema will be added soon.

## Directory structure 

### Views 

The UI is built with [franken-ui](https://franken-ui.dev/docs/introduction), which is an HTML only UI kit, perfect for this use case.
It uses tailwind-css to style elements. There are a bunch of pre-defined tailwind definitions which begin with `uk-`, so look for these 

For making ajax requests to the server [htmx](https://htmx.org/) is used.

This is where all the templ files reside. This package is further divided into components, <models>, and layouts.

```
view/
    layouts/ : Contains layout definitions and other dependencies to make the HTML look and run the way it should. Fonts, css, js, htmx, etc.
    <models>/ : Create a folder for model pages, an example would be cases, which would have an index page, edit and delete page 
    components/ : Any reusable piece of HTML, especially ones that have js or css rules that might need to be consistent
```

### Pkg 

All services that make the server function are defined here. Any new ones should be defined here as and when needed.

```
pkg/
    models/ : all the models which will be shared between UI and handlers are defined here, makes it easy to seemlessly return HTML in response to requests
    handlers/ : All route handler and their definitions. Each model has their own handler file, and defines all CRUD and types here
    repository/ : All database handlers and their definitions.
    middleware/ : Self explanatory
```

## Versioning

There is an `internal/version` package that handles the current version. The `Version` variable is updated during build time in production. 
In a dev setting, it will use the default value of `dev`.
