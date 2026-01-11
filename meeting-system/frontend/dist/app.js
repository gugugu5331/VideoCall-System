/* global RTCPeerConnection, RTCSessionDescription, RTCIceCandidate */

const WS_TYPES = {
  OFFER: 1,
  ANSWER: 2,
  ICE: 3,
  JOIN: 4,
  LEAVE: 5,
  USER_JOINED: 6,
  USER_LEFT: 7,
  CHAT: 8,
  MEDIA_CONTROL: 10,
  PING: 11,
  PONG: 12,
  ERROR: 13,
  ROOM_INFO: 14,
  AI_LIVE_CLAIM: 15,
  AI_LIVE_STATUS: 16,
  AI_LIVE_RESULT: 17,
};

const els = {
  secureBadge: document.getElementById("secureBadge"),
  insecureWarning: document.getElementById("insecureWarning"),
  authCard: document.getElementById("authCard"),
  mainCard: document.getElementById("mainCard"),
  authMsg: document.getElementById("authMsg"),

  tabLogin: document.getElementById("tabLogin"),
  tabRegister: document.getElementById("tabRegister"),
  loginForm: document.getElementById("loginForm"),
  registerForm: document.getElementById("registerForm"),
  logoutBtn: document.getElementById("logoutBtn"),

  loginUsername: document.getElementById("loginUsername"),
  loginPassword: document.getElementById("loginPassword"),
  regUsername: document.getElementById("regUsername"),
  regNickname: document.getElementById("regNickname"),
  regEmail: document.getElementById("regEmail"),
  regPassword: document.getElementById("regPassword"),

  currentUser: document.getElementById("currentUser"),
  currentMeeting: document.getElementById("currentMeeting"),

  createTitle: document.getElementById("createTitle"),
  createType: document.getElementById("createType"),
  createBtn: document.getElementById("createBtn"),
  createResult: document.getElementById("createResult"),

  meetingIdInput: document.getElementById("meetingIdInput"),
  joinBtn: document.getElementById("joinBtn"),
  joinResult: document.getElementById("joinResult"),

  startMediaBtn: document.getElementById("startMediaBtn"),
  leaveBtn: document.getElementById("leaveBtn"),
  muteBtn: document.getElementById("muteBtn"),
  videoBtn: document.getElementById("videoBtn"),
  screenBtn: document.getElementById("screenBtn"),
  wsState: document.getElementById("wsState"),
  iceState: document.getElementById("iceState"),

  participants: document.getElementById("participants"),
  videoGrid: document.getElementById("videoGrid"),
  fxEnableBeauty: document.getElementById("fxEnableBeauty"),
  fxEnableFilter: document.getElementById("fxEnableFilter"),
  fxBeautyType: document.getElementById("fxBeautyType"),
  fxBeauty: document.getElementById("fxBeauty"),
  fxBeautyVal: document.getElementById("fxBeautyVal"),
  fxSlim: document.getElementById("fxSlim"),
  fxSlimVal: document.getElementById("fxSlimVal"),
  fxFilter: document.getElementById("fxFilter"),
  seiStatus: document.getElementById("seiStatus"),
  aiDanmaku: document.getElementById("aiDanmaku"),

  chatLog: document.getElementById("chatLog"),
  chatForm: document.getElementById("chatForm"),
  chatInput: document.getElementById("chatInput"),

  aiHealthBtn: document.getElementById("aiHealthBtn"),
  aiInfoBtn: document.getElementById("aiInfoBtn"),
  aiClearBtn: document.getElementById("aiClearBtn"),
  aiStatusPill: document.getElementById("aiStatusPill"),
  aiStatusText: document.getElementById("aiStatusText"),
  aiLiveToggleBtn: document.getElementById("aiLiveToggleBtn"),
  aiLiveClearBtn: document.getElementById("aiLiveClearBtn"),
  aiLiveAsr: document.getElementById("aiLiveAsr"),
  aiLiveEmotion: document.getElementById("aiLiveEmotion"),
  aiLiveSynth: document.getElementById("aiLiveSynth"),
  aiLiveSpeaker: document.getElementById("aiLiveSpeaker"),
  aiLiveLevel: document.getElementById("aiLiveLevel"),
  aiLiveLog: document.getElementById("aiLiveLog"),
  aiEmotionText: document.getElementById("aiEmotionText"),
  aiEmotionBtn: document.getElementById("aiEmotionBtn"),
  aiEmotionExampleBtn: document.getElementById("aiEmotionExampleBtn"),
  aiAsrFile: document.getElementById("aiAsrFile"),
  aiAsrFormat: document.getElementById("aiAsrFormat"),
  aiAsrSampleRate: document.getElementById("aiAsrSampleRate"),
  aiAsrLang: document.getElementById("aiAsrLang"),
  aiAsrBtn: document.getElementById("aiAsrBtn"),
  aiSynthFile: document.getElementById("aiSynthFile"),
  aiSynthFormat: document.getElementById("aiSynthFormat"),
  aiSynthSampleRate: document.getElementById("aiSynthSampleRate"),
  aiSynthBtn: document.getElementById("aiSynthBtn"),
  aiOutput: document.getElementById("aiOutput"),
  aiPanelBody: document.getElementById("aiPanelBody"),
  fxPanelBody: document.getElementById("fxPanelBody"),
  fxPanelContent: document.getElementById("fxPanelContent"),
  toggleFxPanel: document.getElementById("toggleFxPanel"),
  openPanelDrawer: document.getElementById("openPanelDrawer"),
  openFxDrawer: document.getElementById("openFxDrawer"),
  openAiDrawer: document.getElementById("openAiDrawer"),
  closePanelDrawer: document.getElementById("closePanelDrawer"),
  closeFxDrawer: document.getElementById("closeFxDrawer"),
  closeAiDrawer: document.getElementById("closeAiDrawer"),
  panelDrawer: document.getElementById("panelDrawer"),
  fxDrawer: document.getElementById("fxDrawer"),
  aiDrawer: document.getElementById("aiDrawer"),
};

const state = {
  csrfToken: null,
  token: null,
  user: null,
  meetingId: null,
  roomId: null,
  peerId: null,
  sessionId: null,
  ws: null,
  wsReady: false,
  wsCloseExpected: false,
  roomIceServers: [],
  localStream: null,
  screenStream: null,
  sfuPc: null,
  sfuPeerId: null,
  sfuAudioTransceiver: null,
  sfuVideoTransceiver: null,
  sfuOfferPoll: null,
  sfuIcePoll: null,
  sfuOfferInFlight: false,
  sfuIceInFlight: false,
  sfuReconnectTimer: null,
  sfuLastReconnectAt: 0,
  sfuPendingLocalCandidates: [],
  sfuLastOfferSdp: null,
  remoteStreams: new Map(), // streamId -> MediaStream
  remoteAudioEls: new Map(), // key -> HTMLAudioElement
  participantsByPeerId: new Map(),
  aiHealthChecked: false,
  aiLiveEnabled: false,
  aiLiveIsLeader: false,
  aiLiveStatus: null,
  aiLiveCaptureStarting: false,
  aiLiveClaimTimer: null,
  aiLiveLeadCapable: true,
  aiLiveInputs: new Map(), // speakerKey -> MediaStream (audio)
  aiLiveAnalyzers: new Map(), // speakerKey -> { stream, source, analyser, gain, lastRms }
  aiLiveLineEls: new Map(), // lineId -> { root, textEl, tagsEl }
  aiLiveQueue: [],
  aiLiveQueueRunning: false,
  aiAudioCtx: null,
  aiCaptureSourceKey: null,
  aiCaptureSource: null,
  aiCaptureProcessor: null,
  aiCaptureSink: null,
  aiCaptureChunks: [],
  aiCaptureRecording: false,
  aiCaptureStartedAt: 0,
  aiCaptureLastVoiceAt: 0,
  aiCaptureVoiceStreak: 0,
  aiCurrentSpeakerKey: null,
  aiLastHighlightedKey: null,
  aiLiveTimer: null,
  fxBeautyEnabled: true,
  fxFilterEnabled: false,
  fxBeautyType: "natural",
  fxBeauty: 20,
  fxSlim: 0,
  fxFilter: "none",
  fxVersion: 1,
  fxLastInjectedVersion: 0,
  fxLastInjectedAt: 0,
  fxSeiReady: false,
  fxSenderAttached: false,
  fxRemote: new Map(), // tileKey -> { beautyEnabled, filterEnabled, beautyType, beauty, filter, updatedAt }
  fxReceiverAttached: new Set(), // tileKey
  fxProcessingSupported: false,
  fxProcessedTrack: null,
  fxPipelineStop: null,
  fxPipelineSourceId: null,
  fxMirrorX: false,
  danmakuLane: 0,
};

function uuid() {
  if (globalThis.crypto?.randomUUID) return globalThis.crypto.randomUUID();
  return `${Date.now()}_${Math.random().toString(16).slice(2)}`;
}

function sleep(ms) {
  return new Promise((r) => setTimeout(r, ms));
}

const FX_SEI_UUID = "b0f7b0a1-6a3d-4c53-9b2e-6a7d3e9f1c20";
let fxSeiUuidBytes = null;

function uuidStringToBytes(uuidStr) {
  const hex = String(uuidStr || "").replaceAll("-", "");
  if (hex.length !== 32) throw new Error("invalid uuid");
  const out = new Uint8Array(16);
  for (let i = 0; i < 16; i++) {
    out[i] = Number.parseInt(hex.slice(i * 2, i * 2 + 2), 16);
  }
  return out;
}

function getFxSeiUuidBytes() {
  if (fxSeiUuidBytes) return fxSeiUuidBytes;
  fxSeiUuidBytes = uuidStringToBytes(FX_SEI_UUID);
  return fxSeiUuidBytes;
}

function bytesEqual(a, b) {
  if (!a || !b || a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) {
    if (a[i] !== b[i]) return false;
  }
  return true;
}

function utf8Encode(str) {
  try {
    if (globalThis.TextEncoder) return new TextEncoder().encode(String(str ?? ""));
  } catch {
    // ignore
  }
  const s = unescape(encodeURIComponent(String(str ?? "")));
  const out = new Uint8Array(s.length);
  for (let i = 0; i < s.length; i++) out[i] = s.charCodeAt(i);
  return out;
}

function utf8Decode(bytes) {
  try {
    if (globalThis.TextDecoder) return new TextDecoder().decode(bytes);
  } catch {
    // ignore
  }
  let s = "";
  for (let i = 0; i < bytes.length; i++) s += String.fromCharCode(bytes[i]);
  try {
    return decodeURIComponent(escape(s));
  } catch {
    return s;
  }
}

function canUseEncodedInsertableStreams() {
  try {
    const senderOk =
      typeof RTCRtpSender !== "undefined" && RTCRtpSender.prototype && typeof RTCRtpSender.prototype.createEncodedStreams === "function";
    const receiverOk =
      typeof RTCRtpReceiver !== "undefined" && RTCRtpReceiver.prototype && typeof RTCRtpReceiver.prototype.createEncodedStreams === "function";
    return senderOk && receiverOk;
  } catch {
    return false;
  }
}

function supportsFxProcessing() {
  return (
    typeof MediaStreamTrackProcessor === "function" &&
    typeof MediaStreamTrackGenerator === "function" &&
    typeof VideoFrame !== "undefined" &&
    (typeof OffscreenCanvas !== "undefined" || typeof document !== "undefined")
  );
}

function setSeiStatus(text, kind = "info") {
  if (!els.seiStatus) return;
  els.seiStatus.textContent = text || "-";
  els.seiStatus.style.color =
    kind === "error" ? "rgba(255,77,79,0.92)" : kind === "ok" ? "rgba(61,214,208,0.92)" : kind === "warn" ? "rgba(255,183,77,0.92)" : "";
}

function isSafari() {
  if (typeof navigator === "undefined") return false;
  const ua = navigator.userAgent || "";
  return /Safari/.test(ua) && !/Chrome|Chromium|Edg/.test(ua);
}

function isAndroid() {
  if (typeof navigator === "undefined") return false;
  return /Android/i.test(navigator.userAgent || "");
}

function updateFxUI() {
  if (els.fxEnableBeauty) els.fxEnableBeauty.checked = Boolean(state.fxBeautyEnabled);
  if (els.fxEnableFilter) els.fxEnableFilter.checked = Boolean(state.fxFilterEnabled);
  if (els.fxBeautyType && els.fxBeautyType.value !== state.fxBeautyType) els.fxBeautyType.value = state.fxBeautyType;
  if (els.fxBeauty && els.fxBeauty.value !== String(state.fxBeauty)) els.fxBeauty.value = String(state.fxBeauty);
  if (els.fxBeautyVal) els.fxBeautyVal.textContent = String(state.fxBeauty);
  if (els.fxSlim && els.fxSlim.value !== String(state.fxSlim)) els.fxSlim.value = String(state.fxSlim);
  if (els.fxSlimVal) els.fxSlimVal.textContent = String(state.fxSlim);
  if (els.fxFilter && els.fxFilter.value !== state.fxFilter) els.fxFilter.value = state.fxFilter;

  const beautyDisabled = !state.fxBeautyEnabled;
  if (els.fxBeauty) els.fxBeauty.disabled = beautyDisabled;
  if (els.fxBeautyType) els.fxBeautyType.disabled = beautyDisabled;

  const filterDisabled = !state.fxFilterEnabled;
  if (els.fxFilter) els.fxFilter.disabled = filterDisabled;
}

function setLocalFx({ beautyEnabled, filterEnabled, beautyType, beauty, slim, filter } = {}) {
  let changed = false;

  if (typeof beautyEnabled === "boolean" && beautyEnabled !== state.fxBeautyEnabled) {
    state.fxBeautyEnabled = beautyEnabled;
    changed = true;
  }

  if (typeof filterEnabled === "boolean" && filterEnabled !== state.fxFilterEnabled) {
    state.fxFilterEnabled = filterEnabled;
    changed = true;
  }

  if (typeof beautyType === "string") {
    const bt = beautyType || "natural";
    if (bt !== state.fxBeautyType) {
      state.fxBeautyType = bt;
      changed = true;
    }
  }

  if (typeof beauty === "number" && Number.isFinite(beauty)) {
    const b = Math.max(0, Math.min(100, Math.round(beauty)));
    if (b !== state.fxBeauty) {
      state.fxBeauty = b;
      changed = true;
    }
  }

  if (typeof slim === "number" && Number.isFinite(slim)) {
    const s = Math.max(0, Math.min(100, Math.round(slim)));
    if (s !== state.fxSlim) {
      state.fxSlim = s;
      changed = true;
    }
  }

  if (typeof filter === "string") {
    const f = filter || "none";
    if (f !== state.fxFilter) {
      state.fxFilter = f;
      changed = true;
    }
  }

  if (!changed) return;
  state.fxVersion += 1;
  updateFxUI();
  applyFxToTile("local");
}

