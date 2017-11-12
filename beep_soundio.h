#include <soundio/soundio.h>

struct SoundIo *soundio;
struct SoundIoDevice *soundioDev;
struct SoundIoOutStream *outStream;

char *init_sound_device(char *);
char *open_sound_device(void);
void playback(short *buf1, short *buf2, int bufSize);
char *open_stream();
void stop_playback();

static void write_callback(struct SoundIoOutStream *outstream,
	int frame_count_min, int frame_count_max);
static void underflow_callback(struct SoundIoOutStream *outstream);
static void write_sample_s16le(char *ptr, double sample);
