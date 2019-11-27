# miniredis
> Experimental key-value persistent DB with TCP.

[![Travis](https://img.shields.io/travis/IzmirGophers/miniredis.svg)](https://travis-ci.org/IzmirGophers/miniredis)
[![Go Report Card](https://goreportcard.com/badge/github.com/IzmirGophers/miniredis)](https://goreportcard.com/report/github.com/IzmirGophers/miniredis)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/IzmirGophers/miniredis)
[![codecov](https://codecov.io/gh/IzmirGophers/miniredis/branch/master/graph/badge.svg)](https://codecov.io/gh/IzmirGophers/miniredis)
[![GitHub version](https://badge.fury.io/gh/IzmirGophers%2Fminiredis.svg)](https://github.com/IzmirGophers/miniredis/releases)


Miniredis is project a mini project written for for GDG Istanbul Golang Workshop.

## Installation

OS X & Linux:

```sh
$ go get https://github.com/IzmirGophers/miniredis
$ cd $GOPATH/src/github.com/IzmirGophers/miniredis
$ go install
```

## Usage example

Miniredis is running on tcp, you can send commands through any client you can establish tcp connection.

## Commands 

| Command | Params | Example |
| ------ | ------ |----------- |
| SET   | key, val | SET foo bar |
| GET | key | GET foo |
| DEL    | key | DEL foo |
| DBSIZE    |  | DBSIZE |
| KEYS    |  | KEYS |


## Benchmark

``WIP``
 
## Meta
<table>
   <tr>
      <td align="center">
          <a href="https://github.com/riza">
              <img src="https://avatars1.githubusercontent.com/u/2565849?s=460&v=4" width="100px;" alt="Sinan Ülker"/>
              <br />
              <sub>
                  <b>Rıza Sabuncu</b>
              </sub>
          </a>
      </td>
  </tr>
</table>


Distributed under the GPL license. See ``LICENCE`` for more information.

## Contributors


<table>
 <tr>
    <td align="center">
    	<a href="https://github.com/unicod3">
    		<img src="https://avatars2.githubusercontent.com/u/2614110?s=460&v=4" width="100px;" alt="Sinan Ülker"/>
    		<br />
   			<sub>
    			<b>Sinan Ülker</b>
    		</sub>
    	</a>
    </td>
    <td align="center">
    	<a href="https://github.com/c1982">
    		<img src="https://avatars2.githubusercontent.com/u/45575?s=460&v=4" width="100px;" alt="Oğuzhan Yılmaz"/>
    		<br />
    		<sub>
    			<b>Oğuzhan Yılmaz</b>
    		</sub>
    	</a>
    </td>
    <td align="center">
    	<a href="https://github.com/hto">
			<img src="https://avatars3.githubusercontent.com/u/3604669?s=460&v=4" width="100px;" alt="Halil Tuğcan Özaktaş"/>
	    <br />
    	<sub>
		    <b>Halil Tuğcan Özaktaş</b>
	    </sub>
    </a>
    </td>
  </tr>
</table>

## Contributing

1. Fork it (<https://github.com/IzmirGophers/miniredis/fork>)
2. Create your feature branch (`git checkout -b feature/fooBar`)
3. Commit your changes (`git commit -am 'Add some fooBar'`)
4. Push to the branch (`git push origin feature/fooBar`)
5. Create a new Pull Request

