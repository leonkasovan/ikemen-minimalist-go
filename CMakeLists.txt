cmake_minimum_required(VERSION 3.16)
project(ikemen-minimalist
    VERSION 1.0.0
    LANGUAGES CXX C
)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# -------------------------------------------------------------------
#  Dependencies
# -------------------------------------------------------------------

# SDL2 — try pkg-config first, then find module fallback
find_package(SDL2 QUIET)
if(NOT SDL2_FOUND)
    find_path(SDL2_INCLUDE_DIR SDL2/SDL.h)
    find_library(SDL2_LIBRARY SDL2)
    if(SDL2_INCLUDE_DIR AND SDL2_LIBRARY)
        set(SDL2_FOUND TRUE)
    endif()
endif()

if(NOT SDL2_FOUND)
    message(FATAL_ERROR
        "SDL2 not found. Install SDL2 development libraries.\n"
        "  Windows/MSYS2: pacman -S mingw-w64-x86_64-SDL2\n"
        "  macOS:         brew install sdl2\n"
        "  Ubuntu/Debian: apt install libsdl2-dev\n"
        "  Also ensure SDL2_GL_GetProcAddress is available.")
endif()

# -------------------------------------------------------------------
#  Source files
# -------------------------------------------------------------------

set(SOURCES
    src/main.cpp
    src/air.cpp
    src/sff.cpp
    src/animation_instance.cpp
    src/drawlist.cpp
    src/renderer.cpp
    src/glad.c
    src/stb_image.cpp
)

set(HEADERS
    src/types.h
    src/air.h
    src/sff.h
    src/animation_instance.h
    src/drawlist.h
    src/renderer.h
    include/glad/glad.h
)

# -------------------------------------------------------------------
#  stb_image — optional PNG sprite support
#  Download from: https://raw.githubusercontent.com/nothings/stb/master/stb_image.h
#  Place at: ${CMAKE_SOURCE_DIR}/include/stb/stb_image.h
#  Then pass -DIKEMEN_HAVE_STB_IMAGE=ON to cmake.
# -------------------------------------------------------------------
option(IKEMEN_HAVE_STB_IMAGE "Enable PNG sprite support via stb_image" OFF)
if(IKEMEN_HAVE_STB_IMAGE)
    if(NOT EXISTS ${CMAKE_SOURCE_DIR}/include/stb/stb_image.h)
        message(FATAL_ERROR
            "stb_image.h not found at include/stb/stb_image.h.\n"
            "Download it from: https://raw.githubusercontent.com/nothings/stb/master/stb_image.h")
    endif()
    target_compile_definitions(ikemen-minimalist PRIVATE IKEMEN_HAVE_STB_IMAGE)
    message(STATUS "stb_image enabled: PNG sprite support active")
else()
    message(STATUS "stb_image disabled: PNG sprites will show a warning")
endif()

# -------------------------------------------------------------------
#  Build target
# -------------------------------------------------------------------

add_executable(ikemen-minimalist ${SOURCES} ${HEADERS})

target_include_directories(ikemen-minimalist PRIVATE
    ${CMAKE_SOURCE_DIR}/include
    ${SDL2_INCLUDE_DIR}
)

target_link_libraries(ikemen-minimalist PRIVATE
    ${SDL2_LIBRARY}
    # On Windows you may also need:
    # ${SDL2MAIN_LIBRARY}
)

# OpenGL linking (desktop GL)
if(WIN32)
    target_link_libraries(ikemen-minimalist PRIVATE opengl32)
elseif(APPLE)
    find_library(COCOA_LIBRARY Cocoa)
    find_library(OPENGL_LIBRARY OpenGL)
    target_link_libraries(ikemen-minimalist PRIVATE
        ${COCOA_LIBRARY} ${OPENGL_LIBRARY})
else()
    target_link_libraries(ikemen-minimalist PRIVATE GL)
endif()

# Compiler warnings
if(MSVC)
    target_compile_options(ikemen-minimalist PRIVATE /W4)
else()
    target_compile_options(ikemen-minimalist PRIVATE -Wall -Wextra -Wpedantic)
endif()

# -------------------------------------------------------------------
#  Install
# -------------------------------------------------------------------
install(TARGETS ikemen-minimalist RUNTIME DESTINATION bin)
