#ifndef GLAD_H_
#define GLAD_H_

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

/* OpenGL 3.3 core constants */
#define GL_TRUE 1
#define GL_FALSE 0

/* Clear bits */
#define GL_COLOR_BUFFER_BIT 0x00004000

/* Shader types */
#define GL_VERTEX_SHADER 0x8B31
#define GL_FRAGMENT_SHADER 0x8B30

/* Shader/program queries */
#define GL_COMPILE_STATUS 0x8B81
#define GL_LINK_STATUS    0x8B82
#define GL_INFO_LOG_LENGTH 0x8B84

/* Buffer targets */
#define GL_ARRAY_BUFFER 0x8892
#define GL_ELEMENT_ARRAY_BUFFER 0x8893

/* Buffer usage */
#define GL_STATIC_DRAW 0x88E4
#define GL_DYNAMIC_DRAW 0x88E8

/* Vertex arrays */
#define GL_VERTEX_ATTRIB_ARRAY_ENABLED 0x8622
#define GL_VERTEX_ATTRIB_ARRAY_STRIDE  0x8624

/* Draw mode */
#define GL_TRIANGLES 0x0004
#define GL_TRIANGLE_STRIP 0x0005
#define GL_TRIANGLE_FAN 0x0006

/* Blend */
#define GL_BLEND 0x0BE2
#define GL_FUNC_ADD 0x8006
#define GL_FUNC_REVERSE_SUBTRACT 0x800B
#define GL_SRC_ALPHA 0x0302
#define GL_ONE 0x0001
#define GL_ONE_MINUS_SRC_ALPHA 0x0303

/* Texture targets */
#define GL_TEXTURE_2D 0x0DE1
#define GL_TEXTURE0 0x84C0
#define GL_TEXTURE1 0x84C1

/* Texture params */
#define GL_TEXTURE_MIN_FILTER 0x2801
#define GL_TEXTURE_MAG_FILTER 0x2800
#define GL_TEXTURE_WRAP_S 0x2802
#define GL_TEXTURE_WRAP_T 0x2803

/* Texture filters */
#define GL_NEAREST 0x2600
#define GL_LINEAR 0x2601
#define GL_CLAMP_TO_EDGE 0x812F

/* Internal formats */
#define GL_R8 0x8229
#define GL_RGBA8 0x8058

/* Base formats */
#define GL_RED 0x1903
#define GL_RGB 0x1907
#define GL_RGBA 0x1908
#define GL_BGR 0x80E0

/* Pixel types */
#define GL_UNSIGNED_BYTE 0x1401
#define GL_FLOAT 0x1406

/* Pixel store */
#define GL_UNPACK_ALIGNMENT 0x0CF5

/* Texture units */
#define GL_TEXTURE0 0x84C0
#define GL_TEXTURE1 0x84C1

/* Type aliases */
typedef uint32_t GLenum;
typedef uint32_t GLuint;
typedef int32_t GLint;
typedef int32_t GLsizei;
typedef float GLfloat;
typedef uint8_t GLboolean;
typedef int8_t GLbyte;
typedef int32_t GLbitfield;
typedef char GLchar;
typedef float GLclampf;
typedef double GLdouble;
typedef int32_t GLfixed;
typedef int32_t GLintptr;
typedef int32_t GLsizeiptr;

