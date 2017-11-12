#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <math.h>

#include "beep_soundio.h"

double latency = 0.2;
int sample_rate = 44100;

static void (*write_sample)(char *ptr, double sample);
static char buf[1024];
static short *sound_buf;
static int sound_frame;
static int sound_buf_len;

char *open_sound_device(void) {
	int i = soundio_default_output_device_index(soundio);
	if (i < 0) {
		snprintf(buf, sizeof buf, "output device not found");
		return buf;
	}
	
	soundioDev = soundio_get_output_device(soundio, i);
	if (!soundioDev)
		return "unable to get output device";

	//fprintf(stderr, "Output device: %s\n", soundioDev->name);

	if (soundioDev->probe_error) {
		snprintf(buf, sizeof buf,
			"Cannot probe device: %s\n", soundio_strerror(soundioDev->probe_error));
		return buf;
	}

	return 0;
}

char *open_stream() {
	outStream = soundio_outstream_create(soundioDev);
	if (!outStream)
		return "unable to get create output stream";

	outStream->write_callback = write_callback;
	outStream->underflow_callback = underflow_callback;
	outStream->name = "beep";
	outStream->software_latency = latency;
	outStream->sample_rate = sample_rate;
	outStream->format = SoundIoFormatS16LE;

	write_sample = write_sample_s16le;

	int err = soundio_outstream_open(outStream);
	if (err) {
		snprintf(buf, sizeof buf,
			"unalbe to open output stream: %s", soundio_strerror(err));
	}
}

char *init_sound_device(char *os) {
	soundio = soundio_create();
	if (!soundio)
		return "failed to initialize sound device";

	enum SoundIoBackend backend = SoundIoBackendNone;
	if (!strcmp(os, "linux"))
		backend = SoundIoBackendPulseAudio; // SoundIoBackendAlsa
	else if (!strcmp(os, "darwin"))
		backend = SoundIoBackendCoreAudio;
	else if (!strcmp(os, "windows"))
		backend = SoundIoBackendWasapi;

	int err = soundio_connect_backend(soundio, backend);
	if (err) {
		snprintf(buf, sizeof buf,
			"unalbe to connect to backend: %s", soundio_strerror(err));
		return buf;
	}

	soundio_flush_events(soundio);

	return 0;
}

static void write_callback(struct SoundIoOutStream *outstream,
	int frame_count_min, int frame_count_max) {

    double float_sample_rate = outstream->sample_rate;
    struct SoundIoChannelArea *areas;
    int err;

	int frames_left = frame_count_max;
	if (sound_buf_len - sound_frame < frames_left)
		frames_left = sound_buf_len - sound_frame;

	for (;;) {
		//printf("frames_left=%i\n", frames_left);
		if (frames_left <= 0) {
			break;
		}

		int frame_count = frames_left;
		//printf("frame_count=%i\n", frame_count);
		if ((err = soundio_outstream_begin_write(outstream, &areas, &frame_count))) {
			fprintf(stderr, "unrecoverable stream error: %s\n", soundio_strerror(err));
			exit(1);
		}

		const struct SoundIoChannelLayout *layout = &outstream->layout;

		for (int frame = 0; frame < frame_count; frame++) {
			double sample = sound_buf[sound_frame];
			for (int ch = 0; ch < layout->channel_count; ch++) {
				write_sample(areas[ch].ptr, sample);
				areas[ch].ptr += areas[ch].step;
			}
			sound_frame++;
		}

		if ((err = soundio_outstream_end_write(outstream))) {
			if (err == SoundIoErrorUnderflow)
				return;
			fprintf(stderr, "unrecoverable stream error: %s\n", soundio_strerror(err));
			exit(1);
			break;
		}

		//printf("frame_count=%i, frames_left=%i\n", frame_count, frames_left);

		frames_left -= frame_count;
		if (frames_left <= 0) {
			break;
		}
	}
}

static void underflow_callback(struct SoundIoOutStream *outstream) {
	//printf("underflow_callback\n");
}

void playback(short *buf1, short *buf2, int buf_size) {
	sound_buf = buf1;
	sound_frame = 0;
	sound_buf_len = buf_size;

	if (!outStream) {
		char *s = open_stream();
		if (s) {
			fprintf(stderr, "failed to open stream: %s\n", s);
			return;
		}

		int err = soundio_outstream_start(outStream);
		if (err) {
			fprintf(stderr, "failed to start stream: %s\n", soundio_strerror(err));
			return;
		}
	}
}

void stop_playback() {
	soundio_outstream_destroy(outStream);
	outStream = 0;
	//printf("stop_playback\n");
}

static void write_sample_s16le(char *ptr, double sample) {
    int16_t *buf = (int16_t *)ptr;
    *buf = sample;
}