function getFxForTile(tileKey) {
  if (tileKey === "local") {
    return {
      beautyEnabled: state.fxBeautyEnabled,
      filterEnabled: state.fxFilterEnabled,
      beautyType: state.fxBeautyType,
      beauty: state.fxBeauty,
      slim: state.fxSlim,
      filter: state.fxFilter,
    };
  }
  const fx = state.fxRemote.get(tileKey);
  if (fx && typeof fx.beauty === "number" && typeof fx.filter === "string") return fx;
  return { beautyEnabled: false, filterEnabled: false, beautyType: "natural", beauty: 0, slim: 0, filter: "none" };
}

function fxToCssFilter(fx) {
  const rawBeauty = Math.max(0, Math.min(100, Number(fx?.beauty ?? 0)));
  const rawSlim = Math.max(0, Math.min(100, Number(fx?.slim ?? 0)));
  const beautyEnabled = typeof fx?.beautyEnabled === "boolean" ? fx.beautyEnabled : rawBeauty > 0;
  const beautyType = String(fx?.beautyType || "natural");
  const beauty = beautyEnabled ? rawBeauty : 0;

  let blurMax = 0.9;
  let brightenMax = 0.06;
  if (beautyType === "smooth") {
    blurMax = 1.4;
    brightenMax = 0.08;
  } else if (beautyType === "strong") {
    blurMax = 2.2;
    brightenMax = 0.1;
  } else if (beautyType === "bright") {
    blurMax = 0.8;
    brightenMax = 0.12;
  } else if (beautyType === "soft") {
    blurMax = 1.2;
    brightenMax = 0.09;
  }

  const blur = (beauty / 100) * blurMax;
  const brighten = 1 + (beauty / 100) * brightenMax;
  const soften = beauty > 0 ? `blur(${blur.toFixed(2)}px) brightness(${brighten.toFixed(3)})` : "";

  const rawFilter = String(fx?.filter || "none");
  const filterEnabled = typeof fx?.filterEnabled === "boolean" ? fx.filterEnabled : rawFilter !== "none";
  const f = filterEnabled ? rawFilter : "none";
  const tone =
    f === "warm"
      ? "sepia(0.22) saturate(1.25) hue-rotate(-8deg)"
      : f === "cool"
        ? "saturate(1.15) hue-rotate(12deg)"
        : f === "gray"
          ? "grayscale(1)"
          : f === "vivid"
            ? "contrast(1.12) saturate(1.45)"
            : f === "retro"
              ? "sepia(0.38) contrast(1.05) saturate(0.85)"
              : f === "film"
                ? "contrast(1.08) saturate(0.9) brightness(0.98)"
                : f === "soft"
                  ? "blur(0.6px) saturate(1.1)"
                  : f === "cyber"
                    ? "contrast(1.25) saturate(1.35) hue-rotate(18deg)"
                    : "";

  const parts = [soften, tone].filter(Boolean);
  return parts.length ? parts.join(" ") : "";
}

function applyFxToTile(tileKey) {
  const tile = document.getElementById(`tile_${tileKey}`);
  if (!tile) return;
  const video = tile.querySelector("video");
  if (!video) return;
  const fx = getFxForTile(tileKey);
  video.style.filter = fxToCssFilter(fx);
}

function setRemoteFx(tileKey, fx) {
  if (!tileKey) return;
  const beauty = Math.max(0, Math.min(100, Number(fx?.beauty ?? fx?.b ?? 0)));
  const slim = Math.max(0, Math.min(100, Number(fx?.slim ?? fx?.s ?? 0)));
  const filter = String(fx?.filter ?? fx?.f ?? "none") || "none";
  const beautyEnabled =
    typeof fx?.beautyEnabled === "boolean" ? fx.beautyEnabled : typeof fx?.eb === "boolean" ? fx.eb : beauty > 0;
  const filterEnabled =
    typeof fx?.filterEnabled === "boolean" ? fx.filterEnabled : typeof fx?.ef === "boolean" ? fx.ef : filter !== "none";
  const beautyType = typeof fx?.beautyType === "string" ? fx.beautyType : typeof fx?.bt === "string" ? fx.bt : "natural";
  const updatedAt = Number(fx?.updatedAt ?? fx?.t ?? Date.now());
  state.fxRemote.set(tileKey, {
    beautyEnabled: Boolean(beautyEnabled),
    filterEnabled: Boolean(filterEnabled),
    beautyType: beautyType || "natural",
    beauty: Math.round(beauty),
    slim: Math.round(slim),
    filter,
    updatedAt,
  });
  applyFxToTile(tileKey);
}

function clearRemoteFx(tileKey) {
  if (!tileKey) return;
  state.fxRemote.delete(tileKey);
}

function initFxControls() {
  updateFxUI();

  state.fxProcessingSupported = supportsFxProcessing();
  state.fxMirrorX = isAndroid();

  if (els.fxEnableBeauty) {
    els.fxEnableBeauty.addEventListener("change", () => {
      setLocalFx({ beautyEnabled: Boolean(els.fxEnableBeauty.checked) });
    });
  }

  if (els.fxEnableFilter) {
    els.fxEnableFilter.addEventListener("change", () => {
      setLocalFx({ filterEnabled: Boolean(els.fxEnableFilter.checked) });
    });
  }

  if (els.fxBeautyType) {
    els.fxBeautyType.addEventListener("change", () => {
      setLocalFx({ beautyType: String(els.fxBeautyType.value || "natural") });
    });
  }

  if (els.fxBeauty) {
    els.fxBeauty.addEventListener("input", () => {
      const v = Number(els.fxBeauty.value || 0);
      setLocalFx({ beauty: v });
    });
  }

  if (els.fxFilter) {
    els.fxFilter.addEventListener("change", () => {
      setLocalFx({ filter: String(els.fxFilter.value || "none") });
    });
  }

  if (els.fxSlim) {
    els.fxSlim.addEventListener("input", () => {
      const v = Number(els.fxSlim.value || 0);
      setLocalFx({ slim: v });
    });
  }

  state.fxSeiReady = canUseEncodedInsertableStreams();
  if (state.fxSeiReady) {
    setSeiStatus("可用", "ok");
  } else if (isSafari()) {
    setSeiStatus("不支持（Safari 缺少 Insertable Streams）", "warn");
  } else {
    setSeiStatus("不支持", "warn");
  }
}

function createCanvasRenderer(width, height) {
  const canvas =
    typeof OffscreenCanvas !== "undefined" ? new OffscreenCanvas(Math.max(width, 2), Math.max(height, 2)) : (() => {
      const c = document.createElement("canvas");
      c.width = Math.max(width, 2);
      c.height = Math.max(height, 2);
      return c;
    })();
  const ctx = canvas.getContext("2d");
  if (!ctx) return null;
  return {
    canvas,
    width: canvas.width,
    height: canvas.height,
    render(frame, params) {
      const w = frame.displayWidth || frame.codedWidth || this.width || 640;
      const h = frame.displayHeight || frame.codedHeight || this.height || 360;
      if (w !== this.width || h !== this.height) {
        this.width = w;
        this.height = h;
        canvas.width = w;
        canvas.height = h;
      }
      const radius = Math.max(0.6, params.radius || 1.0);
      const brightness = 1 + (params.brightness || 0);
      const saturation = params.saturation || 1;
      const blurPx = radius * 1.5;
      ctx.filter = `blur(${blurPx.toFixed(2)}px) brightness(${brightness.toFixed(3)}) saturate(${saturation.toFixed(3)})`;
      try {
        ctx.save();
        if (params.mirrorX) {
          ctx.translate(w, 0);
          ctx.scale(-1, 1);
        }
        if (params.slim && params.slim > 0) {
          const sx = Math.max(0.65, 1 - params.slim * 0.28);
          const sy = 1;
          ctx.translate(w / 2, h / 2);
          ctx.scale(sx, sy);
          ctx.drawImage(frame, -w / 2, -h / 2, w, h);
        } else {
          ctx.drawImage(frame, 0, 0, w, h);
        }
        ctx.restore();
        return new VideoFrame(canvas, { timestamp: frame.timestamp, duration: frame.duration });
      } catch {
        return null;
      }
    },
  };
}

function createFxRenderer(width, height) {
  const canvas =
    typeof OffscreenCanvas !== "undefined" ? new OffscreenCanvas(Math.max(width, 2), Math.max(height, 2)) : (() => {
      const c = document.createElement("canvas");
      c.width = Math.max(width, 2);
      c.height = Math.max(height, 2);
      return c;
    })();

  const gl = canvas.getContext("webgl", { premultipliedAlpha: false, preserveDrawingBuffer: false });
  if (!gl) return createCanvasRenderer(width, height);
  gl.pixelStorei(gl.UNPACK_FLIP_Y_WEBGL, true);

  const vsSrc = `
    attribute vec2 a_position;
    attribute vec2 a_texCoord;
    varying vec2 v_texCoord;
    void main() {
      v_texCoord = a_texCoord;
      gl_Position = vec4(a_position, 0.0, 1.0);
    }
  `;
  const fsSrc = `
    precision mediump float;
    varying vec2 v_texCoord;
    uniform sampler2D u_texture;
    uniform vec2 u_texel;
    uniform float u_beautyMix;
    uniform float u_radius;
    uniform int u_filterMode;
    uniform float u_brightness;
    uniform float u_saturation;
    uniform float u_slim;

    vec2 warpSlim(vec2 uv) {
      if (u_slim <= 0.0) return uv;
      vec2 c = vec2(0.5, 0.5);
      vec2 delta = uv - c;
      float dist = length(delta);
      float weight = exp(-pow(dist * 5.2, 2.0));
      float kx = max(0.65, 1.0 - u_slim * 0.28);
      float factor = mix(1.0, kx, weight);
      delta.x *= factor;
      return c + delta;
    }

    vec3 applyFilter(vec3 c) {
      if (u_filterMode == 1) { // warm
        c.r *= 1.05; c.g *= 1.02; c.b *= 0.98;
      } else if (u_filterMode == 2) { // cool
        c.r *= 0.98; c.g *= 1.02; c.b *= 1.06;
      } else if (u_filterMode == 3) { // gray
        float g = dot(c, vec3(0.299, 0.587, 0.114));
        c = vec3(g);
      } else if (u_filterMode == 4) { // vivid
        c = mix(vec3(dot(c, vec3(0.299, 0.587, 0.114))), c, 1.35);
      }
      return c;
    }

    vec3 adjustSaturation(vec3 c, float sat) {
      float g = dot(c, vec3(0.299, 0.587, 0.114));
      return mix(vec3(g), c, sat);
    }

    void main() {
      vec2 o = u_texel * u_radius;
      vec3 sum = vec3(0.0);
      vec2 base = warpSlim(v_texCoord);
      sum += texture2D(u_texture, warpSlim(base + vec2(-o.x, -o.y))).rgb;
      sum += texture2D(u_texture, warpSlim(base + vec2( 0.0, -o.y))).rgb;
      sum += texture2D(u_texture, warpSlim(base + vec2( o.x, -o.y))).rgb;
      sum += texture2D(u_texture, warpSlim(base + vec2(-o.x,  0.0))).rgb;
      sum += texture2D(u_texture, base).rgb;
      sum += texture2D(u_texture, warpSlim(base + vec2( o.x,  0.0))).rgb;
      sum += texture2D(u_texture, warpSlim(base + vec2(-o.x,  o.y))).rgb;
      sum += texture2D(u_texture, warpSlim(base + vec2( 0.0,  o.y))).rgb;
      sum += texture2D(u_texture, warpSlim(base + vec2( o.x,  o.y))).rgb;
      vec3 blur = sum / 9.0;

      vec3 orig = texture2D(u_texture, base).rgb;
      vec3 mixed = mix(orig, blur, clamp(u_beautyMix, 0.0, 1.0));
      vec3 filtered = applyFilter(mixed);
      filtered = adjustSaturation(filtered, u_saturation);
      filtered *= (1.0 + u_brightness);
      gl_FragColor = vec4(filtered, 1.0);
    }
  `;

  function compile(type, src) {
    const shader = gl.createShader(type);
    gl.shaderSource(shader, src);
    gl.compileShader(shader);
    if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) return null;
    return shader;
  }
  const vs = compile(gl.VERTEX_SHADER, vsSrc);
  const fs = compile(gl.FRAGMENT_SHADER, fsSrc);
  if (!vs || !fs) return null;

  const program = gl.createProgram();
  gl.attachShader(program, vs);
  gl.attachShader(program, fs);
  gl.linkProgram(program);
  if (!gl.getProgramParameter(program, gl.LINK_STATUS)) return null;

  const posLoc = gl.getAttribLocation(program, "a_position");
  const texLoc = gl.getAttribLocation(program, "a_texCoord");
  const uTexel = gl.getUniformLocation(program, "u_texel");
  const uBeauty = gl.getUniformLocation(program, "u_beautyMix");
  const uRadius = gl.getUniformLocation(program, "u_radius");
  const uFilterMode = gl.getUniformLocation(program, "u_filterMode");
  const uBrightness = gl.getUniformLocation(program, "u_brightness");
  const uSaturation = gl.getUniformLocation(program, "u_saturation");
  const uSlim = gl.getUniformLocation(program, "u_slim");

  const posBuf = gl.createBuffer();
  gl.bindBuffer(gl.ARRAY_BUFFER, posBuf);
  gl.bufferData(
    gl.ARRAY_BUFFER,
    new Float32Array([-1, -1, 1, -1, -1, 1, 1, 1]),
    gl.STATIC_DRAW,
  );

  const texBuf = gl.createBuffer();
  gl.bindBuffer(gl.ARRAY_BUFFER, texBuf);
  gl.bufferData(
    gl.ARRAY_BUFFER,
    new Float32Array([0, 0, 1, 0, 0, 1, 1, 1]),
    gl.STATIC_DRAW,
  );

  const texBufMirror = gl.createBuffer();
  gl.bindBuffer(gl.ARRAY_BUFFER, texBufMirror);
  gl.bufferData(
    gl.ARRAY_BUFFER,
    new Float32Array([1, 0, 0, 0, 1, 1, 0, 1]),
    gl.STATIC_DRAW,
  );

  const texture = gl.createTexture();
  gl.bindTexture(gl.TEXTURE_2D, texture);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);

  return {
    canvas,
    gl,
    program,
    posLoc,
    texLoc,
    uTexel,
    uBeauty,
    uRadius,
    uFilterMode,
    uBrightness,
    uSaturation,
    texture,
    width: canvas.width,
    height: canvas.height,
    render(frame, params) {
      const w = frame.displayWidth || frame.codedWidth || params.width || canvas.width || 640;
      const h = frame.displayHeight || frame.codedHeight || params.height || canvas.height || 360;
      if (w !== this.width || h !== this.height) {
        this.width = w;
        this.height = h;
        canvas.width = w;
        canvas.height = h;
      }

      gl.viewport(0, 0, w, h);
      gl.useProgram(program);

      gl.bindBuffer(gl.ARRAY_BUFFER, posBuf);
      gl.enableVertexAttribArray(posLoc);
      gl.vertexAttribPointer(posLoc, 2, gl.FLOAT, false, 0, 0);

      gl.bindBuffer(gl.ARRAY_BUFFER, params.mirrorX ? texBufMirror : texBuf);
      gl.enableVertexAttribArray(texLoc);
      gl.vertexAttribPointer(texLoc, 2, gl.FLOAT, false, 0, 0);

      gl.activeTexture(gl.TEXTURE0);
      gl.bindTexture(gl.TEXTURE_2D, texture);
      try {
        gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, frame);
      } catch {
        return null;
      }

      gl.uniform2f(uTexel, 1 / w, 1 / h);
      gl.uniform1f(uBeauty, Math.max(0, Math.min(1, params.beautyMix || 0)));
      gl.uniform1f(uRadius, Math.max(0.6, params.radius || 1.0));
      gl.uniform1i(uFilterMode, params.filterMode || 0);
      gl.uniform1f(uBrightness, params.brightness || 0);
      gl.uniform1f(uSaturation, params.saturation || 1);
      gl.uniform1f(uSlim, Math.max(0, Math.min(1, params.slim || 0)));

      gl.drawArrays(gl.TRIANGLE_STRIP, 0, 4);

      try {
        return new VideoFrame(canvas, { timestamp: frame.timestamp, duration: frame.duration });
      } catch {
        return null;
      }
    },
  };
}