/* Function pointer types */
typedef void (GL_APIENTRY *PFNGLVIEWPORTPROC)(GLint x, GLint y, GLsizei width, GLsizei height);
typedef void (GL_APIENTRY *PFNGLENABLEPROC)(GLenum cap);
typedef void (GL_APIENTRY *PFNGLDISABLEPROC)(GLenum cap);
typedef void (GL_APIENTRY *PFNGLCLEARCOLORPROC)(GLfloat red, GLfloat green, GLfloat blue, GLfloat alpha);
typedef void (GL_APIENTRY *PFNGLCLEARPROC)(GLbitfield mask);
typedef void (GL_APIENTRY *PFNGLPIXELSTOREIPROC)(GLenum pname, GLint param);
typedef void (GL_APIENTRY *PFNGLGENVERTEXARRAYSPROC)(GLsizei n, GLuint *arrays);
typedef void (GL_APIENTRY *PFNGLBINDVERTEXARRAYPROC)(GLuint array);
typedef void (GL_APIENTRY *PFNGLDELETEVERTEXARRAYSPROC)(GLsizei n, const GLuint *arrays);
typedef void (GL_APIENTRY *PFNGLGENBUFFERSPROC)(GLsizei n, GLuint *buffers);
typedef void (GL_APIENTRY *PFNGLBINDBUFFERPROC)(GLenum target, GLuint buffer);
typedef void (GL_APIENTRY *PFNGLDELETEBUFFERSPROC)(GLsizei n, const GLuint *buffers);
typedef void (GL_APIENTRY *PFNGLBUFFERDATAPROC)(GLenum target, GLsizeiptr size, const void *data, GLenum usage);
typedef void (GL_APIENTRY *PFNGLBUFFERSUBDATAPROC)(GLenum target, GLintptr offset, GLsizeiptr size, const void *data);
typedef void (GL_APIENTRY *PFNGLENABLEVERTEXATTRIBARRAYPROC)(GLuint index);
typedef void (GL_APIENTRY *PFNGLVERTEXATTRIBPOINTERPROC)(GLuint index, GLint size, GLenum type, GLboolean normalized, GLsizei stride, const void *pointer);
typedef GLint (GL_APIENTRY *PFNGLGETATTRIBLOCATIONPROC)(GLuint program, const GLchar *name);
typedef void (GL_APIENTRY *PFNGLUNIFORM1IPROC)(GLint location, GLint v0);
typedef void (GL_APIENTRY *PFNGLUNIFORM1FPROC)(GLint location, GLfloat v0);
typedef void (GL_APIENTRY *PFNGLUNIFORM2FPROC)(GLint location, GLfloat v0, GLfloat v1);
typedef void (GL_APIENTRY *PFNGLUNIFORM3FPROC)(GLint location, GLfloat v0, GLfloat v1, GLfloat v2);
typedef void (GL_APIENTRY *PFNGLUNIFORM4FPROC)(GLint location, GLfloat v0, GLfloat v1, GLfloat v2, GLfloat v3);
typedef void (GL_APIENTRY *PFNGLUNIFORMMATRIX4FVPROC)(GLint location, GLsizei count, GLboolean transpose, const GLfloat *value);
typedef GLint (GL_APIENTRY *PFNGLGETUNIFORMLOCATIONPROC)(GLuint program, const GLchar *name);
typedef GLuint (GL_APIENTRY *PFNGLCREATEPROGRAMPROC)(void);
typedef void (GL_APIENTRY *PFNGLDELETEPROGRAMPROC)(GLuint program);
typedef void (GL_APIENTRY *PFNGLUSEPROGRAMPROC)(GLuint program);
typedef void (GL_APIENTRY *PFNGLATTACHSHADERPROC)(GLuint program, GLuint shader);
typedef void (GL_APIENTRY *PFNGLLINKPROGRAMPROC)(GLuint program);
typedef void (GL_APIENTRY *PFNGLGETPROGRAMIVPROC)(GLuint program, GLenum pname, GLint *params);
typedef void (GL_APIENTRY *PFNGLGETPROGRAMINFOLOGPROC)(GLuint program, GLsizei bufSize, GLsizei *length, GLchar *infoLog);
typedef GLuint (GL_APIENTRY *PFNGLCREATESHADERPROC)(GLenum type);
typedef void (GL_APIENTRY *PFNGLDELETESHADERPROC)(GLuint shader);
typedef void (GL_APIENTRY *PFNGLSHADERSOURCEPROC)(GLuint shader, GLsizei count, const GLchar *const*string, const GLint *length);
typedef void (GL_APIENTRY *PFNGLCOMPILESHADERPROC)(GLuint shader);
typedef void (GL_APIENTRY *PFNGLGETSHADERIVPROC)(GLuint shader, GLenum pname, GLint *params);
typedef void (GL_APIENTRY *PFNGLGETSHADERINFOLOGPROC)(GLuint shader, GLsizei bufSize, GLsizei *length, GLchar *infoLog);
typedef void (GL_APIENTRY *PFNGLACTIVETEXTUREPROC)(GLenum texture);
typedef void (GL_APIENTRY *PFNGLGENTEXTURESPROC)(GLsizei n, GLuint *textures);
typedef void (GL_APIENTRY *PFNGLBINDTEXTUREPROC)(GLenum target, GLuint texture);
typedef void (GL_APIENTRY *PFNGLDELETETEXTURESPROC)(GLsizei n, const GLuint *textures);
typedef void (GL_APIENTRY *PFNGLTEXIMAGE2DPROC)(GLenum target, GLint level, GLint internalformat, GLsizei width, GLsizei height, GLint border, GLenum format, GLenum type, const void *pixels);
typedef void (GL_APIENTRY *PFNGLTEXPARAMETERIPROC)(GLenum target, GLenum pname, GLint param);
typedef void (GL_APIENTRY *PFNGLBLENDEQUATIONPROC)(GLenum mode);
typedef void (GL_APIENTRY *PFNGLBLENDFUNCPROC)(GLenum sfactor, GLenum dfactor);
typedef void (GL_APIENTRY *PFNGLDRAWARRAYSPROC)(GLenum mode, GLint first, GLsizei count);
typedef void (GL_APIENTRY *PFNGLDRAWELEMENTSPROC)(GLenum mode, GLsizei count, GLenum type, const void *indices);
typedef void (GL_APIENTRY *PFNGLGENFRAMEBUFFERSPROC)(GLsizei n, GLuint *framebuffers);
typedef void (GL_APIENTRY *PFNGLBINDFRAMEBUFFERPROC)(GLenum target, GLuint framebuffer);
typedef void (GL_APIENTRY *PFNGLFRAMEBUFFERTEXTURE2DPROC)(GLenum target, GLenum attachment, GLenum textarget, GLuint texture, GLint level);
typedef void (GL_APIENTRY *PFNGLSCISSORPROC)(GLint x, GLint y, GLsizei width, GLsizei height);
typedef void (GL_APIENTRY *PFNGLGETINTEGERVPROC)(GLenum pname, GLint *data);
typedef const GLubyte* (GL_APIENTRY *PFNGLGETSTRINGPROC)(GLenum name);

