# Project Works

## Introduction

The project want to create an projects tasks tracking system. I will need to have different naming and able to link things and display them in a more comprehensive and trackable way. This is an open source project and should only use open source dependencies.

## Simultaneous engineering

The idea is not based on sprint like Agile. The system is based on simultaneous engineering concept. It will used some of the concept but not all from that method. The following will be used into the application to ease the workflow:

* DFP matrix to represent
* Coupling matrix : give interdependencies between tasks
* Gant diagram : for giving an idea of the release date.
* Pert diagram : for critical path of the project.

## Structure

### General 

The code will need to be uncouple as much as possible to allow further extension and customization. Having a fast view mode to fetch data super fast.  The possibility to add custom fields to element type. Having the GUI into his own part and the backend server on his own. I will want to have the web infra in Go for fast serving. The frontend shoudl prefer Typescript over Javascript. The use of recent GUI framework is allowed but limit his usage to the GUI and not the whole application.

### Code split

Make the frontend and backend code as split as possible. The choice of database should be modular, so changing the database should only change that, the interface for database operation should not change. Quick search and retrieves of many information is wanted. Swapping the databse under the hood as a plugin is also wanted.

Adding more features with a plugins based system for the GUI. The plugins should be load upon startup and can be (de)activated.

The actual model and domain should be fast to write to it. The view could use some caching to reconstruct some dataview (all features childs elements for example, invalidate the cache only when adding or removing child for that particular features). Their will be many more reader then writer for the web pages. Design the architecture to allow that.

The backend should offer and OpenAPI interface that the frontend will use. A cli should be able to call and modify the backend without web browser (with curl and an API key for example).

### Storage

The system will use a database to store the information for the projects elements. their will be many projects, each containing few hundred or more elements each. the file attachments should be save into another file storage (something like an s3 or anything like that).

## Behaviors

### Config 

Make a config file to :

* super admin local account (avoid loosing access if external authentication is lost)
* activate the plugins.
* configure the database max space.
* configure the file storage max space and max file size.
* configure the web server domain, port.

### Authentication

The system should connect to Office 365 to login user. But could use other authentication system. The system need to have a group where user can be added to groups and groups has access to different projects (no individual level access per project). By default user has only anonymous group. Also make API key for automation.

### Projects

The system shall allow multiples projects, which can be classify into multi depth folder for navigation ease.

### Filter and select

The usage of a markup query languages to filter or select some fields is also wanted across the system. So we can filter by project name, tasks type. Modification/creation date time, etc.

### Display

The system will need to display as table list the items with a hierarchical display the elements of a project. Also make it possible to display matrix of related items. 

Each element type can be open as a card. Reference to other card id are clickable to see their card. Ideally the car view can be a stack and we can go back and forward into that stack quickly.

The DPF matrix for example should do: 
* row are requirements
* columns are the features (span over multiple columns of all sub tasks) with task under (1 task / columns).

The Coupling matrix (half matrix, diagonal and below is redondant):
* Tasks vs Tasks

### Fields

All element type should have at least the following:

* `id` : to link to the element and need to be indexed
* `element_type` : repressent (Features, Tasks, Requirements, Bug, Evaluation, Risk...).
* `title` : the title of the task
* `description` : a markdown capable description text with mermaid markdown capability.
* `creation_time` : readonly fill by the system, create date time
* `modification_time` : readonly fill by the system, last modified time
* `sha` : readonly fill by the system, the version sha of the current item so we can compare diff and see history like git on the item
* `linked_tasks` : Other tasks of any type that are related to this one. Each link rely 2 element type with extra information (see link information section).
* `attachements` : attached files to the element (photo, document, save, ...)
* `custom_fields` : possible to add as many custom filds with type (string, textarea, integer, real, choices...)

##### Link information

The link between element type should have the following information:

* `src_element` : the source of the link.
* `dest_element` : the destination of the element.
* `link_type` : The link type used

###### Link type

* `related` : indicate the context is related to the other element 
    * `bug`, `eval` use this
    * all element type can use this relationship.
* `child` : child of other element
    * for `feature` as src and `task` as destination
* `implement` : 
    * for `requirement` as destination and `task` or `feature` as src.

#### Requirements

* `interest` : the level by which the client require this (1-10 scale)
* `supervisors` : the personnes who are the reference for that requirements
* `clients` : clients contact that provide information about this requirements.

#### Features

* `progress` : readonly fill by the system, a virtual property that use the progression of all childs tasks
* `supervisors` : the personnes who are the reference for that feature

#### Task

* `assignee` : the personne in charge of doing the task.
* `status` : the current state of the task (see the Task status section)
* `progress` : the work percentage
* `close_time` : readonly fill by the system when the task was closed, the status is a close one.
* `parent_feature` : feature that task is linked to.
* `start_phase` : The phase into which this task should be started
* `delivery_phase` : The phase into which this task should be completed

##### Task status

* `backlog` : open status, when the tasks have noto been set to be work on yet.
* `todo` : open status, work item that need to be done.
* `in_progress` : open status, the work has start but is not completed yet.
* `in_review` : open status, the element is being review.
* `testing` : open status, the element is being tested.
* `blocked` : open status, the element cannot progress for the time being.
* `done` : close status. the element is completed.
* `rejected` : the element is rejected and will not be done.

#### Bug

#### Eval
