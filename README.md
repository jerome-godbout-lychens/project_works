# Project Works

## Badge

[![standard-readme compliant](https://img.shields.io/badge/readme%20style-standard-brightgreen.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)
[![Static Badge](https://img.shields.io/badge/Run-Docker-blue)
](https://www.docker.com/)
[![Static Badge](https://img.shields.io/badge/cli-Just-blue)](https://github.com/casey/just)

## Background

This project is aim to make an open source alternative to manage project that diverge from the Agile method and more rely on the Simulataneous Engineering (SE). Not all of SE will be implemented but some key concept to focus on the features of the deliverable of the project. Many thing will assume software development but it could probably be used into other application field.

The aim is to provide a project/issues/tasks tracking system that help tracking system features and task that was necessary to acheive thoses features on the system. Being able to track the history of modification as the project progress. So we can see what did change compare to the initial status or any point in time. Branching some modification on an item until it get approved.

This tool aime to have a kind of git built in for work progress and branching for eval/proposition works.

## Install

### Requirements

* Docker or podman with docker compose capability [https://www.docker.com/](https://www.docker.com/)
* Build essential
* (Optional but recommended) Just [https://github.com/casey/just](https://github.com/casey/just)

### Steps

Install `just` cli or type the command from the just file that match the instruction if you prefer.

1. Configure the config file
1. Compile and run the docker images `just docker-build-run`
1. Open you web browser into the displayed link in the console.
1. Config the system.
1. Enjoy!

## App Workflow

### Tips

The goals of this is not to create issues open and close and mesure metric of closes issues. This aim to describe the current situation of the system progress. So focus on making the features and task on that features as a continuous work item in time, do not fear to reopen a tasks, do not create multiple elements that are the same tasks in the end, it help track the change and the evolution of it. Having a task with many modification is better then trying to figure out which out of 500 tasks done rely on the work done. History is key.

Link the element between each other so you can consult them easily.

### Element Type

Main element items:
* `Requirement` : Describe what the project want to create.
* `Feature` : Describe a features the system will need to implement.
* `Task` : A sub of `Feature` that describe a particular work unit that must be done by someone.

Attached items that related to 1 or more parents, can be attached to Main element items.
* `Bug` : A problem identified into the system.
* `Risk`: A foreseen difficulty that might prevent the project.
* `Evaluation`: A part that need to be check for project evaluation.

## License

The [./LICENSE] file dipslay the software license.

## Contribution

The project aime to have a working state before it open up, since the architecture will be modified quiet a lot in time. API and backward compatibility might get broken between version without notices.

## Tests

You should build the tests images when you want to test the setup: `just test-build`

### Unit tests

You can run the unit tests with the test docker by running: `just unit-test-ci-docker`. It will run the unit tests just like the CI. The results appear into the `test-results/` folder.

### Integration tests

You can also make a check of the integration tests if you run the local version of the system. To run the system you should do the `just docker-build-run` and then you should be able to run the tests with `just integration-test-run-docker`.

### Languages

The system will be a backend written in Go and a frontend mostly in Typescript. Trying to minimize packages and framework dependencies to make this stable in time.