function mapFilterMode(filter, filterEnabled) {
  if (!filterEnabled) return { mode: 0, saturation: 1, brightness: 0 };
  const f = String(filter || "none");
  if (f === "warm") return { mode: 1, saturation: 1.08, brightness: 0.02 };
  if (f === "cool") return { mode: 2, saturation: 1.05, brightness: 0.01 };
  if (f === "gray") return { mode: 3, saturation: 0.0, brightness: 0.0 };
  if (f === "vivid") return { mode: 4, saturation: 1.25, brightness: 0.02 };
  if (f === "retro") return { mode: 1, saturation: 0.85, brightness: -0.02 };
  if (f === "film") return { mode: 2, saturation: 0.9, brightness: -0.01 };
  if (f === "soft") return { mode: 0, saturation: 1.1, brightness: 0.03 };
  if (f === "cyber") return { mode: 4, saturation: 1.4, brightness: 0.04 };
  return { mode: 0, saturation: 1.0, brightness: 0.0 };
}

function startFxPipeline(sourceTrack) {
  if (!sourceTrack || sourceTrack.readyState !== "live") return null;
  if (!state.fxProcessingSupported) return null;

  let renderer = null;
  const processor = new MediaStreamTrackProcessor({ track: sourceTrack });
  const generator = new MediaStreamTrackGenerator({ kind: "video" });
  const reader = processor.readable;
  const writer = generator.writable;

  const stopPipeline = () => {
    try {
      reader.cancel().catch(() => null);
    } catch {
      // ignore
    }
    try {
      writer.abort().catch(() => null);
    } catch {
      // ignore
    }
    try {
      generator.stop();
    } catch {
      // ignore
    }
  };

  const transformer = new TransformStream({
    transform(frame, controller) {
      try {
        if (!renderer) {
          const w = frame.displayWidth || frame.codedWidth || 640;
          const h = frame.displayHeight || frame.codedHeight || 360;
          renderer = createFxRenderer(w, h);
          if (!renderer) {
            controller.enqueue(frame);
            return;
          }
        }

        const beautyEnabled = state.fxBeautyEnabled;
        const beautyStrength = beautyEnabled ? Math.max(0, Math.min(1, state.fxBeauty / 100)) : 0;
        const beautyType = String(state.fxBeautyType || "natural");
        const radius = beautyType === "strong" ? 2.2 : beautyType === "smooth" ? 1.6 : 1.1;
        const filterMeta = mapFilterMode(state.fxFilter, state.fxFilterEnabled);
        const brightness = (beautyType === "strong" ? 0.08 : beautyType === "smooth" ? 0.06 : 0.04) * beautyStrength + filterMeta.brightness;
        const slimStrength = Math.max(0, Math.min(1, state.fxSlim / 100));

        const processed = renderer.render(frame, {
          beautyMix: beautyStrength,
          radius,
          filterMode: filterMeta.mode,
          brightness,
          saturation: filterMeta.saturation,
          mirrorX: state.fxMirrorX,
          slim: slimStrength,
        });

        if (!processed) {
          const w = frame.displayWidth || frame.codedWidth || 640;
          const h = frame.displayHeight || frame.codedHeight || 360;
          const fallback = createCanvasRenderer(w, h);
          if (fallback) {
            renderer = fallback;
            processed = renderer.render(frame, {
              beautyMix: beautyStrength,
              radius,
              filterMode: filterMeta.mode,
              brightness,
              saturation: filterMeta.saturation,
              mirrorX: state.fxMirrorX,
              slim: slimStrength,
            });
          }
        }

        if (processed) {
          controller.enqueue(processed);
          frame.close();
        } else {
          controller.enqueue(frame);
        }
      } catch {
        controller.enqueue(frame);
      }
    },
  });

  reader.pipeThrough(transformer).pipeTo(writer).catch(() => null);

  return { track: generator, stop: stopPipeline };
}

function stopFxPipeline() {
  if (state.fxPipelineStop) {
    try {
      state.fxPipelineStop();
    } catch {
      // ignore
    }
  }
  state.fxPipelineStop = null;
  state.fxProcessedTrack = null;
  state.fxPipelineSourceId = null;
}

async function ensureFxPipelineForCamera() {
  const camTrack = state.localStream?.getVideoTracks?.()[0];
  if (!camTrack) return null;
  if (!state.fxProcessingSupported) state.fxProcessingSupported = supportsFxProcessing();
  if (!state.fxProcessingSupported) return null;

  if (state.fxProcessedTrack && state.fxPipelineSourceId === camTrack.id && state.fxProcessedTrack.readyState === "live") {
    return state.fxProcessedTrack;
  }

  stopFxPipeline();
  const pipeline = startFxPipeline(camTrack);
  if (!pipeline || !pipeline.track) return null;
  state.fxProcessedTrack = pipeline.track;
  state.fxPipelineStop = pipeline.stop;
  state.fxPipelineSourceId = camTrack.id;
  try {
    camTrack.addEventListener("ended", stopFxPipeline, { once: true });
  } catch {
    // ignore
  }
  return state.fxProcessedTrack;
}

function setAuthMsg(text, kind = "info") {
  if (!text) {
    els.authMsg.textContent = "";
    els.authMsg.style.color = "";
    return;
  }
  els.authMsg.textContent = text;
  els.authMsg.style.color = kind === "error" ? "rgba(255,77,79,0.95)" : kind === "ok" ? "rgba(61,214,208,0.95)" : "";
}

function setHint(el, text, kind = "info") {
  el.textContent = text || "";
  el.style.color = kind === "error" ? "rgba(255,77,79,0.95)" : kind === "ok" ? "rgba(61,214,208,0.95)" : "";
}

function setAiStatus(pillText, kind = "info", detailText = "") {
  if (!els.aiStatusPill || !els.aiStatusText) return;
  els.aiStatusPill.textContent = pillText || "-";
  els.aiStatusText.textContent = detailText || "";
  els.aiStatusPill.classList.remove("pill--ok", "pill--warn", "pill--err");
  if (kind === "ok") els.aiStatusPill.classList.add("pill--ok");
  if (kind === "warn") els.aiStatusPill.classList.add("pill--warn");
  if (kind === "error") els.aiStatusPill.classList.add("pill--err");
}

function safeJsonStringify(data) {
  try {
    return JSON.stringify(data, null, 2);
  } catch {
    return String(data);
  }
}

function aiAppendOutput(title, data) {
  if (!els.aiOutput) return;
  const ts = new Date().toISOString();
  const block = `[${ts}] ${title}\n${data == null ? "" : typeof data === "string" ? data : safeJsonStringify(data)}\n\n`;
  els.aiOutput.textContent = (els.aiOutput.textContent || "") + block;
  if (els.aiOutput.textContent.length > 60000) {
    els.aiOutput.textContent = els.aiOutput.textContent.slice(-60000);
  }
  els.aiOutput.scrollTop = els.aiOutput.scrollHeight;
}

function aiClearOutput() {
  if (!els.aiOutput) return;
  els.aiOutput.textContent = "";
  if (!state.aiLiveEnabled) setAiStatus("-", "info", "未启用");
}

function fileToBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = () => reject(new Error("文件读取失败"));
    reader.onload = () => {
      const res = String(reader.result || "");
      const idx = res.indexOf(",");
      resolve(idx >= 0 ? res.slice(idx + 1) : res);
    };
    reader.readAsDataURL(file);
  });
}

async function aiHealthCheck() {
  if (!els.aiHealthBtn) return;
  setAiStatus("检测中", "warn", "正在请求 /api/v1/ai/health …");
  els.aiHealthBtn.disabled = true;
  try {
    const json = await apiFetch("/api/v1/ai/health", { auth: false });
    const data = json?.data || {};
    setAiStatus("正常", "ok", `service=${data.service || "ai"} · ${new Date().toLocaleTimeString()}`);
    aiAppendOutput("AI 健康检查", data);
  } catch (err) {
    setAiStatus("异常", "error", err?.message || "健康检查失败");
    aiAppendOutput("AI 健康检查失败", err?.message || String(err));
  } finally {
    els.aiHealthBtn.disabled = false;
  }
}

async function aiGetInfo() {
  if (!els.aiInfoBtn) return;
  els.aiInfoBtn.disabled = true;
  try {
    const json = await apiFetch("/api/v1/ai/info", { auth: false });
    aiAppendOutput("AI 服务信息", json?.data || json);
  } catch (err) {
    aiAppendOutput("AI 服务信息失败", err?.message || String(err));
  } finally {
    els.aiInfoBtn.disabled = false;
  }
}

async function aiRunEmotion() {
  const text = (els.aiEmotionText?.value || "").trim();
  if (!text) throw new Error("请输入文本");
  const body = { text };
  if (state.meetingId) body.meeting_id = Number(state.meetingId);
  const json = await apiFetch("/api/v1/ai/emotion", { method: "POST", auth: false, body });
  const data = json?.data || {};
  aiAppendOutput(`情感：${data.emotion || "-"}`, data);
}

async function aiRunAsr() {
  const file = els.aiAsrFile?.files?.[0];
  if (!file) throw new Error("请选择音频文件");
  if (file.size > 25 * 1024 * 1024) throw new Error("文件过大（建议 ≤ 25MB）");

  const audioData = await fileToBase64(file);
  const format = (els.aiAsrFormat?.value || "wav").trim() || "wav";
  const sampleRate = Number.parseInt(els.aiAsrSampleRate?.value || "16000", 10) || 16000;
  const language = (els.aiAsrLang?.value || "").trim();

  const body = { audio_data: audioData, format, sample_rate: sampleRate };
  if (state.meetingId) body.meeting_id = Number(state.meetingId);
  if (language) body.language = language;

  const json = await apiFetch("/api/v1/ai/asr", { method: "POST", auth: false, body });
  const data = json?.data || {};
  const text = data.text || data.transcription || "";
  aiAppendOutput(`语音识别：${text ? text.slice(0, 80) : "-"}`, data);
}

async function aiRunSynthesis() {
  const file = els.aiSynthFile?.files?.[0];
  if (!file) throw new Error("请选择音频文件");
  if (file.size > 25 * 1024 * 1024) throw new Error("文件过大（建议 ≤ 25MB）");

  const audioData = await fileToBase64(file);
  const format = (els.aiSynthFormat?.value || "wav").trim() || "wav";
  const sampleRate = Number.parseInt(els.aiSynthSampleRate?.value || "16000", 10) || 16000;

  const body = { audio_data: audioData, format, sample_rate: sampleRate };
  if (state.meetingId) body.meeting_id = Number(state.meetingId);
  const json = await apiFetch("/api/v1/ai/synthesis", { method: "POST", auth: false, body });
  const data = json?.data || {};
  aiAppendOutput(`深度伪造：${data.is_synthetic ? "疑似合成" : "较像真实"}`, data);
}

const AI_LIVE = {
  tickMs: 180,
  rmsThreshold: 0.012,
  silenceMs: 700,
  minSegmentMs: 450,
  maxSegmentMs: 8000,
  targetSampleRate: 16000,
};

function formatHms(ts) {
  try {
    return new Date(ts).toLocaleTimeString();
  } catch {
    return "";
  }
}

function getParticipantByPeerId(peerId) {
  if (!peerId) return null;
  return state.participantsByPeerId.get(peerId) || null;
}

function formatParticipantLabel(p) {
  if (!p) return "";
  const name = p.username || (p.user_id ? `user_${p.user_id}` : "user");
  const suffix = p.user_id ? ` (#${p.user_id})` : "";
  return `${name}${suffix}`;
}

function getSpeakerLabel(speakerKey) {
  if (!speakerKey) return "-";
  if (speakerKey === "local") return "我";
  const p = getParticipantByPeerId(speakerKey);
  if (p) return formatParticipantLabel(p);
  return `远端 ${String(speakerKey).slice(0, 8)}`;
}

function setAiLiveSpeakerUI(speakerKey, level) {
  if (els.aiLiveSpeaker) els.aiLiveSpeaker.textContent = speakerKey ? getSpeakerLabel(speakerKey) : "-";
  if (els.aiLiveLevel) els.aiLiveLevel.textContent = typeof level === "number" ? level.toFixed(3) : "-";
}

function clearAiLiveLog() {
  if (els.aiLiveLog) els.aiLiveLog.innerHTML = "";
  state.aiLiveLineEls.clear();
  if (els.aiDanmaku) els.aiDanmaku.innerHTML = "";
}

function addAiLiveTag(tagsEl, text, kind = "info") {
  if (!tagsEl) return;
  const span = document.createElement("span");
  span.className = `tag ${kind === "ok" ? "tag--ok" : kind === "warn" ? "tag--warn" : kind === "error" ? "tag--err" : ""}`;
  span.textContent = text;
  tagsEl.appendChild(span);
}

function addDanmaku({ who, text, tags = [] }) {
  if (!els.aiDanmaku) return;
  const content = `${who ? `${who}: ` : ""}${text || ""}${tags.length ? " · " + tags.map((t) => t.text).join(" / ") : ""}`;
  if (!content.trim()) return;

  const item = document.createElement("div");
  item.className = "danmaku__item";
  item.textContent = content;

  const laneCount = 6;
  const lane = state.danmakuLane % laneCount;
  state.danmakuLane += 1;
  item.style.top = `${8 + lane * 24}px`;
  const duration = 8 + Math.random() * 4;
  item.style.animationDuration = `${duration}s`;

  item.addEventListener("animationend", () => {
    try {
      item.remove();
    } catch {
      // ignore
    }
  });

  els.aiDanmaku.appendChild(item);
}

