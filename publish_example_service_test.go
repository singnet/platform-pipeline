package main

import (
	"fmt"
	"time"

	"github.com/DATA-DOG/godog/gherkin"
)

func exampleserviceIsRegistered(table *gherkin.DataTable) error {
	return serviceIsRegistered(table, exampleServiceDir)
}

func exampleserviceIsPublishedToNetwork() error {
	return serviceIsPublishedToNetwork(exampleServiceDir, "./service.json")
}

func exampleserviceIsRunWithSnetdaemon(table *gherkin.DataTable) error {

	daemonPort := getTableValue(table, "daemon port")
	ethereumEndpointPort := getTableValue(table, "ethereum endpoint port")
	passthroughEndpointPort := getTableValue(table, "passthrough endpoint port")

	snetdConfigTemplate := `
	{
    "AUTO_SSL_DOMAIN": "",
    "AUTO_SSL_CACHE_DIR": "",
    "BLOCKCHAIN_ENABLED": true,
    "CONFIG_PATH": "",
    "DAEMON_LISTENING_PORT": %s,
    "DAEMON_TYPE": "grpc",
    "DB_PATH": "./db",
    "ETHEREUM_JSON_RPC_ENDPOINT": "http://localhost:%s",
    "EXECUTABLE_PATH": "",
    "LOG_LEVEL": 5,
    "PASSTHROUGH_ENABLED": true,
    "PASSTHROUGH_ENDPOINT": "http://localhost:%s",
    "POLL_SLEEP": "",
    "PRIVATE_KEY": "%s",
    "SERVICE_TYPE": "jsonrpc",
    "SSL_CERT": "",
    "SSL_KEY": "",
    "WIRE_ENCODING": "json"
    }`

	snetdConfig := fmt.Sprintf(snetdConfigTemplate,
		daemonPort, ethereumEndpointPort, passthroughEndpointPort, accountPrivateKey)

	file := exampleServiceDir + "/snetd.config.json"
	err := writeToFile(file, snetdConfig)

	if err != nil {
		return err
	}

	linkFile(envSingnetRepos+"/snet-daemon/build/snetd-linux-amd64", exampleServiceDir+"/snetd-linux-amd64")

	outputFile := logPath + "/example-service.log"

	fileContains := checkFileContains{
		output:  outputFile,
		strings: []string{},
	}

	command := ExecCommand{
		Command:    exampleServiceDir + "/scripts/run-snet-service",
		Directory:  exampleServiceDir,
		OutputFile: outputFile,
	}

	err = runCommandAsync(command)

	if err != nil {
		return err
	}

	_, err = checkWithTimeout(5000, 500, checkFileContainsStringsFunc(fileContains))

	if err != nil {
		return err
	}

	time.Sleep(2 * time.Second)

	command = ExecCommand{
		Command:   exampleServiceDir + "/scripts/test-call",
		Directory: exampleServiceDir,
	}

	return runCommand(command)
}

func singularityNETJobIsCreated(table *gherkin.DataTable) error {

	maxPrice := getTableValue(table, "max price")

	args := []string{
		"create-jobs",
		"--funded",
		"--signed",
		"--max-price", maxPrice,
		"--yes",
	}

	command := ExecCommand{
		Command:   "snet",
		Directory: exampleServiceDir,
		Args:      args,
	}

	err := runCommand(command)

	if err != nil {
		return err
	}

	args = []string{
		"client", "call", "classify",
		fmt.Sprintf(`{"image_type": "jpg", "image": "%s"}`, testImage),
	}

	command = ExecCommand{
		Command:   "snet",
		Directory: exampleServiceDir,
		Args:      args,
	}

	return runCommand(command)
}

