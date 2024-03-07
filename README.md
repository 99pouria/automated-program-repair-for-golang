# Automated Program Repair for Golang

This project is an automated program repair tool specifically designed for Golang. It aims to assist developers by automatically fixing issues in their Golang code, particularly focusing on concurrency-related problems such as the lack of wait groups in concurrent code.

## Overview

The program takes a path to a Golang code file, the name of a function within that file, and the path to its test cases as input. It then analyzes the code and test cases to identify issues, specifically looking for concurrency problems like missing wait groups. Once identified, the program attempts to automatically repair these issues.

## Features

- **Concurrency Issue Detection**: The tool can identify and repair concurrency issues, specifically focusing on the lack of wait groups in concurrent Golang code.
- **Function-specific Repair**: It operates on a specified function within the Golang code, ensuring that only relevant repairs are made.
- **Automatic Repair**: The tool automatically attempts to fix identified issues, reducing the manual effort required to debug and fix concurrency problems.

## Usage

To run the program, use the following command structure:

```
make run path=path/to/go/code.go func=funcname tests=path/to/testcases.txt
```

Replace `path/to/go/code.go` with the path to your Golang code file, `funcname` with the name of the function you want to repair, and `path/to/testcases.txt` with the path to your test cases file.

## Example

```bash
make run path=examples/go_sum.go func=CSum tests=examples/testcase_go_sum.txt
```

This command will analyze the `CSum` within `./examples/go_sum.go` and attempt to repair concurrency issues based on the test cases provided in `./examples/testcase_go_sum.txt`.

## Note

This tool currently focuses on fixing concurrency issues, specifically lack of wait groups in Golang code. Future versions may include additional repair capabilities for other types of issues.

## Contributing

Contributions to enhance the tool's capabilities or fix bugs are welcome. Please open an issue or submit a pull request with your changes.

## License

This project is licensed under the MIT License. See the `LICENSE` file for more details.
