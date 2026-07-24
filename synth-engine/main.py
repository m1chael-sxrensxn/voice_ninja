import asyncio
import supriya
from supriya import SynthDefBuilder
from supriya.ugens import (
    Saw,
    SinOsc,
    HPF,
    LinLin,
    LPF,
    RLPF,
    BPeakEQ,
    DelayC,
    DelayN,
    FreeVerb,
    Envelope,
    EnvGen,
    Out,
    XLine,
)


def make_gabberkick():

    builder = SynthDefBuilder(
        freq=440,
        bend=1,
        ffreq=3000,
        gate=1,
        amp=0.1,
        out=0,
    )

    with builder:

        pitch_envelope = Envelope(
            amplitudes=[0, 1, 0],
            durations=[0.001, 0.08],
            curves=[-1, -1],
        )

        pitch_env = EnvGen.kr(
            envelope=pitch_envelope
        )

        freq_mod = (
            pitch_env
            * 48
            * builder["bend"]
        )

        # midiratio conversion:
        freq = (
            builder["freq"]
            * (2 ** (freq_mod / 12))
        )


        snd = Saw.ar(frequency=freq)

        snd = (
            (snd * 100).tanh()
            +
            ((snd.sign() - snd) * (10 ** (-8 / 20)))
        )


        high = HPF.ar(
            source=snd,
            frequency=300
        )


        lfo = LinLin.ar(
            source=SinOsc.ar(frequency=8),
            output_minimum=-1,
            output_maximum=1,
        )


        high = (
            high
            +
            DelayC.ar(
                source=high,
                maximum_delay_time=0.01,
                delay_time=lfo
            )
            *
            (10 ** (-2 / 20))
        )


        snd = (
            LPF.ar(
                source=snd,
                frequency=100
            )
            +
            high
        )


        snd = RLPF.ar(
            source=snd,
            frequency=7000,
            reciprocal_of_q=2
        )


        snd = BPeakEQ.ar(
            source=snd,
            frequency=builder["ffreq"]
            * XLine.kr(
                start=1,
                stop=0.8,
                duration=0.3
            ),
            reciprocal_of_q=0.5,
            gain=15
        )


        snd = (
            snd
            *
            EnvGen.kr(
                envelope=Envelope.asr(
                    0.001,
                    1,
                    0.05
                ),
                gate=builder["gate"]
            )
        )


        Out.ar(
            bus=builder["out"],
            source=snd * builder["amp"]
        )


    return builder.build(name="gabberkick")

def make_hoover():

    builder = SynthDefBuilder(
        freq=440,
        gate=1,
        amp=0.1,
        out=0,
    )

    with builder:

        freq_env = EnvGen.kr(
            envelope=Envelope(
                amplitudes=[-5, 6, 0],
                durations=[0.1, 1.7],
            )
        )

        freq = (
            builder["freq"]
            *
            (2 ** (freq_env / 12))
        )


        voices = []

        for i in range(20):

            voice = (
                Saw.ar(frequency=freq)
                +
                Saw.ar(frequency=freq * 0.5)
            )

            voice = DelayN.ar(
                source=voice,
                maximum_delay_time=0.01,
                delay_time=0.005
            )

            voices.append(voice)


        snd = sum(voices) / len(voices)


        snd = snd.atan() * 3


        snd = (
            snd
            *
            EnvGen.kr(
                envelope=Envelope.asr(
                    0.01,
                    1,
                    1
                ),
                gate=builder["gate"]
            )
        )


        snd = FreeVerb.ar(
            source=snd,
            mix=0.3,
            room_size=0.9
        )


        Out.ar(
            bus=builder["out"],
            source=snd * builder["amp"]
        )


    return builder.build(name="hoover")

durations = [
    1,
    1,
    1,
    1,
    3/4,
    1/4,
    1/2,
    3/4,
    1/4,
    1/2,
]
seconds_per_beat = 60 / 210

def dbamp(db):
    return 10 ** (db / 20)

def midi_to_freq(note):
    return 440 * (
        2 ** ((note - 69) / 12)
    )

def get_bend(duration):
    if duration < 0.5:
        return 0.4

    return 1

def exponential_sweep(index, total, low=100, high=4000):
    x = index / (total - 1)
    return low * ((high / low) ** x)

async def play_kicks(server, synthdef):
    total_steps = len(durations) * 4
    step = 0

    while True:
        for duration in durations:
            ffreq = exponential_sweep(
                step,
                total_steps,
                low=100,
                high=4000,
            )

            node = server.add_synth(
                synthdef,
                freq=60,
                amp=dbamp(-23),
                ffreq=ffreq,
                bend=get_bend(duration),
                gate=1,
            )

            step = (step + 1) % total_steps

            await asyncio.sleep(
                duration * seconds_per_beat
            )
            node.set(
                gate=0
            )

async def play_hoover(server, synthdef):
    node = server.add_synth(
        synthdef,
        freq=midi_to_freq(74),
        amp=dbamp(-20),
        gate=1,
    )

    await asyncio.sleep(7)

    node.set(
        gate=0
    )

async def main() -> None:
    """
    The example entry-point function.
    """
    # Create a server and boot it:
    server = supriya.Server().boot()

    # create synths on the server
    gabberkick = make_gabberkick()
    hoover = make_hoover()
    server.add_synthdefs(gabberkick)
    server.add_synthdefs(hoover)
    server.sync()

    # Play the example rave track
    await asyncio.gather(
        play_kicks(server, gabberkick),
        play_hoover(server, hoover),
    )

    # Quit the server:
    server.quit()
    print("done")

asyncio.run(main())