function upsertAiLiveLine(lineId, { who, timeText, text, tags = [] }) {
  if (!els.aiLiveLog) return;

  const existing = state.aiLiveLineEls.get(lineId);
  if (existing) {
    if (existing.textEl) existing.textEl.textContent = text || "";
    if (existing.tagsEl) {
      existing.tagsEl.innerHTML = "";
      for (const t of tags) addAiLiveTag(existing.tagsEl, t.text, t.kind);
    }
    return;
  }

  const root = document.createElement("div");
  root.className = "tline";

  const meta = document.createElement("div");
  meta.className = "tline__meta";

  const whoEl = document.createElement("div");
  whoEl.className = "tline__who";
  whoEl.textContent = who || "-";

  const timeEl = document.createElement("div");
  timeEl.className = "tline__time";
  timeEl.textContent = timeText || "";

  const tagsEl = document.createElement("div");
  tagsEl.className = "tline__tags";
  for (const t of tags) addAiLiveTag(tagsEl, t.text, t.kind);

  const textEl = document.createElement("div");
  textEl.className = "tline__text";
  textEl.textContent = text || "";

  meta.appendChild(whoEl);
  meta.appendChild(timeEl);
  meta.appendChild(tagsEl);
  root.appendChild(meta);
  root.appendChild(textEl);

  els.aiLiveLog.appendChild(root);
  els.aiLiveLog.scrollTop = els.aiLiveLog.scrollHeight;

  state.aiLiveLineEls.set(lineId, { root, textEl, tagsEl });

  addDanmaku({ who, text, tags });
}

function bytesToBase64(bytes) {
  let binary = "";
  const chunkSize = 0x2000;
  for (let i = 0; i < bytes.length; i += chunkSize) {
    binary += String.fromCharCode(...bytes.subarray(i, i + chunkSize));
  }
  return btoa(binary);
}

function downsampleToTarget(input, inputRate, outputRate) {
  if (!input || input.length === 0) return new Float32Array();
  if (!inputRate || inputRate === outputRate) return input;

  const ratio = inputRate / outputRate;
  const newLength = Math.max(1, Math.round(input.length / ratio));
  const output = new Float32Array(newLength);
  let offsetResult = 0;
  let offsetBuffer = 0;
  while (offsetResult < output.length) {
    const nextOffsetBuffer = Math.round((offsetResult + 1) * ratio);
    let sum = 0;
    let count = 0;
    for (let i = offsetBuffer; i < nextOffsetBuffer && i < input.length; i++) {
      sum += input[i];
      count++;
    }
    output[offsetResult] = count ? sum / count : 0;
    offsetResult++;
    offsetBuffer = nextOffsetBuffer;
  }
  return output;
}

function encodeWavPCM16Mono(samples, sampleRate) {
  const buffer = new ArrayBuffer(44 + samples.length * 2);
  const view = new DataView(buffer);

  function writeString(offset, str) {
    for (let i = 0; i < str.length; i++) view.setUint8(offset + i, str.charCodeAt(i));
  }

  writeString(0, "RIFF");
  view.setUint32(4, 36 + samples.length * 2, true);
  writeString(8, "WAVE");
  writeString(12, "fmt ");
  view.setUint32(16, 16, true);
  view.setUint16(20, 1, true);
  view.setUint16(22, 1, true);
  view.setUint32(24, sampleRate, true);
  view.setUint32(28, sampleRate * 2, true);
  view.setUint16(32, 2, true);
  view.setUint16(34, 16, true);
  writeString(36, "data");
  view.setUint32(40, samples.length * 2, true);

  let offset = 44;
  for (let i = 0; i < samples.length; i++, offset += 2) {
    const s = Math.max(-1, Math.min(1, samples[i]));
    view.setInt16(offset, s < 0 ? s * 0x8000 : s * 0x7fff, true);
  }

  return new Uint8Array(buffer);
}

function ensureAiAudioContext() {
  if (state.aiAudioCtx) return state.aiAudioCtx;
  const Ctx = window.AudioContext || window.webkitAudioContext;
  if (!Ctx) throw new Error("当前浏览器不支持 AudioContext");

  const ctx = new Ctx({ latencyHint: "interactive" });
  state.aiAudioCtx = ctx;

  if (!ctx.createScriptProcessor) {
    throw new Error("当前浏览器不支持 ScriptProcessorNode（无法启用实时检测）");
  }

  const processor = ctx.createScriptProcessor(4096, 1, 1);
  const sink = ctx.createGain();
  sink.gain.value = 0;
  processor.connect(sink);
  sink.connect(ctx.destination);

  processor.onaudioprocess = (ev) => {
    if (!state.aiLiveEnabled) return;
    if (!state.aiCaptureRecording) return;
    if (!state.aiCaptureSourceKey) return;

    const input = ev.inputBuffer.getChannelData(0);
    if (!input || input.length === 0) return;
    state.aiCaptureChunks.push(new Float32Array(input));
  };

  state.aiCaptureProcessor = processor;
  state.aiCaptureSink = sink;

  return ctx;
}

function rmsFromAnalyser(analyser) {
  const n = analyser.fftSize || 2048;
  const buf = new Float32Array(n);
  analyser.getFloatTimeDomainData(buf);
  let sum = 0;
  for (let i = 0; i < buf.length; i++) sum += buf[i] * buf[i];
  return Math.sqrt(sum / buf.length);
}

function ensureAiLiveAnalyser(speakerKey, stream) {
  if (!speakerKey || !stream) return;
  if (state.aiLiveAnalyzers.has(speakerKey)) return;
  if (!state.aiAudioCtx) return;

  let source;
  try {
    source = state.aiAudioCtx.createMediaStreamSource(stream);
  } catch {
    return;
  }
  const analyser = state.aiAudioCtx.createAnalyser();
  analyser.fftSize = 2048;
  analyser.smoothingTimeConstant = 0.6;
  const gain = state.aiAudioCtx.createGain();
  gain.gain.value = 0;

  source.connect(analyser);
  analyser.connect(gain);
  gain.connect(state.aiAudioCtx.destination);

  state.aiLiveAnalyzers.set(speakerKey, { stream, source, analyser, gain, lastRms: 0 });
}

function clearAiLiveAnalyzers() {
  for (const entry of state.aiLiveAnalyzers.values()) {
    try {
      entry.source?.disconnect();
    } catch {
      // ignore
    }
    try {
      entry.analyser?.disconnect();
    } catch {
      // ignore
    }
    try {
      entry.gain?.disconnect();
    } catch {
      // ignore
    }
  }
  state.aiLiveAnalyzers.clear();
}

function aiLiveUpsertInput(speakerKey, stream) {
  if (!speakerKey || !stream) return;
  const hasAudio = stream.getAudioTracks?.().length > 0;
  if (!hasAudio) return;
  state.aiLiveInputs.set(speakerKey, stream);
  if (state.aiLiveEnabled) ensureAiLiveAnalyser(speakerKey, stream);
}

function aiLiveRemoveInput(speakerKey) {
  if (!speakerKey) return;
  state.aiLiveInputs.delete(speakerKey);
  const entry = state.aiLiveAnalyzers.get(speakerKey);
  if (entry) {
    try {
      entry.source?.disconnect();
    } catch {
      // ignore
    }
    try {
      entry.analyser?.disconnect();
    } catch {
      // ignore
    }
    try {
      entry.gain?.disconnect();
    } catch {
      // ignore
    }
    state.aiLiveAnalyzers.delete(speakerKey);
  }
}

function highlightSpeakerTile(speakerKey) {
  const prev = state.aiLastHighlightedKey;
  if (prev && prev !== speakerKey) {
    const prevId = prev === "local" ? "tile_local" : `tile_sfu_${prev}`;
    document.getElementById(prevId)?.classList.remove("tile--active");
  }
  state.aiLastHighlightedKey = speakerKey;
  if (!speakerKey) return;
  const id = speakerKey === "local" ? "tile_local" : `tile_sfu_${speakerKey}`;
  document.getElementById(id)?.classList.add("tile--active");
}

function aiLiveSetCaptureSource(speakerKey) {
  if (!state.aiAudioCtx || !state.aiCaptureProcessor) return;
  if (!speakerKey) return;

  if (state.aiCaptureSourceKey === speakerKey && state.aiCaptureSource) return;

  if (state.aiCaptureSource) {
    try {
      state.aiCaptureSource.disconnect();
    } catch {
      // ignore
    }
  }

  const stream = state.aiLiveInputs.get(speakerKey);
  if (!stream) return;

  try {
    const source = state.aiAudioCtx.createMediaStreamSource(stream);
    source.connect(state.aiCaptureProcessor);
    state.aiCaptureSource = source;
    state.aiCaptureSourceKey = speakerKey;
  } catch {
    // ignore
  }
}

function aiLiveStartRecording(now) {
  state.aiCaptureRecording = true;
  state.aiCaptureStartedAt = now;
  state.aiCaptureLastVoiceAt = now;
  state.aiCaptureChunks = [];
}

function getAiLiveOptions() {
  const asr = Boolean(els.aiLiveAsr?.checked);
  const emotion = asr && Boolean(els.aiLiveEmotion?.checked);
  const synth = Boolean(els.aiLiveSynth?.checked);
  return { asr, emotion, synth };
}

function scheduleAiLiveClaim() {
  if (!state.aiLiveEnabled) return;
  if (!state.wsReady) return;
  if (!state.aiLiveLeadCapable) return;
  if (state.aiLiveClaimTimer) return;

  const delay = 250 + Math.floor(Math.random() * 450);
  state.aiLiveClaimTimer = setTimeout(() => {
    state.aiLiveClaimTimer = null;
    if (!state.aiLiveEnabled) return;
    if (!state.wsReady) return;
    if (!state.aiLiveLeadCapable) return;
    wsSend({
      id: `ai_live_claim_${uuid()}`,
      type: WS_TYPES.AI_LIVE_CLAIM,
      peer_id: state.peerId,
      payload: { enable: true },
      timestamp: new Date().toISOString(),
    });
  }, delay);
}

function applyAiLiveStatus(status) {
  if (!status) return;
  state.aiLiveStatus = status;

  const enabled = Boolean(status.enabled);
  const leaderSessionId = String(status.leader_session_id || "");
  const leaderUsername = String(status.leader_username || "");
  const isMeLeader = Boolean(state.sessionId && leaderSessionId && leaderSessionId === state.sessionId);

  if (!state.aiLiveEnabled) {
    if (state.aiLiveIsLeader) stopAiLiveCapture().catch(() => null);
    return;
  }

  if (!enabled || !leaderSessionId) {
    if (state.aiLiveIsLeader) stopAiLiveCapture().catch(() => null);
    setAiStatus("等待中", "warn", "会议 AI Live 未启用，正在尝试接管…");
    scheduleAiLiveClaim();
    return;
  }

  if (isMeLeader) {
    if (!state.aiLiveLeadCapable) {
      wsSend({
        id: `ai_live_release_${uuid()}`,
        type: WS_TYPES.AI_LIVE_CLAIM,
        peer_id: state.peerId,
        payload: { enable: false },
        timestamp: new Date().toISOString(),
      });
      setAiStatus("不可用", "error", "本端无法运行 AI Live（已释放领导者权限）");
      return;
    }

    setAiStatus("启动中", "warn", "你是 AI Live 领导者，正在初始化音频捕获…");
    startAiLiveCapture()
      .then(() => {
        setAiStatus("运行中", "ok", "AI Live 领导者：你（本会议仅一人调用 AI）");
      })
      .catch((err) => {
        state.aiLiveLeadCapable = false;
        wsSend({
          id: `ai_live_release_${uuid()}`,
          type: WS_TYPES.AI_LIVE_CLAIM,
          peer_id: state.peerId,
          payload: { enable: false },
          timestamp: new Date().toISOString(),
        });
        setAiStatus("不可用", "error", err?.message || "无法初始化音频捕获（已释放领导者权限）");
      });
    return;
  }

  // follower mode
  if (state.aiLiveIsLeader) stopAiLiveCapture().catch(() => null);
  const leaderLabel = leaderUsername || (status.leader_user_id ? `user_${status.leader_user_id}` : "其他用户");
  setAiStatus("跟随中", "warn", `AI Live 由 ${leaderLabel} 运行（本端仅接收结果）`);
}

function applyAiLiveResult(payload) {
  if (!payload) return;
  if (!state.aiLiveEnabled) return;

  const lineId = payload.line_id || payload.lineId;
  if (!lineId) return;

  const who = payload.speaker_label || payload.who || "-";
  const ts = payload.timestamp_ms || payload.timestampMs || payload.timestamp || 0;
  const timeText = ts ? formatHms(Number(ts)) : "";
  const text = payload.text || "";
  const tags = Array.isArray(payload.tags) ? payload.tags : [];

  upsertAiLiveLine(lineId, { who, timeText, text, tags });
}

function aiLiveEnqueueJob(job) {
  state.aiLiveQueue.push(job);
  aiLiveRunQueue().catch(() => null);
}

async function aiLiveRunQueue() {
  if (state.aiLiveQueueRunning) return;
  state.aiLiveQueueRunning = true;
  try {
    while (state.aiLiveEnabled && state.aiLiveQueue.length) {
      const job = state.aiLiveQueue.shift();
      if (!job) continue;
      await aiLiveProcessJob(job);
    }
  } finally {
    state.aiLiveQueueRunning = false;
  }
}

async function aiLiveProcessJob(job) {
  if (!state.aiLiveIsLeader) return;
  const opts = getAiLiveOptions();
  const who = job.speakerLabel || "-";
  const timeText = formatHms(job.timestamp);

  upsertAiLiveLine(job.lineId, {
    who,
    timeText,
    text: "识别中…",
    tags: [{ text: "处理中", kind: "warn" }],
  });
  wsSend({
    id: `ai_live_result_${uuid()}`,
    type: WS_TYPES.AI_LIVE_RESULT,
    peer_id: state.peerId,
    payload: {
      line_id: job.lineId,
      speaker_key: job.speakerKey || "",
      speaker_label: who,
      timestamp_ms: Number(job.timestamp || Date.now()),
      text: "识别中…",
      tags: [{ text: "处理中", kind: "warn" }],
    },
    timestamp: new Date().toISOString(),
  });

  const tasks = [];
  const meetingId = state.meetingId ? Number(state.meetingId) : null;
  if (opts.asr) {
    const body = { audio_data: job.audioBase64, format: "wav", sample_rate: AI_LIVE.targetSampleRate, language: "zh" };
    if (meetingId) body.meeting_id = meetingId;
    tasks.push(
      apiFetch("/api/v1/ai/asr", {
        method: "POST",
        auth: false,
        timeoutMs: 60000,
        body,
      })
        .then((json) => ({ kind: "asr", json }))
        .catch((err) => ({ kind: "asr", err }))
    );
  }
  if (opts.synth) {
    const body = { audio_data: job.audioBase64, format: "wav", sample_rate: AI_LIVE.targetSampleRate };
    if (meetingId) body.meeting_id = meetingId;
    tasks.push(
      apiFetch("/api/v1/ai/synthesis", {
        method: "POST",
        auth: false,
        timeoutMs: 60000,
        body,
      })
        .then((json) => ({ kind: "synth", json }))
        .catch((err) => ({ kind: "synth", err }))
    );
  }

  const results = await Promise.all(tasks);
  const tags = [];

  let asrText = "";
  for (const r of results) {
    if (r.kind === "asr") {
      if (r.err) {
        tags.push({ text: `ASR失败`, kind: "error" });
        aiAppendOutput("ASR Error", r.err?.message || String(r.err));
      } else {
        const data = r.json?.data || {};
        asrText = data.text || data.transcription || "";
        const conf = typeof data.confidence === "number" ? data.confidence : null;
        tags.push({ text: conf != null ? `ASR ${(conf * 100).toFixed(0)}%` : "ASR", kind: "ok" });
      }
    }
    if (r.kind === "synth") {
      if (r.err) {
        tags.push({ text: `合成检测失败`, kind: "error" });
        aiAppendOutput("Synthesis Error", r.err?.message || String(r.err));
      } else {
        const data = r.json?.data || {};
        const isSyn = Boolean(data.is_synthetic);
        const conf = typeof data.confidence === "number" ? data.confidence : null;
        const label = isSyn ? "疑似合成" : "较像真实";
        tags.push({ text: conf != null ? `${label} ${(conf * 100).toFixed(0)}%` : label, kind: isSyn ? "warn" : "ok" });
      }
    }
  }

  if (opts.emotion && asrText) {
    try {
      const body = { text: asrText };
      if (meetingId) body.meeting_id = meetingId;
      const json = await apiFetch("/api/v1/ai/emotion", { method: "POST", auth: false, timeoutMs: 15000, body });
      const data = json?.data || {};
      const emo = data.emotion || "";
      const conf = typeof data.confidence === "number" ? data.confidence : null;
      if (emo) tags.push({ text: conf != null ? `情绪 ${emo} ${(conf * 100).toFixed(0)}%` : `情绪 ${emo}`, kind: "ok" });
    } catch (err) {
      tags.push({ text: "情绪失败", kind: "error" });
      aiAppendOutput("Emotion Error", err?.message || String(err));
    }
  }

  const finalText = asrText || "（未识别到文本）";
  upsertAiLiveLine(job.lineId, { who, timeText, text: finalText, tags });
  wsSend({
    id: `ai_live_result_${uuid()}`,
    type: WS_TYPES.AI_LIVE_RESULT,
    peer_id: state.peerId,
    payload: {
      line_id: job.lineId,
      speaker_key: job.speakerKey || "",
      speaker_label: who,
      timestamp_ms: Number(job.timestamp || Date.now()),
      text: finalText,
      tags,
    },
    timestamp: new Date().toISOString(),
  });
}

