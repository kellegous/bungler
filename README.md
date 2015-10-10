# Bungler is a stupid way to get java libraries from [Maven Central](http://search.maven.org/).

## Overview
Maven is not my favorite software, but it seems to have infected the Java world
making it damn near impossible to setup even the simplest of projects without
hours wasted just trying to find the right dependencies and setup something to
download them without covering yourself in the filth that is maven as a build
system. "I JUST WANT THE FUCKING JARS!" I usually scream after losing hours.

Bungler is just a thing to download the fucking jars.

## Installing
Bungler is written in Go. I would have written it in java, but I had nothing
reasonable to download the jars.

Install with
`go get github.com/kellegous/bungler`

## Use
Bungler just downloads jars; there is no cache and no weird central location for
jars. Here are some examples:

#### Download the latest Guava into lib
```
bungler --dst=lib com.google.guava/guava
```

#### Download Guava version 18.0 into lib
```
bungler --dst=lib com.google.guava/guava/18.0
```

By default, bungler downloads the binary and source jars. You can request the
javadoc jars with the `--types` flag.

#### Download all types of jars
```
bungler --types=src,jar,doc junit/junit
```

#### Download multiple libs at once
```
bungler com.google.guava/guava junit/junit
```

## TODO:
 * Bungler should not download jars if they are already present, unless a force flag is given.
 * Support all the random shit that maven supports.
