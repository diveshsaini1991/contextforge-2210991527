package calculator

import (
	"errors"
	"math"
)

// Add returns the sum of two integers
func Add(a, b int) int {
	return a + b
}

// Subtract returns the difference between two integers
func Subtract(a, b int) int {
	return a - b
}

// Multiply returns the product of two integers
func Multiply(a, b int) int {
	return a * b
}

// Divide returns the quotient of two integers
// Returns an error if dividing by zero
func Divide(a, b int) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return float64(a) / float64(b), nil
}

// Power calculates a raised to the power of b
func Power(a, b float64) float64 {
	return math.Pow(a, b)
}

// Factorial calculates the factorial of n
// Returns an error for negative numbers
func Factorial(n int) (int, error) {
	if n < 0 {
		return 0, errors.New("factorial of negative number")
	}
	if n == 0 || n == 1 {
		return 1, nil
	}
	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return result, nil
}

// isPrime checks if a number is prime (unexported function)
func isPrime(n int) bool {
	if n <= 1 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}
	for i := 3; i*i <= n; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}

// PrimeFactors returns the prime factors of n
func PrimeFactors(n int) []int {
	if n <= 1 {
		return []int{}
	}

	factors := []int{}
	for i := 2; i <= n; i++ {
		for n%i == 0 {
			factors = append(factors, i)
			n = n / i
		}
	}
	return factors
}
