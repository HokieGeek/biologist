# biologist [![Build Status](https://travis-ci.org/HokieGeek/biologist.svg?branch=master)](https://travis-ci.org/HokieGeek/biologist) [![Coverage](http://gocover.io/_badge/gitlab.com/HokieGeek/biologist?0)](http://gocover.io/gitlab.com/HokieGeek/biologist) [![GoDoc](http://godoc.org/gitlab.com/hokiegeek/biologist?status.png)](http://godoc.org/gitlab.com/hokiegeek/biologist)

This library makes use of [life](https://gitlab.com/HokieGeek/life), my [Conway's Game of Life](http://www.conwaylife.com/wiki/Conway%27s_Game_of_Life) engine. It enables the client to create instances of a simulation for analysis. I plan on using this as a project for learning machine umm... learning.

Current status of analysis: It can only detect when the simulation goes into a cycle.

The biologistd binary provides a RESTful service for the creation and control of the simulations being analyzed. 

For a GUI frontend, take a look at [biologist-web](https://gitlab.com/HokieGeek/biologist-web)