#ifndef GL_APIENTRY
#define GL_APIENTRY
#endif

/* Exported function pointers */
extern PFNGLVIEWPORTPROC glad_glViewport;
extern PFNGLENABLEPROC glad_glEnable;
extern PFNGLDISABLEPROC glad_glDisable;
extern PFNGLCLEARCOLORPROC glad_glClearColor;
extern PFNGLCLEARPROC glad_glClear;
extern PFNGLPIXELSTOREIPROC glad_glPixelStorei;
extern PFNGLGENVERTEXARRAYSPROC glad_glGenVertexArrays;
extern PFNGLBINDVERTEXARRAYPROC glad_glBindVertexArray;
extern PFNGLDELETEVERTEXARRAYSPROC glad_glDeleteVertexArrays;
extern PFNGLGENBUFFERSPROC glad_glGenBuffers;
extern PFNGLBINDBUFFERPROC glad_glBindBuffer;
extern PFNGLDELETEBUFFERSPROC glad_glDeleteBuffers;
extern PFNGLBUFFERDATAPROC glad_glBufferData;
extern PFNGLBUFFERSUBDATAPROC glad_glBufferSubData;
extern PFNGLENABLEVERTEXATTRIBARRAYPROC glad_glEnableVertexAttribArray;
extern PFNGLVERTEXATTRIBPOINTERPROC glad_glVertexAttribPointer;
extern PFNGLGETATTRIBLOCATIONPROC glad_glGetAttribLocation;
extern PFNGLUNIFORM1IPROC glad_glUniform1i;
extern PFNGLUNIFORM1FPROC glad_glUniform1f;
extern PFNGLUNIFORM2FPROC glad_glUniform2f;
extern PFNGLUNIFORM3FPROC glad_glUniform3f;
extern PFNGLUNIFORM4FPROC glad_glUniform4f;
extern PFNGLUNIFORMMATRIX4FVPROC glad_glUniformMatrix4fv;
extern PFNGLGETUNIFORMLOCATIONPROC glad_glGetUniformLocation;
extern PFNGLCREATEPROGRAMPROC glad_glCreateProgram;
extern PFNGLDELETEPROGRAMPROC glad_glDeleteProgram;
extern PFNGLUSEPROGRAMPROC glad_glUseProgram;
extern PFNGLATTACHSHADERPROC glad_glAttachShader;
extern PFNGLLINKPROGRAMPROC glad_glLinkProgram;
extern PFNGLGETPROGRAMIVPROC glad_glGetProgramiv;
extern PFNGLGETPROGRAMINFOLOGPROC glad_glGetProgramInfoLog;
extern PFNGLCREATESHADERPROC glad_glCreateShader;
extern PFNGLDELETESHADERPROC glad_glDeleteShader;
extern PFNGLSHADERSOURCEPROC glad_glShaderSource;
extern PFNGLCOMPILESHADERPROC glad_glCompileShader;
extern PFNGLGETSHADERIVPROC glad_glGetShaderiv;
extern PFNGLGETSHADERINFOLOGPROC glad_glGetShaderInfoLog;
extern PFNGLACTIVETEXTUREPROC glad_glActiveTexture;
extern PFNGLGENTEXTURESPROC glad_glGenTextures;
extern PFNGLBINDTEXTUREPROC glad_glBindTexture;
extern PFNGLDELETETEXTURESPROC glad_glDeleteTextures;
extern PFNGLTEXIMAGE2DPROC glad_glTexImage2D;
extern PFNGLTEXPARAMETERIPROC glad_glTexParameteri;
extern PFNGLBLENDEQUATIONPROC glad_glBlendEquation;
extern PFNGLBLENDFUNCPROC glad_glBlendFunc;
extern PFNGLDRAWARRAYSPROC glad_glDrawArrays;
extern PFNGLDRAWELEMENTSPROC glad_glDrawElements;
extern PFNGLGENFRAMEBUFFERSPROC glad_glGenFramebuffers;
extern PFNGLBINDFRAMEBUFFERPROC glad_glBindFramebuffer;
extern PFNGLFRAMEBUFFERTEXTURE2DPROC glad_glFramebufferTexture2D;
extern PFNGLSCISSORPROC glad_glScissor;
extern PFNGLGETINTEGERVPROC glad_glGetIntegerv;
extern PFNGLGETSTRINGPROC glad_glGetString;