function aiLiveStopAndEnqueue(reason) {
  if (!state.aiCaptureRecording) return;
  if (!state.aiLiveIsLeader) {
    state.aiCaptureRecording = false;
    state.aiCaptureVoiceStreak = 0;
    state.aiCaptureChunks = [];
    return;
  }

  const now = Date.now();
  const durationMs = now - state.aiCaptureStartedAt;
  const sourceKey = state.aiCaptureSourceKey;
  state.aiCaptureRecording = false;
  state.aiCaptureVoiceStreak = 0;

  if (!sourceKey || durationMs < AI_LIVE.minSegmentMs) {
    state.aiCaptureChunks = [];
    return;
  }

  // concat
  let total = 0;
  for (const c of state.aiCaptureChunks) total += c.length;
  const merged = new Float32Array(total);
  let offset = 0;
  for (const c of state.aiCaptureChunks) {
    merged.set(c, offset);
    offset += c.length;
  }
  state.aiCaptureChunks = [];

  const inputRate = state.aiAudioCtx?.sampleRate || 48000;
  const down = downsampleToTarget(merged, inputRate, AI_LIVE.targetSampleRate);
  const wavBytes = encodeWavPCM16Mono(down, AI_LIVE.targetSampleRate);
  const audioBase64 = bytesToBase64(wavBytes);

  const lineId = `live_${uuid()}`;
  const speakerLabel = getSpeakerLabel(sourceKey);
  aiLiveEnqueueJob({
    lineId,
    speakerKey: sourceKey,
    speakerLabel,
    timestamp: Date.now(),
    audioBase64,
    durationMs,
    reason,
  });
}

async function startAiLive() {
  if (state.aiLiveEnabled) return;
  if (!state.wsReady) throw new Error("请先加入会议并连接信令");
  state.aiLiveEnabled = true;
  state.aiLiveLeadCapable = true;
  setAiStatus("申请中", "warn", "正在申请会议 AI Live（避免重复调用 AI）…");
  if (els.aiLiveToggleBtn) els.aiLiveToggleBtn.textContent = "停止实时检测";
  wsSend({
    id: `ai_live_claim_${uuid()}`,
    type: WS_TYPES.AI_LIVE_CLAIM,
    peer_id: state.peerId,
    payload: { enable: true },
    timestamp: new Date().toISOString(),
  });
}

async function startAiLiveCapture() {
  if (!state.aiLiveEnabled) return;
  if (state.aiLiveIsLeader && state.aiAudioCtx) return;
  if (state.aiLiveCaptureStarting) return;
  state.aiLiveCaptureStarting = true;

  try {
    state.aiLiveIsLeader = true;
    state.aiLiveQueue = [];
    state.aiLiveQueueRunning = false;
    state.aiCaptureRecording = false;
    state.aiCaptureChunks = [];
    state.aiCaptureSourceKey = null;
    state.aiCaptureSource = null;
    state.aiCaptureVoiceStreak = 0;
    state.aiCurrentSpeakerKey = null;

    try {
      const ctx = ensureAiAudioContext();
      await ctx.resume();
    } catch (err) {
      state.aiLiveIsLeader = false;
      state.aiLiveLeadCapable = false;
      throw err;
    }

    // setup analyzers for existing inputs
    for (const [key, stream] of state.aiLiveInputs.entries()) {
      ensureAiLiveAnalyser(key, stream);
    }

    if (state.aiLiveTimer) clearInterval(state.aiLiveTimer);
    state.aiLiveTimer = setInterval(() => aiLiveTick(), AI_LIVE.tickMs);
  } finally {
    state.aiLiveCaptureStarting = false;
  }
}

async function stopAiLiveCapture() {
  if (state.aiLiveTimer) {
    clearInterval(state.aiLiveTimer);
    state.aiLiveTimer = null;
  }

  state.aiLiveIsLeader = false;
  state.aiLiveQueue = [];
  state.aiLiveQueueRunning = false;

  state.aiCaptureRecording = false;
  state.aiCaptureChunks = [];
  state.aiCaptureVoiceStreak = 0;

  try {
    state.aiCaptureSource?.disconnect();
  } catch {
    // ignore
  }
  state.aiCaptureSource = null;
  state.aiCaptureSourceKey = null;

  clearAiLiveAnalyzers();
  highlightSpeakerTile(null);
  setAiLiveSpeakerUI(null, null);

  if (state.aiAudioCtx) {
    try {
      await state.aiAudioCtx.close();
    } catch {
      // ignore
    }
  }
  state.aiAudioCtx = null;
  state.aiCaptureProcessor = null;
  state.aiCaptureSink = null;
}

async function stopAiLive() {
  if (!state.aiLiveEnabled) return;
  state.aiLiveEnabled = false;
  if (state.aiLiveClaimTimer) {
    clearTimeout(state.aiLiveClaimTimer);
    state.aiLiveClaimTimer = null;
  }

  if (state.aiLiveIsLeader) {
    wsSend({
      id: `ai_live_release_${uuid()}`,
      type: WS_TYPES.AI_LIVE_CLAIM,
      peer_id: state.peerId,
      payload: { enable: false },
      timestamp: new Date().toISOString(),
    });
  }

  await stopAiLiveCapture();

  if (els.aiLiveToggleBtn) els.aiLiveToggleBtn.textContent = "启用实时检测";
  setAiStatus("-", "info", "未启用");
}

function aiLiveTick() {
  if (!state.aiLiveEnabled || !state.aiLiveIsLeader || !state.aiAudioCtx) return;

  let bestKey = null;
  let bestLevel = 0;
  for (const [key, entry] of state.aiLiveAnalyzers.entries()) {
    if (!entry?.analyser) continue;
    let level = 0;
    try {
      level = rmsFromAnalyser(entry.analyser);
    } catch {
      level = 0;
    }
    entry.lastRms = level;
    if (level > bestLevel) {
      bestLevel = level;
      bestKey = key;
    }
  }

  const now = Date.now();
  const isVoice = bestKey && bestLevel >= AI_LIVE.rmsThreshold;

  if (isVoice) {
    setAiLiveSpeakerUI(bestKey, bestLevel);
    if (bestKey !== state.aiCurrentSpeakerKey) {
      if (state.aiCaptureRecording) aiLiveStopAndEnqueue("switch");
      state.aiCurrentSpeakerKey = bestKey;
      highlightSpeakerTile(bestKey);
      aiLiveSetCaptureSource(bestKey);
    }

    // VAD streak
    state.aiCaptureVoiceStreak++;
    if (!state.aiCaptureRecording && state.aiCaptureVoiceStreak >= 2) {
      aiLiveStartRecording(now);
    }

    if (state.aiCaptureRecording) {
      state.aiCaptureLastVoiceAt = now;
      if (now - state.aiCaptureStartedAt > AI_LIVE.maxSegmentMs) {
        aiLiveStopAndEnqueue("max");
        state.aiCaptureVoiceStreak = 2; // allow immediate restart
        aiLiveStartRecording(now);
      }
    }
    return;
  }

  // no voice
  setAiLiveSpeakerUI(state.aiCurrentSpeakerKey, bestLevel);
  state.aiCaptureVoiceStreak = 0;
  if (state.aiCaptureRecording && now - state.aiCaptureLastVoiceAt > AI_LIVE.silenceMs) {
    aiLiveStopAndEnqueue("silence");
  }
}

function setWsState(text) {
  els.wsState.textContent = text;
}

function setIceState(text) {
  els.iceState.textContent = text;
}

function isSecure() {
  return window.isSecureContext || location.hostname === "localhost" || location.hostname === "127.0.0.1";
}

function updateSecureBadge() {
  if (isSecure()) {
    els.secureBadge.textContent = "Secure";
    els.secureBadge.classList.remove("badge--warn");
    els.insecureWarning.classList.add("hidden");
  } else {
    els.secureBadge.textContent = "Not Secure";
    els.secureBadge.classList.add("badge--warn");
    els.insecureWarning.classList.remove("hidden");
  }
}

function loadSession() {
  const raw = localStorage.getItem("ms_session");
  if (!raw) return;
  try {
    const obj = JSON.parse(raw);
    state.token = obj?.token || null;
    state.user = obj?.user || null;
  } catch {
    localStorage.removeItem("ms_session");
  }
}

function saveSession() {
  if (!state.token || !state.user) return;
  localStorage.setItem("ms_session", JSON.stringify({ token: state.token, user: state.user }));
}

function clearSession() {
  localStorage.removeItem("ms_session");
  state.token = null;
  state.user = null;
}

async function apiFetch(path, { method = "GET", headers = {}, body, auth = true, csrf = false, timeoutMs = 10000 } = {}) {
  const h = new Headers(headers);
  if (!h.has("Content-Type") && body != null) h.set("Content-Type", "application/json");
  if (!h.has("Accept")) h.set("Accept", "application/json");
  if (auth && state.token) h.set("Authorization", `Bearer ${state.token}`);
  if (csrf) {
    const token = await getCsrfToken();
    h.set("X-CSRF-Token", token);
  }

  const controller = typeof AbortController !== "undefined" ? new AbortController() : null;
  const timer = controller && timeoutMs ? setTimeout(() => controller.abort(), timeoutMs) : null;

  let res;
  try {
    res = await fetch(path, {
      method,
      headers: h,
      body: body == null ? undefined : typeof body === "string" ? body : JSON.stringify(body),
      signal: controller ? controller.signal : undefined,
    });
  } catch (err) {
    const e = new Error(err?.name === "AbortError" ? "请求超时" : "网络请求失败");
    e.status = 0;
    e.cause = err;
    throw e;
  } finally {
    if (timer) clearTimeout(timer);
  }

  const text = await res.text();
  let json;
  try {
    json = text ? JSON.parse(text) : null;
  } catch {
    json = null;
  }

  if (!res.ok) {
    const msg = json?.message || json?.error || `${res.status} ${res.statusText}`;
    const err = new Error(msg);
    err.status = res.status;
    err.payload = json;
    throw err;
  }

  return json;
}

async function getCsrfToken() {
  if (state.csrfToken) return state.csrfToken;
  const json = await apiFetch("/api/v1/csrf-token", { auth: false });
  const token = json?.data?.csrf_token;
  if (!token) throw new Error("CSRF token 获取失败");
  state.csrfToken = token;
  return token;
}

async function refreshProfile() {
  const json = await apiFetch("/api/v1/users/profile", { auth: true });
  const profile = json?.data;
  if (!profile) throw new Error("用户资料获取失败");
  state.user = profile;
  saveSession();
}

function showAuthedUI() {
  els.authCard.classList.add("hidden");
  els.mainCard.classList.remove("hidden");
  els.logoutBtn.classList.remove("hidden");
  els.currentUser.textContent = `${state.user?.username || "-"} (#${state.user?.id || "-"})`;

  if (!state.aiHealthChecked) {
    state.aiHealthChecked = true;
    aiHealthCheck().catch(() => null);
  }
}

function showUnauthedUI() {
  els.mainCard.classList.add("hidden");
  els.authCard.classList.remove("hidden");
  els.logoutBtn.classList.add("hidden");
  els.currentUser.textContent = "-";
  els.currentMeeting.textContent = "-";
  state.aiHealthChecked = false;
  aiClearOutput();
  stopAiLive().catch(() => null);
  clearAiLiveLog();
}

function setAuthTab(tab) {
  const isLogin = tab === "login";
  els.tabLogin.classList.toggle("tab--active", isLogin);
  els.tabRegister.classList.toggle("tab--active", !isLogin);
  els.tabLogin.setAttribute("aria-selected", String(isLogin));
  els.tabRegister.setAttribute("aria-selected", String(!isLogin));
  els.loginForm.classList.toggle("hidden", !isLogin);
  els.registerForm.classList.toggle("hidden", isLogin);
  setAuthMsg("");
}

function openDrawer(name) {
  const drawers = {
    panel: els.panelDrawer,
    fx: els.fxDrawer,
    ai: els.aiDrawer,
  };
  Object.entries(drawers).forEach(([key, el]) => {
    if (!el) return;
    el.classList.toggle("open", key === name);
  });
}

function setupPanelToggles() {
  const bindToggle = (btn, target, showLabel, hideLabel) => {
    if (!btn || !target) return;
    btn.addEventListener("click", () => {
      const collapsed = target.classList.toggle("collapsed");
      btn.textContent = collapsed ? showLabel : hideLabel;
    });
  };

  // 美颜面板仅折叠参数区域，保留标题和开关
  bindToggle(els.toggleFxPanel, els.fxPanelContent, "展开美颜", "收起");
}

function setupDrawers() {
  if (els.openPanelDrawer) els.openPanelDrawer.addEventListener("click", () => openDrawer("panel"));
  if (els.openFxDrawer) els.openFxDrawer.addEventListener("click", () => openDrawer("fx"));
  if (els.openAiDrawer) els.openAiDrawer.addEventListener("click", () => openDrawer("ai"));
  if (els.closePanelDrawer) els.closePanelDrawer.addEventListener("click", () => openDrawer(null));
  if (els.closeFxDrawer) els.closeFxDrawer.addEventListener("click", () => openDrawer(null));
  if (els.closeAiDrawer) els.closeAiDrawer.addEventListener("click", () => openDrawer(null));
  // 默认显示控制台，便于快速创建/加入会议
  openDrawer("panel");
}

