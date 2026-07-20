console.log("Hello This is a Javascript project.");

function visualize(stream, canvas) {

    const ctx = canvas.getContext("2d");

    const audioCtx = new AudioContext();

    const analyser = audioCtx.createAnalyser();

    analyser.fftSize = 256;

    const source = audioCtx.createMediaStreamSource(stream);

    source.connect(analyser);

    const data = new Uint8Array(analyser.frequencyBinCount);

    function draw() {

        requestAnimationFrame(draw);

        analyser.getByteFrequencyData(data);

        ctx.fillStyle = "#111";
        ctx.fillRect(0, 0, canvas.width, canvas.height);

        const barWidth = canvas.width / data.length / 2;

        for (let i = data.length; i >= 0; i--) {
            const h = data[i] / 255 * canvas.height;

            ctx.fillStyle = "lime";

            ctx.fillRect(
                (data.length - i) * barWidth,
                canvas.height / 2 - h / 2,
                barWidth - 1,
                h
            );
        }

        for (let i = 0; i < data.length; i++) {
            const h = data[i] / 255 * canvas.height;

            ctx.fillStyle = "lime";

            ctx.fillRect(
                (data.length + i) * barWidth,
                canvas.height / 2 - h / 2,
                barWidth - 1,
                h
            );
        }

    }

    draw();

}

const pc = new RTCPeerConnection();

const stream = await navigator.mediaDevices.getUserMedia({
    audio: true,
    video: false
});

stream.getTracks().forEach(track => {
    pc.addTrack(track, stream);
});

const offer = await pc.createOffer();
await pc.setLocalDescription(offer);

await new Promise(resolve => {
    if (pc.iceGatheringState === "complete") {
        resolve();
    } else {
        pc.addEventListener("icegatheringstatechange", () => {
            if (pc.iceGatheringState === "complete")
                resolve();
        });
    }
});

pc.ontrack = (event) => {
    console.log('remote track received')
    const stream = event.streams[0];
    const remoteAudio = document.getElementById("remoteAudio");
    remoteAudio.srcObject = stream;
    visualize(stream, document.getElementById("remoteCanvas"));
};

const response = await fetch("http://127.0.0.1:8080/offer", {
    method: "POST",
    headers: {
        "Content-Type": "application/json"
    },
    body: JSON.stringify(pc.localDescription)
});

const answer = await response.json();

await pc.setRemoteDescription(answer);
