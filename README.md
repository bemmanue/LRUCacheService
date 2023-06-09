# Задание

Написать сервис кэширования. Кэш ограниченной емкости, метод вытеснения ключей lru. Сервис должен быть потокобезопасный, принимать любые значения

Сервис должен реализовывать минимально следующий интерфейс:
```go
type ICache interface {
    Add(key, value interface{})
    Get(key interface{}) (value interface{}, ok bool)
    Remove(key interface{})
}
```

Лучше так:
```go
type ICache interface {
    Cap() int
    Clear()
    Add(key, value interface{})
    Get(key interface{}) (value interface{}, ok bool)
    Remove(key interface{})
}
```

Совсем хорошо:
```go
type ICache interface {
    Cap() int
    Clear()
    Add(key, value interface{})
    AddWithTTL(key, value interface{}, ttl time.Duration)
    Get(key interface{}) (value interface{}, ok bool)
    Remove(key interface{})
}
```