function ensureLocalPeerId() {
  state.peerId = state.peerId || `web_${uuid().slice(0, 8)}`;
  return state.peerId;
}

function addChatLine(username, text, kind = "normal") {
  const p = document.createElement("div");
  p.className = "chatline";
  const safeUser = username || "system";
  const safeText = text || "";
  p.innerHTML = `<strong>${escapeHtml(safeUser)}</strong>：${escapeHtml(safeText)}`;
  if (kind === "error") p.style.color = "rgba(255,77,79,0.92)";
  els.chatLog.appendChild(p);
  els.chatLog.scrollTop = els.chatLog.scrollHeight;
}

function escapeHtml(s) {
  return String(s).replaceAll("&", "&amp;").replaceAll("<", "&lt;").replaceAll(">", "&gt;").replaceAll('"', "&quot;").replaceAll("'", "&#039;");
}

function renderParticipants(participants) {
  els.participants.innerHTML = "";
  for (const p of participants) {
    const li = document.createElement("li");
    const left = document.createElement("div");
    left.textContent = `${p.username || `user_${p.user_id}`} (#${p.user_id})`;
    const badge = document.createElement("span");
    badge.className = `pill ${p.is_self ? "pill--self" : ""}`;
    badge.textContent = p.is_self ? "我" : "在线";
    li.appendChild(left);
    li.appendChild(badge);
    els.participants.appendChild(li);
  }
}

function getWsBaseUrl() {
  const protocol = location.protocol === "https:" ? "wss" : "ws";
  return `${protocol}://${location.host}`;
}

function wsSend(message) {
  if (!state.ws || !state.wsReady) return;
  state.ws.send(JSON.stringify(message));
}

async function waitForRoomIceServers(timeoutMs = 2500) {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    if (state.roomIceServers && state.roomIceServers.length > 0) return state.roomIceServers;
    await sleep(100);
  }
  return state.roomIceServers;
}

async function connectWebSocket() {
  if (!state.user?.id || !state.meetingId || !state.token) throw new Error("缺少 user/meeting/token");
  disconnectWebSocket();

  ensureLocalPeerId();
  const qs = new URLSearchParams({
    user_id: String(state.user.id),
    meeting_id: String(state.meetingId),
    peer_id: String(state.peerId),
    token: state.token,
  });
  const url = `${getWsBaseUrl()}/ws/signaling?${qs.toString()}`;

  setWsState("connecting");
  state.wsReady = false;
  state.wsCloseExpected = false;
  const ws = new WebSocket(url);
  state.ws = ws;

  const connectTimer = setTimeout(() => {
    if (state.ws !== ws) return;
    if (ws.readyState === WebSocket.CONNECTING) {
      setWsState("timeout");
      setHint(els.joinResult, "信令连接超时，请刷新重试", "error");
      try {
        ws.close();
      } catch {
        // ignore
      }
    }
  }, 8000);

  ws.onopen = () => {
    if (state.ws !== ws) return;
    clearTimeout(connectTimer);
    state.wsReady = true;
    setWsState("open");
    setHint(els.joinResult, "信令已连接", "ok");
    wsSend({
      id: `join_${uuid()}`,
      type: WS_TYPES.JOIN,
      peer_id: state.peerId,
      payload: {
        meeting_id: Number(state.meetingId),
        user_id: Number(state.user.id),
        peer_id: state.peerId,
      },
      timestamp: new Date().toISOString(),
    });
    els.leaveBtn.disabled = false;
    updateCallButtons();
  };

  ws.onclose = (ev) => {
    if (state.ws !== ws) return;
    clearTimeout(connectTimer);
    state.wsReady = false;
    setWsState("closed");
    setIceState("-");
    els.leaveBtn.disabled = true;
    updateCallButtons();
    stopAiLive().catch(() => null);

    const expected = state.wsCloseExpected;
    state.wsCloseExpected = false;
    state.ws = null;

    if (!expected && state.meetingId) {
      const code = ev?.code ? `（code ${ev.code}）` : "";
      setHint(els.joinResult, `信令连接已断开${code}`, "error");
    }
  };

  ws.onerror = () => {
    if (state.ws !== ws) return;
    setWsState("error");
    if (state.meetingId) setHint(els.joinResult, "信令连接失败（请检查网络/证书/服务状态）", "error");
  };

  ws.onmessage = async (ev) => {
    if (state.ws !== ws) return;
    let msg;
    try {
      msg = JSON.parse(ev.data);
    } catch {
      return;
    }

    const t = msg?.type;
    if (!t) return;

    if (t === WS_TYPES.AI_LIVE_STATUS) {
      applyAiLiveStatus(msg.payload);
      return;
    }

    if (t === WS_TYPES.AI_LIVE_RESULT) {
      applyAiLiveResult(msg.payload);
      return;
    }

    if (t === WS_TYPES.ROOM_INFO) {
      const info = msg.payload;
      state.sessionId = info?.session_id || msg?.session_id || state.sessionId;
      if (info?.ai_live) applyAiLiveStatus(info.ai_live);
      const participants = (info?.participants || []).map((p) => ({
        user_id: p.user_id,
        username: p.username,
        session_id: p.session_id,
        peer_id: p.peer_id,
        is_self: p.is_self,
      }));
      renderParticipants(participants);
      state.participantsByPeerId = new Map();
      for (const p of participants) {
        if (p?.peer_id) state.participantsByPeerId.set(p.peer_id, p);
      }

      // 刷新已存在的远端视频标签（peer_id -> username）
      for (const [streamKey, stream] of state.remoteStreams.entries()) {
        if (stream?.getVideoTracks?.().length > 0) {
          ensureVideoTile(`sfu_${streamKey}`, getSpeakerLabel(streamKey), stream, { muted: false });
        }
      }

      state.roomIceServers = (info?.ice_servers || []).map((s) => ({
        urls: s.urls,
        username: s.username || undefined,
        credential: s.credential || undefined,
      }));

      addChatLine("system", `房间同步：${participants.length} 人`);
      connectSfu().catch((err) => addChatLine("error", err?.message || "媒体连接失败", "error"));
      return;
    }

    if (t === WS_TYPES.USER_JOINED) {
      const p = msg?.payload;
      const remoteUserId = Number(p?.user_id || msg?.from_user_id || 0);
      if (!remoteUserId || remoteUserId === state.user.id) return;
      addChatLine("system", `${p?.username || `user_${remoteUserId}`} 加入了会议`);
      return;
    }

    if (t === WS_TYPES.USER_LEFT) {
      const p = msg?.payload;
      const remoteUserId = Number(p?.user_id || msg?.from_user_id || 0);
      if (!remoteUserId) return;
      addChatLine("system", `${p?.username || `user_${remoteUserId}`} 离开了会议`);
      return;
    }

    if (t === WS_TYPES.CHAT) {
      const p = msg?.payload;
      const username = p?.username || `user_${msg?.from_user_id || 0}`;
      addChatLine(username, p?.content || "");
      return;
    }

    if (t === WS_TYPES.ERROR) {
      const p = msg?.payload;
      addChatLine("error", p?.message || "WebSocket error", "error");
      return;
    }
  };
}

function disconnectWebSocket() {
  if (!state.ws) return;
  state.wsCloseExpected = true;
  try {
    state.ws.close();
  } catch {
    // ignore
  }
  state.ws = null;
  state.wsReady = false;
  state.sessionId = null;
  setWsState("-");
}

function disconnectSfu() {
  if (state.sfuOfferPoll) {
    clearInterval(state.sfuOfferPoll);
    state.sfuOfferPoll = null;
  }
  if (state.sfuIcePoll) {
    clearInterval(state.sfuIcePoll);
    state.sfuIcePoll = null;
  }
  if (state.sfuReconnectTimer) {
    clearTimeout(state.sfuReconnectTimer);
    state.sfuReconnectTimer = null;
  }

  for (const streamId of state.remoteStreams.keys()) {
    removeVideoTile(`sfu_${streamId}`);
    aiLiveRemoveInput(streamId);
  }
  state.remoteStreams.clear();
  for (const key of state.remoteAudioEls.keys()) {
    removeRemoteAudio(key);
  }

  state.sfuOfferInFlight = false;
  state.sfuIceInFlight = false;
  state.sfuPendingLocalCandidates = [];
  state.sfuLastOfferSdp = null;
  state.sfuPeerId = null;
  state.sfuAudioTransceiver = null;
  state.sfuVideoTransceiver = null;

  if (state.sfuPc) {
    try {
      state.sfuPc.close();
    } catch {
      // ignore
    }
  }
  state.sfuPc = null;
  state.fxSenderAttached = false;
  state.fxReceiverAttached = new Set();
  state.fxLastInjectedVersion = 0;
  state.fxLastInjectedAt = 0;
  state.fxRemote.clear();
  setIceState("-");
}

function scheduleSfuReconnect(reason) {
  if (!state.meetingId || !state.roomId || !state.user?.id) return;
  const now = Date.now();
  if (now - state.sfuLastReconnectAt < 3000) return;
  state.sfuLastReconnectAt = now;

  if (state.sfuReconnectTimer) return;
  addChatLine("system", `${reason || "媒体连接异常"}，正在重连…`);

  state.sfuReconnectTimer = setTimeout(() => {
    state.sfuReconnectTimer = null;
    disconnectSfu();
    connectSfu().catch((err) => addChatLine("error", err?.message || "媒体重连失败", "error"));
  }, 500);
}

function ensureRemoteAudio(key, stream) {
  if (!key || !stream) return;
  let audio = state.remoteAudioEls.get(key);
  if (!audio) {
    audio = document.createElement("audio");
    audio.autoplay = true;
    audio.playsInline = true;
    audio.controls = false;
    audio.style.display = "none";
    document.body.appendChild(audio);
    state.remoteAudioEls.set(key, audio);
  }
  audio.muted = false;
  audio.srcObject = stream;
  audio.play?.().catch(() => null);
}

function removeRemoteAudio(key) {
  const audio = state.remoteAudioEls.get(key);
  if (!audio) return;
  try {
    audio.srcObject = null;
  } catch {
    // ignore
  }
  try {
    audio.remove();
  } catch {
    // ignore
  }
  state.remoteAudioEls.delete(key);
}

async function sendSfuIceCandidate(candidate) {
  if (!state.sfuPeerId) return;
  try {
    await apiFetch("/api/v1/webrtc/ice-candidate", {
      method: "POST",
      timeoutMs: 2500,
      body: {
        peer_id: state.sfuPeerId,
        candidate,
      },
    });
  } catch {
    // ignore
  }
}

async function pollSfuIceCandidatesOnce() {
  if (!state.sfuPc || !state.sfuPeerId) return;
  if (state.sfuIceInFlight) return;
  state.sfuIceInFlight = true;
  try {
    const json = await apiFetch(`/api/v1/webrtc/peer/${encodeURIComponent(state.sfuPeerId)}/ice-candidates`, { auth: false, timeoutMs: 2500 });
    const candidates = json?.candidates || [];
    for (const cand of candidates) {
      try {
        await state.sfuPc.addIceCandidate(new RTCIceCandidate(cand));
      } catch {
        // ignore
      }
    }

    const complete = Boolean(json?.complete);
    if (complete && (!candidates.length || state.sfuPc.iceConnectionState === "connected" || state.sfuPc.iceConnectionState === "completed")) {
      clearInterval(state.sfuIcePoll);
      state.sfuIcePoll = null;
    }
  } catch (err) {
    if (err?.status === 404 && String(err?.message || "").includes("peer not found")) scheduleSfuReconnect("SFU peer not found");
  } finally {
    state.sfuIceInFlight = false;
  }
}

async function pollSfuOfferOnce() {
  if (!state.sfuPc || !state.sfuPeerId) return;
  if (state.sfuOfferInFlight) return;
  state.sfuOfferInFlight = true;
  try {
    const json = await apiFetch(`/api/v1/webrtc/peer/${encodeURIComponent(state.sfuPeerId)}/offer`, { auth: false, timeoutMs: 2500 });
    const offer = json?.offer;
    if (!offer?.sdp) return;

    if (state.sfuLastOfferSdp && offer.sdp === state.sfuLastOfferSdp) return;

    await state.sfuPc.setRemoteDescription(new RTCSessionDescription(offer));
    const answer = await state.sfuPc.createAnswer();
    await state.sfuPc.setLocalDescription(answer);
    await apiFetch(`/api/v1/webrtc/peer/${encodeURIComponent(state.sfuPeerId)}/answer`, {
      method: "POST",
      auth: false,
      body: {
        answer: {
          type: state.sfuPc.localDescription.type,
          sdp: state.sfuPc.localDescription.sdp,
        },
      },
    });

    state.sfuLastOfferSdp = offer.sdp;
  } catch (err) {
    if (err?.status === 404 && String(err?.message || "").includes("peer not found")) scheduleSfuReconnect("SFU peer not found");
  } finally {
    state.sfuOfferInFlight = false;
  }
}

function u8Concat(chunks) {
  let len = 0;
  for (const c of chunks) len += c.length;
  const out = new Uint8Array(len);
  let offset = 0;
  for (const c of chunks) {
    out.set(c, offset);
    offset += c.length;
  }
  return out;
}

function encodeSeiHeader(payloadType, payloadSize) {
  const bytes = [];
  let t = payloadType;
  while (t >= 255) {
    bytes.push(255);
    t -= 255;
  }
  bytes.push(t);
  let s = payloadSize;
  while (s >= 255) {
    bytes.push(255);
    s -= 255;
  }
  bytes.push(s);
  return new Uint8Array(bytes);
}

function h264EscapeRbsp(data) {
  const out = [];
  let zeroCount = 0;
  for (let i = 0; i < data.length; i++) {
    const b = data[i];
    if (zeroCount >= 2 && b <= 3) {
      out.push(3);
      zeroCount = 0;
    }
    out.push(b);
    if (b === 0) zeroCount += 1;
    else zeroCount = 0;
  }
  return new Uint8Array(out);
}

function h264UnescapeRbsp(data) {
  const out = [];
  let zeroCount = 0;
  for (let i = 0; i < data.length; i++) {
    const b = data[i];
    if (zeroCount >= 2 && b === 3) {
      continue;
    }
    out.push(b);
    if (b === 0) zeroCount += 1;
    else zeroCount = 0;
  }
  return new Uint8Array(out);
}

function buildFxSeiNal() {
  const payloadObj = {
    v: 1,
    eb: Boolean(state.fxBeautyEnabled),
    ef: Boolean(state.fxFilterEnabled),
    bt: String(state.fxBeautyType || "natural"),
    b: state.fxBeautyEnabled ? state.fxBeauty : 0,
    s: state.fxSlim || 0,
    f: state.fxFilterEnabled ? state.fxFilter : "none",
    t: Date.now(),
  };
  const userData = utf8Encode(JSON.stringify(payloadObj));
  const payload = u8Concat([getFxSeiUuidBytes(), userData]);
  const header = encodeSeiHeader(5, payload.length);
  const rbsp = u8Concat([header, payload, new Uint8Array([0x80])]);
  const escaped = h264EscapeRbsp(rbsp);
  const nal = new Uint8Array(1 + escaped.length);
  nal[0] = 0x06;
  nal.set(escaped, 1);
  return nal;
}

