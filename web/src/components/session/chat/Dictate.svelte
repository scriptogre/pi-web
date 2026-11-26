<script>
  import { icon, Check, Mic, X } from '../../../shared/icons.js';

  // idle → starting → listening → finishing → idle. ✕ and errors return to idle from any state.
  let state = $state('idle');
  let bars = $state(Array(28).fill(0));
  let session = null; // { box, before, after, original, finals, interim, audio, stream, socket }

  // Streams text in at the cursor. Text outside the cursor range never changes.
  function render() {
    const { box, before, after, finals, interim } = session;
    const words = [...finals, interim].filter(Boolean).join(' ');
    const head = before + (words && before && !/\s$/.test(before) ? ' ' : '') + words;
    box.value = head + (words && after && !/^\s/.test(after) ? ' ' : '') + after;
    box.selectionStart = box.selectionEnd = head.length;
    box.dispatchEvent(new Event('input', { bubbles: true }));
  }

  async function start() {
    const box = document.getElementById('pi-chat-message');
    // iOS only allows audio that starts inside the tap, so create the context before any await.
    const audio = new AudioContext();
    const current = (session = {
      box,
      before: box.value.slice(0, box.selectionStart),
      after: box.value.slice(box.selectionEnd),
      original: box.value,
      finals: [],
      interim: '',
      audio,
    });
    state = 'starting';
    box.readOnly = true;
    try {
      current.stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      const response = await fetch('/api/dictate');
      const grant = await response.json();
      if (!response.ok) throw new Error(grant.error || response.statusText);
      if (session !== current) return finish(current); // ✕ was tapped while starting

      const query = new URLSearchParams({
        model: 'nova-3',
        language: grant.language,
        encoding: 'linear16',
        sample_rate: audio.sampleRate,
        channels: 1,
        smart_format: true,
        interim_results: true,
      });
      const socket = (current.socket = new WebSocket(`wss://api.deepgram.com/v1/listen?${query}`, ['bearer', grant.token]));
      socket.onopen = () => state === 'starting' && (state = 'listening');
      socket.onclose = ({ reason }) => {
        if (state === 'starting') alert(`Dictation failed: ${reason || 'Deepgram closed the connection'}`);
        finish(current, true);
      };
      socket.onmessage = ({ data }) => {
        const message = JSON.parse(data);
        const text = message.channel?.alternatives[0]?.transcript;
        if (text === undefined || session !== current) return;
        current.interim = message.is_final ? '' : text;
        if (message.is_final && text) current.finals.push(text);
        render();
      };

      // shortcut: ScriptProcessor is deprecated but runs everywhere; move to an AudioWorklet if Safari drops it.
      const processor = audio.createScriptProcessor(4096, 1, 1);
      processor.onaudioprocess = ({ inputBuffer }) => {
        const samples = inputBuffer.getChannelData(0);
        const level = Math.sqrt(samples.reduce((sum, x) => sum + x * x, 0) / samples.length);
        bars = [...bars.slice(1), Math.min(1, level * 6)];
        if (state === 'listening' && socket.readyState === WebSocket.OPEN)
          socket.send(Int16Array.from(samples, (x) => Math.max(-1, Math.min(1, x)) * 0x7fff).buffer);
      };
      audio.createMediaStreamSource(current.stream).connect(processor);
      processor.connect(audio.destination);
    } catch (error) {
      finish(current, false);
      alert(`Dictation failed: ${error.message}`);
    }
  }

  // ✓: Deepgram flushes the last words, then closes the socket, which calls finish().
  function keep() {
    const current = session;
    state = 'finishing';
    current.socket.send(JSON.stringify({ type: 'CloseStream' }));
    setTimeout(() => finish(current, true), 3000);
  }

  // Every exit path releases the mic, so the iOS orange dot always goes away.
  function release({ stream, audio, socket }) {
    stream?.getTracks().forEach((track) => track.stop());
    if (audio.state !== 'closed') audio.close();
    socket?.close();
  }

  function finish(current, kept) {
    release(current);
    if (session !== current) return;
    session = null;
    current.box.readOnly = false;
    if (!kept) {
      current.box.value = current.original;
      current.box.dispatchEvent(new Event('input', { bubbles: true }));
    }
    state = 'idle';
    bars = bars.map(() => 0);
  }
</script>

<!-- eslint-disable svelte/no-at-html-tags -- trusted: Lucide icon SVG -->
{#if state === 'idle'}
  <button type="button" class="pi-dictate-mic" aria-label="Dictate" onclick={start}>{@html icon(Mic, { size: 18 })}</button>
{:else}
  <div class="pi-dictate">
    <button type="button" aria-label="Discard dictation" onclick={() => finish(session, false)}
      >{@html icon(X, { size: 18 })}</button
    >
    <div class="pi-dictate-wave">
      {#each bars as bar, i (i)}<span style:height="{3 + bar * 21}px"></span>{/each}
    </div>
    <button type="button" class="pi-dictate-keep" aria-label="Keep dictation" disabled={state !== 'listening'} onclick={keep}
      >{@html icon(Check, { size: 18 })}</button
    >
  </div>
{/if}

