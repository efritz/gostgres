# Gostgres

Postgres in Go as a learning exercise.

## Try me out

Simply run `go build ./cmd/repl && ./repl` to drop into a psql-like shell where you can issue SQL commands to an in-memory database.

This shell starts with an empty tablespace, but a sample database based on [pagila](https://github.com/devrimgunduz/pagila) can be loaded via `load sample`.

```
$ go build ./cmd/repl && ./repl
gostgres ❯ load sample
gostgres ❯ SELECT f.title, f.rating, c.name
gostgres ❯ FROM film f
gostgres ❯ JOIN film_category fc ON fc.film_id = f.film_id
gostgres ❯ JOIN category c ON c.category_id = fc.category_id
gostgres ❯ WHERE rating = 'R'
gostgres ❯ ORDER BY f.title
gostgres ❯ LIMIT 20;
         title        | rating |     name
----------------------+--------+-------------
      AIRPORT POLLOCK |      R |      Horror
           ALONE TRIP |      R |       Music
  AMELIE HELLFIGHTERS |      R |       Music
      AMERICAN CIRCUS |      R |      Action
 ANACONDA CONFESSIONS |      R |   Animation
     ANALYZE HOOSIERS |      R |      Horror
    ANYTHING SAVANNAH |      R |      Horror
 APOCALYPSE FLAMINGOS |      R |         New
     ARMY FLINTSTONES |      R | Documentary
          BADMAN DAWN |      R |      Sci-Fi
     BANGER PINOCCHIO |      R |       Music
       BEAR GRACELAND |      R |    Children
      BEAST HUNCHBACK |      R |    Classics
       BEVERLY OUTLAW |      R |      Sci-Fi
        BOOGIE AMELIE |      R |       Music
        BOULEVARD MOB |      R |         New
      BROOKLYN DESERT |      R |     Foreign
  BROTHERHOOD BLANKET |      R | Documentary
        BUBBLE GROSSE |      R |      Sports
      CAMPUS REMEMBER |      R |      Action
(20 rows)

Time: 187.273833ms
```