function findAnnexBNalUnits(data) {
  const starts = [];
  for (let i = 0; i + 3 < data.length; i++) {
    if (data[i] === 0 && data[i + 1] === 0 && data[i + 2] === 1) {
      starts.push({ index: i, len: 3 });
      i += 2;
      continue;
    }
    if (data[i] === 0 && data[i + 1] === 0 && data[i + 2] === 0 && data[i + 3] === 1) {
      starts.push({ index: i, len: 4 });
      i += 3;
    }
  }
  if (starts.length === 0) return null;
  const units = [];
  for (let i = 0; i < starts.length; i++) {
    const s = starts[i];
    const next = i + 1 < starts.length ? starts[i + 1].index : data.length;
    const nalStart = s.index + s.len;
    if (nalStart >= next) continue;
    units.push({ start: s.index, nalStart, nalEnd: next });
  }
  return units.length ? units : null;
}

function parseAvccNalUnits(data) {
  const units = [];
  let offset = 0;
  while (offset + 4 <= data.length) {
    const recordStart = offset;
    const nalLen = ((data[offset] << 24) | (data[offset + 1] << 16) | (data[offset + 2] << 8) | data[offset + 3]) >>> 0;
    offset += 4;
    if (nalLen === 0 || offset + nalLen > data.length) return null;
    const nalStart = offset;
    const nalEnd = offset + nalLen;
    units.push({ recordStart, nalStart, nalEnd });
    offset = nalEnd;
  }
  if (!units.length || offset !== data.length) return null;
  return units;
}

function detectH264FrameFormat(data) {
  const annexb = findAnnexBNalUnits(data);
  if (annexb) {
    const first = annexb[0];
    const t = data[first.nalStart] & 0x1f;
    if (t > 0 && t < 24) return "annexb";
  }
  const avcc = parseAvccNalUnits(data);
  if (avcc) {
    const first = avcc[0];
    const t = data[first.nalStart] & 0x1f;
    if (t > 0 && t < 24) return "avcc";
  }
  return null;
}

function injectSeiIntoAnnexB(frame, seiNal) {
  const units = findAnnexBNalUnits(frame);
  if (!units) return null;
  let insertOffset = frame.length;
  for (const u of units) {
    const t = frame[u.nalStart] & 0x1f;
    if (t >= 1 && t <= 5) {
      insertOffset = u.start;
      break;
    }
  }
  const startCode = new Uint8Array([0, 0, 0, 1]);
  return u8Concat([frame.slice(0, insertOffset), startCode, seiNal, frame.slice(insertOffset)]);
}

function injectSeiIntoAvcc(frame, seiNal) {
  const units = parseAvccNalUnits(frame);
  if (!units) return null;
  let insertOffset = frame.length;
  for (const u of units) {
    const t = frame[u.nalStart] & 0x1f;
    if (t >= 1 && t <= 5) {
      insertOffset = u.recordStart;
      break;
    }
  }
  const n = seiNal.length >>> 0;
  const lenPrefix = new Uint8Array([(n >>> 24) & 255, (n >>> 16) & 255, (n >>> 8) & 255, n & 255]);
  return u8Concat([frame.slice(0, insertOffset), lenPrefix, seiNal, frame.slice(insertOffset)]);
}

function injectSeiIntoH264Frame(frame, seiNal) {
  const fmt = detectH264FrameFormat(frame);
  if (fmt === "annexb") return injectSeiIntoAnnexB(frame, seiNal);
  if (fmt === "avcc") return injectSeiIntoAvcc(frame, seiNal);
  return null;
}

function parseFxUserDataFromSeiNal(nal) {
  if (!nal || nal.length < 2) return null;
  if ((nal[0] & 0x1f) !== 6) return null;
  const rbsp = h264UnescapeRbsp(nal.subarray(1));

  let offset = 0;
  while (offset < rbsp.length) {
    if (rbsp[offset] === 0x80) return null;

    let payloadType = 0;
    while (offset < rbsp.length && rbsp[offset] === 255) {
      payloadType += 255;
      offset += 1;
    }
    if (offset >= rbsp.length) return null;
    payloadType += rbsp[offset++];

    let payloadSize = 0;
    while (offset < rbsp.length && rbsp[offset] === 255) {
      payloadSize += 255;
      offset += 1;
    }
    if (offset >= rbsp.length) return null;
    payloadSize += rbsp[offset++];

    if (offset + payloadSize > rbsp.length) return null;

    if (payloadType === 5 && payloadSize >= 16) {
      const seiUuid = rbsp.subarray(offset, offset + 16);
      if (bytesEqual(seiUuid, getFxSeiUuidBytes())) {
        return rbsp.subarray(offset + 16, offset + payloadSize);
      }
    }

    offset += payloadSize;
  }
  return null;
}

function extractFxFromH264Frame(frame) {
  const fmt = detectH264FrameFormat(frame);
  const units = fmt === "annexb" ? findAnnexBNalUnits(frame) : fmt === "avcc" ? parseAvccNalUnits(frame) : null;
  if (!units) return null;

  for (const u of units) {
    const nal = frame.subarray(u.nalStart, u.nalEnd);
    if (!nal.length) continue;
    if ((nal[0] & 0x1f) !== 6) continue;
    const userData = parseFxUserDataFromSeiNal(nal);
    if (!userData) continue;
    const text = utf8Decode(userData);
    try {
      return JSON.parse(text);
    } catch {
      return null;
    }
  }

  return null;
}

function shouldInjectFxSei(encodedFrame) {
  const now = Date.now();
  if (state.fxVersion !== state.fxLastInjectedVersion) return true;
  if (encodedFrame?.type === "key" && now - state.fxLastInjectedAt > 1500) return true;
  return false;
}

function attachFxSeiToSender(sender) {
  if (!state.fxSeiReady || state.fxSenderAttached) return;
  if (!sender || typeof sender.createEncodedStreams !== "function") return;

  let streams;
  try {
    streams = sender.createEncodedStreams();
  } catch {
    return;
  }

  state.fxSenderAttached = true;

  const transformer = new TransformStream({
    transform: (encodedFrame, controller) => {
      try {
        if (shouldInjectFxSei(encodedFrame)) {
          const frame = new Uint8Array(encodedFrame.data);
          const out = injectSeiIntoH264Frame(frame, buildFxSeiNal());
          if (out) {
            encodedFrame.data = out.buffer;
            state.fxLastInjectedVersion = state.fxVersion;
            state.fxLastInjectedAt = Date.now();
          }
        }
      } catch {
        // ignore
      }
      controller.enqueue(encodedFrame);
    },
  });

  streams.readable.pipeThrough(transformer).pipeTo(streams.writable).catch(() => null);
}

function attachFxSeiToReceiver(receiver, tileKey) {
  if (!state.fxSeiReady) return;
  if (!receiver || typeof receiver.createEncodedStreams !== "function") return;
  if (!tileKey || state.fxReceiverAttached.has(tileKey)) return;

  let streams;
  try {
    streams = receiver.createEncodedStreams();
  } catch {
    return;
  }

  state.fxReceiverAttached.add(tileKey);

  const transformer = new TransformStream({
    transform: (encodedFrame, controller) => {
      try {
        const fx = extractFxFromH264Frame(new Uint8Array(encodedFrame.data));
        if (fx) setRemoteFx(tileKey, fx);
      } catch {
        // ignore
      }
      controller.enqueue(encodedFrame);
    },
  });

  streams.readable.pipeThrough(transformer).pipeTo(streams.writable).catch(() => null);
}

function preferH264ForTransceiver(transceiver) {
  try {
    if (!transceiver?.setCodecPreferences || typeof RTCRtpSender === "undefined" || !RTCRtpSender.getCapabilities) return;
    const caps = RTCRtpSender.getCapabilities("video");
    const codecs = caps?.codecs || [];
    const h264 = codecs.filter((c) => String(c.mimeType || "").toLowerCase() === "video/h264");
    if (!h264.length) return;
    const rest = codecs.filter((c) => String(c.mimeType || "").toLowerCase() !== "video/h264");
    transceiver.setCodecPreferences([...h264, ...rest]);
  } catch {
    // ignore
  }
}

async function applyLocalTracksToSfu() {
  if (!state.sfuPc) return;
  const pc = state.sfuPc;

  const audioTrack = state.localStream?.getAudioTracks?.()[0] || null;
  const processedCamTrack = await ensureFxPipelineForCamera().catch(() => null);
  const videoTrack = state.screenStream?.getVideoTracks?.()[0] || processedCamTrack || state.fxProcessedTrack || state.localStream?.getVideoTracks?.()[0] || null;

  if (state.sfuAudioTransceiver?.sender) {
    try {
      await state.sfuAudioTransceiver.sender.replaceTrack(audioTrack);
    } catch {
      // ignore
    }
  } else if (audioTrack && state.localStream) {
    try {
      pc.addTrack(audioTrack, state.localStream);
    } catch {
      // ignore
    }
  }

  if (state.sfuVideoTransceiver?.sender) {
    try {
      await state.sfuVideoTransceiver.sender.replaceTrack(videoTrack);
    } catch {
      // ignore
    }
  } else if (videoTrack) {
    try {
      pc.addTrack(videoTrack, state.screenStream || state.localStream || new MediaStream([videoTrack]));
    } catch {
      // ignore
    }
  }

  const videoSender = state.sfuVideoTransceiver?.sender || pc.getSenders().find((s) => s.track && s.track.kind === "video");
  if (videoSender) attachFxSeiToSender(videoSender);
}

async function connectSfu() {
  if (!state.user?.id || !state.meetingId || !state.roomId) throw new Error("缺少 user/meeting/room_id");
  if (state.sfuPc) return;

  await waitForRoomIceServers().catch(() => null);
  const iceServers =
    state.roomIceServers?.length > 0
      ? state.roomIceServers
      : [
          { urls: `stun:${location.hostname}:3478` },
          { urls: "stun:stun.l.google.com:19302" },
        ];

  const pc = new RTCPeerConnection({ iceServers, encodedInsertableStreams: true });
  state.sfuPc = pc;
  state.sfuPeerId = null;
  state.sfuAudioTransceiver = null;
  state.sfuVideoTransceiver = null;
  state.sfuPendingLocalCandidates = [];
  state.sfuLastOfferSdp = null;

  pc.oniceconnectionstatechange = () => {
    setIceState(pc.iceConnectionState);
  };

  pc.onicecandidate = (ev) => {
    if (!ev.candidate) return;
    const cand = ev.candidate.toJSON ? ev.candidate.toJSON() : ev.candidate;
    if (state.sfuPeerId) {
      sendSfuIceCandidate(cand);
    } else {
      state.sfuPendingLocalCandidates.push(cand);
    }
  };

  pc.ontrack = (ev) => {
    const incomingStream = ev.streams && ev.streams[0] ? ev.streams[0] : null;
    const key = incomingStream?.id || (ev.transceiver?.mid ? `mid_${ev.transceiver.mid}` : `track_${ev.track.id}`);

    let stream = state.remoteStreams.get(key);
    if (!stream) {
      stream = incomingStream || new MediaStream();
      state.remoteStreams.set(key, stream);
    }

    if (!stream.getTracks().some((t) => t.id === ev.track.id)) {
      try {
        stream.addTrack(ev.track);
      } catch {
        // ignore
      }
    }

    // 注册为实时 AI 输入（仅当包含音频轨道）
    aiLiveUpsertInput(key, stream);

    if (ev.track.kind === "video") {
      attachFxSeiToReceiver(ev.receiver, `sfu_${key}`);
    }

    const hasVideo = stream.getVideoTracks().length > 0;
    if (hasVideo) {
      removeRemoteAudio(key);
      ensureVideoTile(`sfu_${key}`, getSpeakerLabel(key), stream, { muted: false });
    } else {
      ensureRemoteAudio(key, stream);
    }

    ev.track.onended = () => {
      const s = state.remoteStreams.get(key);
      if (!s) return;
      const live = s.getTracks().some((t) => t.readyState === "live");
      if (!live) {
        state.remoteStreams.delete(key);
        aiLiveRemoveInput(key);
        removeVideoTile(`sfu_${key}`);
        removeRemoteAudio(key);
        clearRemoteFx(`sfu_${key}`);
        state.fxReceiverAttached.delete(`sfu_${key}`);
      }
    };
  };

  try {
    state.sfuAudioTransceiver = pc.addTransceiver("audio", { direction: "sendrecv" });
  } catch {
    state.sfuAudioTransceiver = null;
  }
  try {
    state.sfuVideoTransceiver = pc.addTransceiver("video", { direction: "sendrecv" });
  } catch {
    state.sfuVideoTransceiver = null;
  }
  preferH264ForTransceiver(state.sfuVideoTransceiver);

  await applyLocalTracksToSfu();

  const offer = await pc.createOffer();
  await pc.setLocalDescription(offer);

  const resp = await apiFetch("/api/v1/webrtc/answer", {
    method: "POST",
    auth: false,
    body: {
      room_id: String(state.roomId),
      user_id: String(state.user.id),
      offer: {
        type: pc.localDescription.type,
        sdp: pc.localDescription.sdp,
      },
    },
  });

  const peerId = resp?.peer_id;
  const answer = resp?.answer;
  if (!peerId || !answer?.sdp) throw new Error("SFU 响应缺少 peer_id/answer");

  state.sfuPeerId = peerId;
  await pc.setRemoteDescription(new RTCSessionDescription(answer));

  for (const cand of state.sfuPendingLocalCandidates.splice(0)) {
    await sendSfuIceCandidate(cand);
  }

  if (!state.sfuIcePoll) state.sfuIcePoll = setInterval(() => pollSfuIceCandidatesOnce().catch(() => null), 600);
  if (!state.sfuOfferPoll) state.sfuOfferPoll = setInterval(() => pollSfuOfferOnce().catch(() => null), 800);

  addChatLine("system", "媒体已连接（SFU）");
}

function ensureVideoTile(key, label, stream, { muted = false } = {}) {
  const existing = document.getElementById(`tile_${key}`);
  const tile = existing || document.createElement("div");
  tile.id = `tile_${key}`;
  tile.className = "tile";

  let video = tile.querySelector("video");
  if (!video) {
    video = document.createElement("video");
    video.playsInline = true;
    video.autoplay = true;
    tile.appendChild(video);
  }
  video.muted = muted;
  video.srcObject = stream;
  applyFxToTile(key);
  video.play?.().catch(() => null);

  let lab = tile.querySelector(".tile__label");
  if (!lab) {
    lab = document.createElement("div");
    lab.className = "tile__label";
    tile.appendChild(lab);
  }
  lab.textContent = label;

  if (!existing) els.videoGrid.appendChild(tile);
}

function removeVideoTile(key) {
  const el = document.getElementById(`tile_${key}`);
  if (el) el.remove();
}