var testImage = "/9j/4AAQSkZJRgABAQAAAQABAAD/2wCEAAkGBxAPDw0PDQ0PDg0PDQ0PDQ0PDQ8ODQ0NFRIWFhUSExUYHyghGB4lJxMWITEhJSor" +
	"Li4uGB8zODMsNygtLisBCgoKDQ0NDxAPFSsdFRktLSs4LCsrKysrKysrKzcrLSsrKy0tKzcrKysrKysrKystNy0rKysrKysrKysr" +
	"KysrK//AABEIAOEA4QMBIgACEQEDEQH/xAAcAAEAAQUBAQAAAAAAAAAAAAAAAQIDBAUGBwj/xABAEAEAAgECAgcEBAoLAQAAAAAA" +
	"AQIDBBEFIQYHEhMxQWFRcZGxMkKBghQiI3KDkqGissEIF1NUYpOjwtHh8BX/xAAWAQEBAQAAAAAAAAAAAAAAAAAAAQL/xAAWEQEB" +
	"AQAAAAAAAAAAAAAAAAAAEQH/2gAMAwEAAhEDEQA/APcQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" +
	"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" +
	"AEJavjHGKabsRe1K3vv2O3aKVnbbfn9scgbRG7jtV0g1XjSKxWfCa0i1fjzYF+OaqfHPMfmxWPkD0Dc3ee//AEs0/SzXn70q6ay0" +
	"+N7frSlHf7m7iseaZ+tPxlk4r+s/GVHWG7n8NLz4Tf8AWmGTSuSNvyto+9MiVuBqq58sfW329tYU243XHMRmmkbzFYmLecztHKRW" +
	"3EQkAAAAAAAAAAAAAABby56UibXvWlY8bWmKxH2y13SPjWPRae+fJtO3KlZtFe3f2b+UcpmZ8oiXzv0g41rOOavutPGfUT2prWkX" +
	"tGGN5+rjjlWvrbefOdvAH0Nk6T6Cv0tdp4/TUl551q8c0mpppq6fU481/wAtE1pbteVZ5uMp1McVmsTP4LW08+xOptvHpvFWB0g6" +
	"DcQ4Vp41GpjT1xfhGPfusk5LTNpiu2+0bRyhUYmk4jk0/b7jJOObREb18vWPZ4s7F0y1lYiLXrl287d5Fp98xZps/jLGmVSOjv03" +
	"1EzMzHYjs8opett7b+M9uszt6Ndbp7xKsztfBNd+W+Cm+3q02SWLkNV0tesjicf3f/Jj/lcjrO4p7dNH6H/tyMoQdl/Wjxfy1GGv" +
	"l+LhhTbrH4raPxtZbf8Aw1rVx6uAdFm6Wa/J9PWZ538fykx8mz6LZb5dXpu8vbJPfYtptabfXj2uQxO06vcXb12jr7dRh393bjcH" +
	"0qECKAAAAAAAAAAAAAA8O/pBcbvXLp9JWZikaeMt9vC1r3tG3+nHxbrqI4TSmnz55rE5N6Ui3nG9Ztaft3j4Of8A6RXDLxl0urrE" +
	"93fDGG0+UXpe1o+MX/Y3HUXxuk0y6e1oi2SKZMUT52rE1vX38o+Erg9ecV1xaLvuC62IjnSMeSPSa3iXaxLXdI9H3+j1WKY37eDJ" +
	"G3rtvHyQfLVrdqtbR9albfGIY9pXcNdsVYnxpNqT762mv8li0tIoySxrr1pWLoKJQlAEKoUK4BfxPSeqLS95r8M+VO1f3bVnb+Tz" +
	"XD4va+ozQ721OaY5Ux0pE+tpmZ/hQevwAKAAAAAAAAAAAAAA03Szo7i4lpMulz8otzx5Ij8bFlj6N4/94TL5u4jwniHANVtel+xF" +
	"+1iy07Vcd4ieVsd48J9PH5vqpj63RYs9Jx58VMuO0bWpkrF6z9kg8e4P14460rXV6bJa8RtN68pn37bxLbf138PmOeDUe6YiGx4t" +
	"1P8AC802tipl0lp/sctrUj3UvvEfY4/inUXkjedJrcWSee1NRinHM/epv8io4HUWre2ovjiYx31WovjiY2mMd8k2pvHumGtyOy4l" +
	"0O1ugwXnXY6xvetaXpkjJS0RTaNp8fKPGHH6iNpUY9lmy5ZasCiUJlAJhMKd1UAyNNG9ofS/VTwzuOG4rTG1s8zln29nwr8v2vn7" +
	"olwy2q1enwUjnkyRX3R5z9kRMvqzSYK4sePHSNqY6UpWPZWsbQir4jc3BIjdIAAAAAAAAAAAAAAAANV0m4VGs0ubBPjakzjn2ZI5" +
	"1l8x8b0VsOTJS9drVtatonymJ2fWEw8f65OjG141uGm9cnLP2a8q5IjlafZv84XDXit1qWTnpsxpEUShMqQSrqt7tl0f4Xk1eoxY" +
	"MUb3yWiPbFY352n0jxQeudRnAezGXX5K7bTOHTzPtmI7do+O3xevd40PCNPTS4MOnxRtjxUiseU2nztPrM82dXNuK2PeJ7xg1uuR" +
	"YGZFlcSxK2X6SC8IhIAAAAAAAAAAAAAACm9ItExaImJ8YmN4mFQDlOK9XfCtVM2y6GlbT9fDfJgn9yYhodR1M8Jn6P4VT0jUzaP3" +
	"ol6SpsDyfN1K6D6up1cfexz/ALWPHUvoonnqdVMe/HH8nrlqLU4geZYOqDhlfpVz5PztRav8OzpODdEtJo940unri35WtG83mPW0" +
	"zM/tdT3SYxA11NIvV07N7tVFAY1MK53a92U9kFqtF2tVUQkBKEgAAAAAAAAAAAAAAAAI2SApmDZUAp2TskBGxskBAkBAkAAAAAAA" +
	"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAB//2Q=="
