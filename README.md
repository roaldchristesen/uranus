# Uranus
## Digitally Mapping and Providing the Event Landscape

_Disclaimer: This repository and the associated database are currently in a beta version. Some aspects of the code and data may still contain errors. Please contact us via email or create an issue on GitHub if you discover any issues._

## Introduction

With this project, we aim to create a detailed and flexible representation of the event landscape. Our goal is to make it easier to create, maintain, publish, and share high-quality event data.

The Uranus database provides a differentiated representation of event venues (Venues), space descriptions (Spaces), and organizations, for example, in a hierarchical structure such as Institution > Association > Working Group. Additionally, it enables the description of events (Events) with flexible scheduling (EventDate).

The generated data is accessible via an open API in various combinations. This allows the development of plugins (e.g., for WordPress) or integrations for websites.

Unlike other systems that primarily focus on individual events, Uranus also considers the relationships between locations, spaces, dates, and organizations. This opens up new possibilities for queries and applications. Locations and events can be visualized and presented in innovative ways on maps and portals.


## Target Groups
-	Event Organizers: Anyone who publicly offers events with participation opportunities, such as:
	-	Associations, initiatives, educational institutions
	-	Organizations in the fields of culture, leisure, and sports
-	Event Enthusiasts: Users who are actively searching for events, based on criteria such as:
	-	Dates
	-	Target audiences
	-	Event types and genres
	-	Event locations
	-	Organizers
	-	Accessibility-friendly events
	-	Festivals, exhibitions, and event series
-	Portals and Institutions: Associations, municipalities, or website operators who want to integrate event data into their platforms
- Journalists and Culture Reporters: Anyone looking for information about cultural events


## Project Status

The project was initiated on March 2, 2025 by OK Lab Flensburg.
We are currently working on the first MVP (Minimal Viable Product) with the goal of presenting the core concept of Uranus in a demo.
The MVP will include essential features such as event data creation and management, as well as API availability.
Additional features will be added in the coming months.

## Installation

### Prerequisites

1. **Database Setup**

- Ensure PostgreSQL with PostGIS extension is installed and running on `localhost` (default port: `5432`).
- Create a database named `uranus`, owned by a user with the same name.
- Make sure the database accepts connections from `localhost`.

2. **Environment Variables**


### Steps

## Export Data

```sh
pg_dump -U oklab -h localhost -d oklab -n uranus --data-only --column-inserts --no-owner --no-comments --verbose -f uranus_data_dump.sql
pg_dump -U oklab -h localhost -d oklab -n uranus --schema-only --no-owner --no-comments --verbose -f uranus_schema_dump.sql
```