async function startLocalMedia() {
  if (state.localStream) return state.localStream;
  if (!isSecure()) throw new Error("请使用 HTTPS 访问以启用摄像头/麦克风");

  const stream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
  state.localStream = stream;
  const audioTrack = state.localStream?.getAudioTracks?.()[0] || null;
  const processedTrack = (await ensureFxPipelineForCamera().catch(() => null)) || state.localStream.getVideoTracks()[0];

  const previewStream = new MediaStream();
  if (processedTrack) previewStream.addTrack(processedTrack);
  if (audioTrack) previewStream.addTrack(audioTrack);

  aiLiveUpsertInput("local", stream);
  ensureVideoTile("local", "我（本地）", previewStream, { muted: true });
  updateCallButtons();
  await applyLocalTracksToSfu();
  return stream;
}

function stopLocalMedia() {
  if (state.localStream) {
    for (const t of state.localStream.getTracks()) t.stop();
  }
  stopFxPipeline();
  state.localStream = null;
  aiLiveRemoveInput("local");
  removeVideoTile("local");
  updateCallButtons();
}

function updateCallButtons() {
  const hasLocal = !!state.localStream;
  els.muteBtn.disabled = !hasLocal;
  els.videoBtn.disabled = !hasLocal;
  els.screenBtn.disabled = !hasLocal;
  els.leaveBtn.disabled = !state.meetingId;

  if (!hasLocal) {
    els.muteBtn.textContent = "静音";
    els.videoBtn.textContent = "关摄像头";
    els.screenBtn.textContent = "共享屏幕";
  }
}

function toggleMute() {
  if (!state.localStream) return;
  const a = state.localStream.getAudioTracks()[0];
  if (!a) return;
  a.enabled = !a.enabled;
  els.muteBtn.textContent = a.enabled ? "静音" : "取消静音";
}

function toggleVideo() {
  if (!state.localStream) return;
  const v = state.localStream.getVideoTracks()[0];
  if (!v) return;
  v.enabled = !v.enabled;
  els.videoBtn.textContent = v.enabled ? "关摄像头" : "开摄像头";
}

async function toggleScreenShare() {
  if (!state.localStream) return;
  if (state.screenStream) {
    stopScreenShare();
    return;
  }

  const stream = await navigator.mediaDevices.getDisplayMedia({ video: true });
  state.screenStream = stream;
  els.screenBtn.textContent = "停止共享";

  const screenTrack = stream.getVideoTracks()[0];
  screenTrack.onended = () => stopScreenShare();

  await replaceVideoTrack(screenTrack);
}

async function stopScreenShare() {
  if (!state.screenStream) return;
  for (const t of state.screenStream.getTracks()) t.stop();
  state.screenStream = null;
  els.screenBtn.textContent = "共享屏幕";

  const camTrack = state.fxProcessedTrack || state.localStream?.getVideoTracks()[0];
  if (camTrack) await replaceVideoTrack(camTrack);
}

async function replaceVideoTrack(newTrack) {
  const previewStream = state.screenStream || state.localStream || (newTrack ? new MediaStream([newTrack]) : null);
  if (previewStream) ensureVideoTile("local", "我（本地）", previewStream, { muted: true });

  if (!state.sfuPc) return;
  const sender = state.sfuVideoTransceiver?.sender || state.sfuPc.getSenders().find((s) => s.track && s.track.kind === "video");
  if (!sender) return;
  try {
    await sender.replaceTrack(newTrack);
  } catch {
    // ignore
  }
}

async function createMeeting() {
  const title = (els.createTitle.value || "").trim() || "新会议";
  const meetingType = els.createType.value;
  const start = new Date(Date.now() + 2 * 60 * 1000);
  const end = new Date(Date.now() + 62 * 60 * 1000);

  setHint(els.createResult, "创建中…");
  const json = await apiFetch("/api/v1/meetings", {
    method: "POST",
    body: {
      title,
      description: "",
      start_time: start.toISOString(),
      end_time: end.toISOString(),
      max_participants: 10,
      meeting_type: meetingType,
      settings: {
        enable_video: meetingType === "video",
        enable_audio: true,
        enable_screen_share: true,
        enable_chat: true,
        enable_recording: false,
        enable_ai: false,
        mute_on_join: false,
        require_approval: false,
      },
    },
  });

  const meeting = json?.data?.meeting;
  if (!meeting?.id) throw new Error("会议创建失败（未返回 meeting.id）");
  setHint(els.createResult, `已创建会议：ID = ${meeting.id}`, "ok");
  els.meetingIdInput.value = String(meeting.id);
}

async function joinMeeting(meetingId) {
  setHint(els.joinResult, "加入中…");
  const json = await apiFetch(`/api/v1/meetings/${encodeURIComponent(meetingId)}/join`, {
    method: "POST",
    body: { password: "" },
  });
  const roomId = json?.data?.room_id;
  if (!roomId) throw new Error("加入会议成功但缺少 room_id");
  state.meetingId = Number(meetingId);
  state.roomId = String(roomId);
  els.currentMeeting.textContent = String(state.meetingId);
  setHint(els.joinResult, "已加入会议，正在连接信令…", "ok");
  await connectWebSocket();
}

async function leaveMeeting() {
  const meetingId = state.meetingId;
  const roomId = state.roomId;
  const userId = state.user?.id;

  // best-effort: 先显式通知信令/媒体服务，避免“关闭页面仍在线/僵尸轨道”。
  try {
    if (state.wsReady) {
      wsSend({
        id: `leave_${uuid()}`,
        type: WS_TYPES.LEAVE,
        peer_id: state.peerId,
        payload: { meeting_id: Number(meetingId), user_id: Number(userId || 0), peer_id: state.peerId },
        timestamp: new Date().toISOString(),
      });
    }
  } catch {
    // ignore
  }

  if (roomId && userId) {
    try {
      await apiFetch(`/api/v1/webrtc/room/${encodeURIComponent(roomId)}/leave`, {
        method: "POST",
        auth: false,
        timeoutMs: 2500,
        body: { user_id: String(userId) },
      });
    } catch {
      // ignore
    }
  }

  state.meetingId = null;
  state.roomId = null;
  state.sessionId = null;
  state.roomIceServers = [];
  els.currentMeeting.textContent = "-";
  disconnectWebSocket();
  disconnectSfu();
  await stopAiLive().catch(() => null);
  stopScreenShare();
  stopLocalMedia();
  clearAiLiveLog();
  setHint(els.joinResult, "已离开会议");

  if (meetingId && userId && state.token) {
    try {
      await apiFetch(`/api/v1/meetings/${encodeURIComponent(meetingId)}/leave`, { method: "POST", body: {} });
    } catch {
      // ignore
    }
  }
}

let didSendUnloadLeave = false;

function bestEffortLeaveOnPageHide() {
  if (didSendUnloadLeave) return;
  if (!state.meetingId || !state.roomId || !state.user?.id) return;
  didSendUnloadLeave = true;

  try {
    if (state.wsReady) {
      wsSend({
        id: `leave_${uuid()}`,
        type: WS_TYPES.LEAVE,
        peer_id: state.peerId,
        payload: { meeting_id: Number(state.meetingId), user_id: Number(state.user.id), peer_id: state.peerId },
        timestamp: new Date().toISOString(),
      });
    }
  } catch {
    // ignore
  }

  // 媒体服务离房（无需鉴权）
  try {
    fetch(`/api/v1/webrtc/room/${encodeURIComponent(state.roomId)}/leave`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ user_id: String(state.user.id) }),
      keepalive: true,
    });
  } catch {
    // ignore
  }

  // 会议服务离会（需要 JWT；使用 keepalive 尽量在 unload 时送达）
  try {
    if (state.token) {
      fetch(`/api/v1/meetings/${encodeURIComponent(state.meetingId)}/leave`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${state.token}`,
        },
        body: "{}",
        keepalive: true,
      });
    }
  } catch {
    // ignore
  }

  try {
    if (state.ws) {
      state.wsCloseExpected = true;
      state.ws.close();
    }
  } catch {
    // ignore
  }
}

async function sendChat(text) {
  if (!state.wsReady) throw new Error("信令未连接");
  const content = (text || "").trim();
  if (!content) return;
  wsSend({
    id: `chat_${uuid()}`,
    type: WS_TYPES.CHAT,
    peer_id: state.peerId,
    payload: {
      content,
      user_id: state.user.id,
      username: state.user.username,
      meeting_id: state.meetingId,
    },
    timestamp: new Date().toISOString(),
  });
  els.chatInput.value = "";
}

async function boot() {
  updateSecureBadge();
  loadSession();
  setAuthTab("login");
  initFxControls();
  setupPanelToggles();
  setupDrawers();

  // 页面关闭/刷新时做一次 best-effort 离会，减少“用户关闭页面仍在线/僵尸轨道”与卡顿。
  window.addEventListener("pagehide", bestEffortLeaveOnPageHide);
  window.addEventListener("beforeunload", bestEffortLeaveOnPageHide);

  els.tabLogin.addEventListener("click", () => setAuthTab("login"));
  els.tabRegister.addEventListener("click", () => setAuthTab("register"));

  els.loginForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    setAuthMsg("登录中…");
    try {
      const json = await apiFetch("/api/v1/auth/login", {
        method: "POST",
        auth: false,
        csrf: true,
        body: { username: els.loginUsername.value.trim(), password: els.loginPassword.value },
      });
      const data = json?.data;
      if (!data?.token || !data?.user) throw new Error("登录返回缺少 token/user");
      state.token = data.token;
      state.user = data.user;
      saveSession();
      await refreshProfile().catch(() => null);
      setAuthMsg("登录成功", "ok");
      showAuthedUI();
    } catch (err) {
      setAuthMsg(err?.message || "登录失败", "error");
    }
  });

  els.registerForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    setAuthMsg("注册中…");
    try {
      await apiFetch("/api/v1/auth/register", {
        method: "POST",
        auth: false,
        csrf: true,
        body: {
          username: els.regUsername.value.trim(),
          email: els.regEmail.value.trim(),
          password: els.regPassword.value,
          nickname: els.regNickname.value.trim(),
        },
      });
      setAuthMsg("注册成功，请切换到“登录”", "ok");
      setAuthTab("login");
    } catch (err) {
      setAuthMsg(err?.message || "注册失败", "error");
    }
  });

  els.logoutBtn.addEventListener("click", async () => {
    await leaveMeeting();
    clearSession();
    showUnauthedUI();
  });

  els.createBtn.addEventListener("click", async () => {
    setHint(els.createResult, "");
    try {
      await createMeeting();
    } catch (err) {
      setHint(els.createResult, err?.message || "创建失败", "error");
    }
  });

  els.joinBtn.addEventListener("click", async () => {
    setHint(els.joinResult, "");
    const meetingId = (els.meetingIdInput.value || "").trim();
    if (!meetingId) {
      setHint(els.joinResult, "请输入会议 ID", "error");
      return;
    }
    try {
      await joinMeeting(meetingId);
    } catch (err) {
      setHint(els.joinResult, err?.message || "加入失败", "error");
    }
  });

  els.startMediaBtn.addEventListener("click", async () => {
    try {
      await startLocalMedia();
      if (state.meetingId && state.roomId) {
        await connectSfu().catch((err) => addChatLine("error", err?.message || "媒体连接失败", "error"));
        await applyLocalTracksToSfu();
      } else {
        addChatLine("system", "已开启本地媒体（加入会议后会自动连接 SFU）");
      }
    } catch (err) {
      addChatLine("error", err?.message || "打开媒体失败", "error");
    }
  });

  els.leaveBtn.addEventListener("click", async () => {
    await leaveMeeting();
  });

  els.muteBtn.addEventListener("click", () => toggleMute());
  els.videoBtn.addEventListener("click", () => toggleVideo());
  els.screenBtn.addEventListener("click", async () => {
    try {
      await toggleScreenShare();
    } catch (err) {
      addChatLine("error", err?.message || "共享屏幕失败", "error");
    }
  });

  els.chatForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    try {
      await sendChat(els.chatInput.value);
    } catch (err) {
      addChatLine("error", err?.message || "发送失败", "error");
    }
  });

  // AI panel
  els.aiClearBtn?.addEventListener("click", () => aiClearOutput());
  els.aiHealthBtn?.addEventListener("click", async () => {
    await aiHealthCheck();
  });
  els.aiInfoBtn?.addEventListener("click", async () => {
    await aiGetInfo();
  });

  els.aiLiveToggleBtn?.addEventListener("click", async () => {
    if (!els.aiLiveToggleBtn) return;
    els.aiLiveToggleBtn.disabled = true;
    try {
      if (state.aiLiveEnabled) {
        await stopAiLive();
      } else {
        await startAiLive();
      }
    } catch (err) {
      aiAppendOutput("Realtime AI Error", err?.message || String(err));
    } finally {
      els.aiLiveToggleBtn.disabled = false;
    }
  });

  els.aiLiveClearBtn?.addEventListener("click", () => clearAiLiveLog());

  els.aiLiveAsr?.addEventListener("change", () => {
    if (!els.aiLiveAsr || !els.aiLiveEmotion) return;
    if (!els.aiLiveAsr.checked) {
      els.aiLiveEmotion.checked = false;
      els.aiLiveEmotion.disabled = true;
    } else {
      els.aiLiveEmotion.disabled = false;
    }
  });

  els.aiEmotionExampleBtn?.addEventListener("click", () => {
    if (!els.aiEmotionText) return;
    els.aiEmotionText.value = "I am very happy today!";
    els.aiEmotionText.focus();
  });
  els.aiEmotionText?.addEventListener("keydown", async (e) => {
    if (e.key !== "Enter") return;
    e.preventDefault();
    els.aiEmotionBtn?.click();
  });
  els.aiEmotionBtn?.addEventListener("click", async () => {
    if (!els.aiEmotionBtn) return;
    els.aiEmotionBtn.disabled = true;
    try {
      await aiRunEmotion();
    } catch (err) {
      aiAppendOutput("Emotion Error", err?.message || String(err));
    } finally {
      els.aiEmotionBtn.disabled = false;
    }
  });
  els.aiAsrBtn?.addEventListener("click", async () => {
    if (!els.aiAsrBtn) return;
    els.aiAsrBtn.disabled = true;
    try {
      await aiRunAsr();
    } catch (err) {
      aiAppendOutput("ASR Error", err?.message || String(err));
    } finally {
      els.aiAsrBtn.disabled = false;
    }
  });
  els.aiSynthBtn?.addEventListener("click", async () => {
    if (!els.aiSynthBtn) return;
    els.aiSynthBtn.disabled = true;
    try {
      await aiRunSynthesis();
    } catch (err) {
      aiAppendOutput("Synthesis Error", err?.message || String(err));
    } finally {
      els.aiSynthBtn.disabled = false;
    }
  });

  // auto restore session
  if (state.token) {
    try {
      await refreshProfile();
      showAuthedUI();
      setAuthMsg("");
    } catch {
      clearSession();
      showUnauthedUI();
    }
  } else {
    showUnauthedUI();
  }

  updateCallButtons();
}

boot().catch((err) => {
  // eslint-disable-next-line no-console
  console.error(err);
});
