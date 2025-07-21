# Code Example Notes

This is an example note that demonstrates code syntax highlighting functionality.

## Python Code Examples

### Basic Python Function
```python
def fibonacci(n):
    """Calculate the nth term of the Fibonacci sequence"""
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

# Usage example
for i in range(10):
    print(f"F({i}) = {fibonacci(i)}")
```

### Class Definition
```python
class Calculator:
    def __init__(self):
        self.result = 0
    
    def add(self, x, y):
        return x + y
    
    def multiply(self, x, y):
        return x * y

calc = Calculator()
print(calc.add(5, 3))  # Output: 8
```

## JavaScript Code Examples

### Modern JavaScript Features
```javascript
// Arrow functions
const greet = (name) => `Hello, ${name}!`;

// Destructuring assignment
const { title, author } = { title: "JavaScript Guide", author: "John Doe" };

// Async functions
async function fetchData(url) {
    try {
        const response = await fetch(url);
        const data = await response.json();
        return data;
    } catch (error) {
        console.error('Error:', error);
    }
}
```

## Go Code Examples

### Go Structs and Interfaces
```go
package main

import "fmt"

type Animal interface {
    Speak() string
}

type Dog struct {
    Name string
}

func (d Dog) Speak() string {
    return fmt.Sprintf("%s says: Woof!", d.Name)
}

func main() {
    dog := Dog{Name: "Buddy"}
    fmt.Println(dog.Speak())
}
```

## CSS Style Examples

### Modern CSS Features
```css
/* CSS Grid Layout */
.container {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 20px;
    padding: 20px;
}

/* Custom Properties */
:root {
    --primary-color: #007bff;
    --secondary-color: #6c757d;
}

.button {
    background-color: var(--primary-color);
    color: white;
    padding: 10px 20px;
    border: none;
    border-radius: 5px;
    cursor: pointer;
    transition: background-color 0.3s ease;
}

.button:hover {
    background-color: var(--secondary-color);
}
```

## SQL Query Examples

### Complex Queries
```sql
-- Multi-table join query
SELECT 
    u.username,
    p.title,
    COUNT(c.id) as comment_count
FROM users u
LEFT JOIN posts p ON u.id = p.user_id
LEFT JOIN comments c ON p.id = c.post_id
WHERE p.created_at >= '2024-01-01'
GROUP BY u.id, p.id
HAVING comment_count > 0
ORDER BY comment_count DESC;
```

## Inline Code

You can use `inline code` in text to mark code snippets, such as `console.log()` or `print()` functions.

## Summary

Code example notes demonstrate:
- Syntax highlighting for multiple programming languages
- Formatted display of code blocks
- Usage of inline code
- Feature showcase of different languages 