/* Convenience macros */
#define glViewport glad_glViewport
#define glEnable glad_glEnable
#define glDisable glad_glDisable
#define glClearColor glad_glClearColor
#define glClear glad_glClear
#define glPixelStorei glad_glPixelStorei
#define glGenVertexArrays glad_glGenVertexArrays
#define glBindVertexArray glad_glBindVertexArray
#define glDeleteVertexArrays glad_glDeleteVertexArrays
#define glGenBuffers glad_glGenBuffers
#define glBindBuffer glad_glBindBuffer
#define glDeleteBuffers glad_glDeleteBuffers
#define glBufferData glad_glBufferData
#define glBufferSubData glad_glBufferSubData
#define glEnableVertexAttribArray glad_glEnableVertexAttribArray
#define glVertexAttribPointer glad_glVertexAttribPointer
#define glGetAttribLocation glad_glGetAttribLocation
#define glUniform1i glad_glUniform1i
#define glUniform1f glad_glUniform1f
#define glUniform2f glad_glUniform2f
#define glUniform3f glad_glUniform3f
#define glUniform4f glad_glUniform4f
#define glUniformMatrix4fv glad_glUniformMatrix4fv
#define glGetUniformLocation glad_glGetUniformLocation
#define glCreateProgram glad_glCreateProgram
#define glDeleteProgram glad_glDeleteProgram
#define glUseProgram glad_glUseProgram
#define glAttachShader glad_glAttachShader
#define glLinkProgram glad_glLinkProgram
#define glGetProgramiv glad_glGetProgramiv
#define glGetProgramInfoLog glad_glGetProgramInfoLog
#define glCreateShader glad_glCreateShader
#define glDeleteShader glad_glDeleteShader
#define glShaderSource glad_glShaderSource
#define glCompileShader glad_glCompileShader
#define glGetShaderiv glad_glGetShaderiv
#define glGetShaderInfoLog glad_glGetShaderInfoLog
#define glActiveTexture glad_glActiveTexture
#define glGenTextures glad_glGenTextures
#define glBindTexture glad_glBindTexture
#define glDeleteTextures glad_glDeleteTextures
#define glTexImage2D glad_glTexImage2D
#define glTexParameteri glad_glTexParameteri
#define glBlendEquation glad_glBlendEquation
#define glBlendFunc glad_glBlendFunc
#define glDrawArrays glad_glDrawArrays
#define glDrawElements glad_glDrawElements
#define glGenFramebuffers glad_glGenFramebuffers
#define glBindFramebuffer glad_glBindFramebuffer
#define glFramebufferTexture2D glad_glFramebufferTexture2D
#define glScissor glad_glScissor
#define glGetIntegerv glad_glGetIntegerv
#define glGetString glad_glGetString

/* Initialization */
int gladLoadGL(void *(*getProcAddr)(const char *name));

#ifdef __cplusplus
}
#endif

#endif /* GLAD_H_ */
