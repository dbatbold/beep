#include <soundio/soundio.h>

char *init_sound_device(char *);
char *open_sound_device(void);
void playback(short *buf1, short *buf2, int bufSize);
char *open_stream();
void stop_playback();
void flush_sound_buffer();
void close_sound_device();

static void write_callback(struct SoundIoOutStream *outstream,
	int frame_count_min, int frame_count_max);
static void underflow_callback(struct SoundIoOutStream *outstream);
static void write_sample_s16le(char *ptr, double sample);
