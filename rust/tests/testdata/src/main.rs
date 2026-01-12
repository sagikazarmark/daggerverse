fn main() {
    println!("Hello, world!");
}


#[cfg(test)]
mod tests {
    #[test]
    fn test_test() {
        assert_eq!("hello", "hello");
    }
}
