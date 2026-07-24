CC      := gcc
TARGET  := {{.Name}}
SRC     := $(wildcard src/*.c)
CFLAGS  := -Wall -Wextra -std=c11

.PHONY: all clean bear

all: $(TARGET)

$(TARGET): $(SRC)
	$(CC) $(CFLAGS) -o $@ $^

# Regenerate compile_commands.json via bear.
# Run this whenever you add/remove source files.
bear:
	bear -- $(MAKE) all

clean:
	rm -f $(TARGET) *.o
