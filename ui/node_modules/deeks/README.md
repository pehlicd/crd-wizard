# deeks - Deep Object Key Extraction

[![NPM version](https://img.shields.io/npm/v/deeks.svg)](https://www.npmjs.org/package/deeks)
[![Downloads](https://img.shields.io/npm/dm/deeks)](https://www.npmjs.org/package/deeks)
[![Minzipped Size](https://img.shields.io/bundlephobia/minzip/deeks)](https://bundlephobia.com/result?p=deeks)
[![Build Status](https://img.shields.io/github/actions/workflow/status/mrodrig/deeks/automated-tests-workflow.yml)](https://github.com/mrodrig/deeks/actions/workflows/automated-tests-workflow.yml)
[![Coverage Status](https://coveralls.io/repos/github/mrodrig/deeks/badge.svg?branch=main)](https://coveralls.io/github/mrodrig/deeks?branch=main)
[![Typings](https://img.shields.io/npm/types/deeks)](https://www.npmjs.org/package/deeks)

**Retrieve all keys and nested keys from objects and arrays of objects.**

## Installing

```bash
npm install --save deeks
```

Example: 
```javascript
const { deepKeys } = require('deeks');
// Alternatively:
// import { deepKeys } from 'deeks';

let generatedKeys = deepKeys({
	make: 'Nissan',
	model: 'GT-R',
	trim: 'NISMO',
	specifications: [
	    {mileage: 10},
	    {cylinders: 6}
	]
}, {
    expandArrayObjects: true,
    ignoreEmptyArraysWhenExpanding: true
});
// => ['make', 'model', 'trim', 'specifications.mileage', 'specifications.cylinders']
```

## API

### deepKeys 

`deepKeys(object, options)`

Returns all keys in an object, even if they're nested several layers deep. 
The array of keys that is returned can then be used with the 
[`doc-path`](https://github.com/mrodrig/doc-path) module to get and set values 
at a specific key path.

Options (optional):
- arrayIndexesAsKeys - `Boolean` (Default `false`) - Should array indexes be used as keys in the generated path?
	- Example:
	```json
	[
		{
			"list": [
				{
					"a": 1
				},
				{
					"a": 2
				}
			]
		}
	]
	```
	- arrayIndexesAsKeys = `false` results in: `['list.a']`
	- arrayIndexesAsKeys = `true` results in: `['list.0.a', 'list.1.a']`
- expandNestedObjects - `Boolean` (Default: `true`) - Should nested objects appearing in the provided object also be expanded, such asthat keys appearing in those objects are extracted and included in the returned key path list?
	- Example:
	```json
	{
		"make": "Nissan",
		"model": "Murano",
		"year": 2013,
		"specifications": {
			"mileage": 7106,
			"trim": "S AWD"
		}
	}
	```
	- expandNestedObjects = `false` results in: `['make', 'model', 'year', 'specifications']`
	- expandNestedObjects = `true` results in: `['make', 'model', 'year', 'specifications.mileage', 'specifications.trim']`
- expandArrayObjects - `Boolean` (Default: `false`) - Should objects appearing in arrays in the provided 
object also be expanded, such that keys appearing in those objects are extracted and 
included in the returned key path list?
	- Example:
	```json
	{
		"make": "Nissan",
		"model": "GT-R",
		"trim": "NISMO",
		"specifications": [
			{"mileage": 10},
			{"cylinders": 6}
		]
	}
	```
	- expandArrayObjects = `false` results in: `['make', 'model', 'trim', 'specifications']`
	- expandArrayObjects = `true` results in: `['make', 'model', 'trim', 'specifications.mileage', 'specifications.cylinders']`
- ignoreEmptyArraysWhenExpanding - `Boolean` (Default: `false`) - Should empty array keys be ignored when expanding array objects?
	- Note: This only has an effect when used with `expandArrayObjects`.
	- Example:
	```json
	{ 
		"features": [ {"name": "A/C" }],
		"rebates": []
	}
	```
	- ignoreEmptyArraysWhenExpanding = `false` results in: `['features.name', 'rebates']`
	- ignoreEmptyArraysWhenExpanding = `true` results in: `['features.name']`
- escapeNestedDots - `Boolean` (Default: `false`) - Should `.` characters that appear in keys be escaped with a preceding `\` character?
	- Example:
	```json
	{
		"a.a": "1",
		"a.b": {
			"c": "2",
			"c.d": "3"
		}
	}
	```
	- escapeNestedDots = `false` results in: `['a.a', 'a.b.c', 'a.b.c.d']`
	- escapeNestedDots = `true` results in: `['a\\.a', 'a\\.b.c', 'a\\.b.c\\.d']`
- ignoreEmptyArrays - `Boolean` (Default: `false`) - Should key paths for empty arrays be ignored in the generated key list?
	- Example:
	```json
	{
		"a": {
			"b": [],
			"c": {
				"f": 4,
				"e": []
			}
		},
		"b": []
	}
	```
	- ignoreEmptyArrays = `false` results in `['a.b', 'a.c.f', 'a.c.e', 'b']`
	- ignoreEmptyArrays = `true` results in `['a.c.f']`

Returns: `string[]`

Example: `['make', 'model', 'specifications.odometer.miles', 'specifications.odometer.kilometers']`

### deepKeysFromList

`deepKeysFromList(array, options)`

Returns all keys in each object in the array, even if the keys are nested 
several layers deep in each of the documents. These can also be used with the 
[`doc-path`](https://github.com/mrodrig/doc-path) module.

Options (optional):
- expandNestedObjects - `Boolean` (Default: `true`) - Should nested objects appearing in the provided object also be expanded, such asthat keys appearing in those objects are extracted and included in the returned key path list?
	- Example:
	```json
	{
		"make": "Nissan",
		"model": "Murano",
		"year": 2013,
		"specifications": {
			"mileage": 7106,
			"trim": "S AWD"
		}
	}
	```
	- expandNestedObjects = `false` results in: `['make', 'model', 'year', 'specifications']`
	- expandNestedObjects = `true` results in: `['make', 'model', 'year', 'specifications.mileage', 'specifications.trim']`
- expandArrayObjects - `Boolean` (Default: `false`) - Should objects appearing in arrays in the provided 
object also be expanded, such that keys appearing in those objects are extracted and 
included in the returned key path list?
	- Example:
	```json
	{
		"make": "Nissan",
		"model": "GT-R",
		"trim": "NISMO",
		"specifications": [
			{"mileage": 10},
			{"cylinders": 6}
		]
	}
	```
	- expandArrayObjects = `false` results in: `['make', 'model', 'trim', 'specifications']`
	- expandArrayObjects = `true` results in: `['make', 'model', 'trim', 'specifications.mileage', 'specifications.cylinders']`
- ignoreEmptyArraysWhenExpanding - `Boolean` (Default: `false`) - Should empty array keys be ignored when expanding array objects?
	- Note: This only has an effect when used with `expandArrayObjects`.
	- Example:
	```json
	[
		{ "features": [ {"name": "A/C" }] },
		{ "features": [] }
	] 
	```
	- ignoreEmptyArraysWhenExpanding = `false` results in: `[ ['features.name', 'features'] ]`
	- ignoreEmptyArraysWhenExpanding = `true` results in: `[ ['features.name'] ]`
- escapeNestedDots - `Boolean` (Default: `false`) - Should `.` characters that appear in keys be escaped with a preceding `\` character.
	- Example:
	```json
	[
		{
			"a.a": "1",
			"a.b": {
				"c": "2",
				"c.d": "3"
			}
		}
	]
	```
	- escapeNestedDots = `false` results in: `[ ['a.a', 'a.b.c', 'a.b.c.d'] ]`
	- escapeNestedDots = `true` results in: `[ ['a\\.a', 'a\\.b.c', 'a\\.b.c\\.d'] ]`
- ignoreEmptyArrays - `Boolean` (Default: `false`) - Should key paths for empty arrays be ignored in the generated key list?
	- Example:
	```json
	[
		{
			"a": {
				"b": [],
				"c": {
					"f": 4,
					"e": []
				}
			},
			"b": []
		}
	]
	```
	- ignoreEmptyArrays = `false` results in `[ ['a.b', 'a.c.f', 'a.c.e', 'b'] ]`
	- ignoreEmptyArrays = `true` results in `[ ['a.c.f'] ]`

Returns: `string[][]`

Example: `[ ['make', 'model', 'specifications.odometer.miles', 'specifications.odometer.kilometers'] ]`

## Examples

This module integrates really nicely with the 
[`doc-path`](https://github.com/mrodrig/doc-path) module, which allows
the programmatic getting and setting of key paths produced by this module.

Additionally, `doc-path@>=3` works with the keys returned when the `escapeNestedDots` option is specified.

Here's an example of how this works:

```javascript
const path = require('doc-path'),
      { deepKeys } = require('deeks');

let car = {
		make: 'Nissan',
		model: 'GT-R',
		trim: 'NISMO',
		specifications: {
			mileage: 10,
			cylinders: '6'
		}
	},
	
	carKeys = deepKeys(car);

for(let keyPath of carKeys) {
    // Clear all values
    path.setPath(car, keyPath, '');
}
```

## Tests

```bash
$ npm test
```

_Note_: This requires `mocha`.

To see test coverage, please run:
```bash
$ npm run coverage
```
