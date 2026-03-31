﻿//////////////////////////////////////////////
//bull 2025/3/31 批量控制
//////////////////////////////////////////////
//////////////////////////////////////////////
//bull 2020/6/10  云挂机H5版 控制模块
//////////////////////////////////////////////
var MCtrlConn = function(a) {
        this.cmsApi = "/ajax.php";
        this.ctr_socket = null;
        this.socket_addr = "";
        this.send_buffer = new ArrayBuffer(1024);
        this.init_socket = function(a) {
            try {
                null != this.ctr_socket && (this.ctr_socket.close(), this.ctr_socket = null), "" != a && (this.socket_addr = a), "" != this.socket_addr && (this.ctr_socket = new WebSocket(this.socket_addr)), this.ctr_socket.onopen = function(a) {
                    console.log("mctrl socket 连接成功...")
                }.bind(this), this.ctr_socket.onmessage = function(a) {}.bind(this), this.ctr_socket.onclose = function(a) {
                    this.ctr_socket = null;
                    console.log("mctrl socket 关闭连接...")
                }.bind(this), this.ctr_socket.onerror = function(a) {
                    console.log("mctrl socket 错误...")
                }.bind(this)
            } catch (c) {
                console.log("mctrl socket 有错误发生")
            }
        };
        this.start = function() {
            var b = this;
            $.ajax({
                type: "GET",
                url: this.cmsApi + "?token=" + a + "&c=mctrl&a=info",
                timeout: 5E3,
                datatype: "json",
                success: function(a) {
                    a = JSON.parse(a);
                    200 == a.code && "" != a.addr && b.init_socket(a.addr)
                },
                error: function(a) {
                    console.log("get_addr error")
                }
            })
        };
        this.sendData = function(a) {
            if (null != this.ctr_socket) {
                for (var c = new DataView(this.send_buffer, 0, a.byteLength - 8), b = 8; b < a.byteLength; b++) c.setUint8(b - 8, a.getUint8(b));
                this.ctr_socket.send(c)
            }
        };
        this.end = function() {
            null != this.ctr_socket && (this.ctr_socket.close(), this.ctr_socket = null)
        }
    },
    VideoGame = function(a) {
        this._config = a || {};
        void 0 == this._config.pcConfig && (this._config.pcConfig = {
            iceServers: [{
                urls: "stun:stun.l.google.com:19302"
            }]
        });
        a = document.getElementById("container");
        this.video = document.getElementById(this._config.id);
        if (void 0 == this.video) {
            // 不注入 h5lgair_video_css，由 play.html 自身 CSS 控制 video 和 #container 样式
            b = document.createElement("video");
            b.setAttribute("id", this._config.id);
            b.style = "opacity: 1;";
            b.autoplay = !0;
            b.muted = !0;
            b.setAttribute("playsinline", "true");
            b.setAttribute("webkit-playsinline", "true");
            b.setAttribute("x5-video-player-type", "h5");
            b.setAttribute("object-fit", "fill");
            b["class"] = "video";
            a.appendChild(b);
            b.style.zindex = -1;
            this.video = b
        }
        this.audio = document.createElement("audio");
        this.audio.setAttribute("id", "gaudio");
        this.audio.autoplay = !0;
        a.appendChild(this.audio);
        this.sendChannel = this.ping_timer = this.ws = this.pconnection = null;
        this.playing = !1;
        this.onDatarecv = this.onDataChannelReady = this.onVideoStart = this.onDisconnected = this.onClose = this.onConnected = this.stat_timer = null;
        this.V_NULL = 0;
        this.V_CONNECT_OK = 1;
        this.V_CONNECT_FAIL = 2;
        this.V_CONNECT_DIS = 3;
        this.V_VIDEO_READY = 4;
        this.status = this.V_NULL;
        this.delay_high_count = 0;
        this.getSrcObjectString = function() {
            var a = document.createElement("video");
            return "srcObject" in a ? "srcObject" : "mozSrcObject" in a ? "mozSrcObject" : "webkitSrcObject" in a ? "webkitSrcObject" : "srcObject"
        }.bind(this);
        this.isVideoPlaying = function(a) {
            return !!(0 < this.video.currentTime && !this.video.paused && !this.video.ended && 2 < this.video.readyState)
        }.bind(this);
        this.ping = function() {
            this.sendMessage({
                id: "heart",
                data: "1"
            })
        }.bind(this);
        this.log = function(a) {}.bind(this);
        this.start = function() {
            try {
                this.pconnection = new RTCPeerConnection(this._config.pcConfig), this.pconnection.onicecandidate = this.handleIceCandidate, this.pconnection.ontrack = this.handleTrack, this.pconnection.ondatachannel = this.ondatachannelHandle, this.pconnection.onaddstream = this.onaddstreamHandle, this.pconnection.onconnectionstatechange = this.onconnectionstatechangeHandle
            } catch (c) {
                this.log("Failed to create PeerConnection, exception: " + c.message);
                window.h5_lgair.app_call("Unsupport h264");
                window.h5_lgair.fun_log("start Unsupport h264 exception:" + c.message, window.h5_lgair.LOG_DEBUG);
                return
            }
            this.ws = new WebSocket(this._config.url);
            this.ws.onopen = function(a) {
                this.log("open")
            }.bind(this);
            this.ws.onclose = function(a) {
                clearInterval(this.ping_timer);
                window.h5_lgair.fun_log("websocket close 连接关闭...", window.h5_lgair.LOG_ERROR);
                window.h5_lgair.on_play_end({
                    text: "连接关闭..."
                }, 0)
            }.bind(this);
            this.ws.onmessage = function(a) {
                this.onMessage(a.data)
            }.bind(this);
            this.ws.onerror = function(a) {
                this.log("onError() error: " + a);
                window.h5_lgair.fun_log("websocket error 系统繁忙...", window.h5_lgair.LOG_ERROR);
                window.h5_lgair.on_play_end({
                    text: "系统繁忙..."
                }, 1)
            }.bind(this)
        }.bind(this);
        this.end = function() {
            null != this.ws && (this.ws.close(), this.ws = null);
            null != this.pconnection && (this.pconnection.close(), this.pconnection = null);
            null != this.stat_timer && (clearInterval(this.stat_timer), LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("stat_timer cleared", this.stat_timer), this.stat_timer = null);
            null != this.ping_timer && (clearInterval(this.ping_timer), LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("ping_timer cleared", this.ping_timer), this.ping_timer = null);
            try {
                var a = this.getSrcObjectString();
                if (this.video) {
                    var b = this.video[a];
                    b && b.getTracks && b.getTracks().forEach(function(a) {
                        a.stop && a.stop()
                    });
                    this.video[a] = null;
                    this.video.src && 0 === this.video.src.indexOf("blob:") && URL.revokeObjectURL(this.video.src);
                    this.video.src = ""
                }
                if (this.audio) {
                    var c = this.audio[a];
                    c && c.getTracks && c.getTracks().forEach(function(a) {
                        a.stop && a.stop()
                    });
                    this.audio[a] = null;
                    this.audio.src = ""
                }
            } catch (d) {
                LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("media cleanup error", d)
            }
        };
        this.play = function() {
            this.status != this.V_VIDEO_READY || this.isVideoPlaying() || (this.video.style.display = "block", this.video.play(), this.audio.play())
        }.bind(this);
        this.pause = function() {
            this.status == this.V_VIDEO_READY && (this.video.pause(), this.audio.pause())
        }.bind(this);
        this.onMessage = function(a) {
            try {
                var c = JSON.parse(a),
                    b = atob(c.data);
                if ("offer" == c.id) window.h5_lgair.fun_log("Recv:offer", window.h5_lgair.LOG_DEBUG), this.ping_timer = setInterval(this.ping, 1E3), LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("ping_timer set", this.ping_timer), this.pconnection.setRemoteDescription(new RTCSessionDescription(JSON.parse(b))), this.pconnection.createAnswer(this.setLocalAndSendMessage, this.handleCreateOfferError);
                else if ("icecandiate" == c.id) {
                    window.h5_lgair.fun_log("onMessagexxxx, candidate:" + b, window.h5_lgair.LOG_DATA);
                    void 0 != this._config.rtc_addr && "" != this._config.rtc_addr && (b = b.replace(RegExp("(.+)(udp|tcp)( \\d+ )(.+)( typ.+)", "gmi"), "$1$2$3" + this._config.rtc_addr + "$5"), window.h5_lgair.fun_log("onMessagexxxx, candidate-Trans:" + b, window.h5_lgair.LOG_DATA));
                    var h = new RTCIceCandidate(JSON.parse(b));
                    this.pconnection.addIceCandidate(h)
                }
            } catch (k) {
                if (this.log("onMessage, exception: " + k.message), window.h5_lgair.app_call("Unsupport h264"), window.h5_lgair.fun_log("Err Unsupport h264 exception:" + k.message, window.h5_lgair.LOG_DEBUG), this.onDisconnected) this.onDisconnected()
            }
        }.bind(this);
        this.sendMessage = function(a) {
            this.ws.send(JSON.stringify(a))
        }.bind(this);
        this.sendData = function(a) {
            this.sendChannel.send(a)
        }.bind(this);
        this.setLocalAndSendMessage = function(a) {
            this.pconnection.setLocalDescription(a);
            window.h5_lgair.fun_log("SendAnswer:" + JSON.stringify(a), window.h5_lgair.LOG_DATA);
            this.sendMessage({
                id: "answer",
                data: btoa(JSON.stringify(a))
            });
            if (-1 == a.sdp.indexOf("H264") && (window.h5_lgair.app_call("Unsupport h264"), window.h5_lgair.fun_log("Unsupport h264:" + JSON.stringify(a), window.h5_lgair.LOG_DEBUG), this.onDisconnected)) this.onDisconnected()
        }.bind(this);
        this.handleCreateOfferError = function(a) {
            this.log("createOffer() error: " + a);
            window.h5_lgair.app_call("Unsupport h264");
            window.h5_lgair.fun_log("CreateOfferError Unsupport h264 exception" + a, window.h5_lgair.LOG_DEBUG)
        }.bind(this);
        this.handleIceCandidate = function(a) {
            this.log("icecandidate event: " + a)
        }.bind(this);
        this.ondatachannelHandle = function(a) {
            this.log("Data channel is created!");
            a.channel.onopen = function() {
                this.sendChannel = a.channel;
                this.log("Data channel is open and ready to be used.");
                if (this.onDataChannelReady) this.onDataChannelReady(a)
            }.bind(this);
            a.channel.onmessage = function(a) {
                this.log("Recv Data:" + a);
                if (this.onDatarecv) this.onDatarecv(a)
            }.bind(this)
        }.bind(this);
        this.onaddstreamHandle = function(a) {
            this.log("onaddstream event detected!");
            try {
                var b = a && a.stream ? a.stream : null;
                var c = b && b.getVideoTracks ? b.getVideoTracks().length : 0;
                var d = b && b.getAudioTracks ? b.getAudioTracks().length : 0;
                LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("onaddstream", {videoTracks: c, audioTracks: d})
            } catch (e) {}
            this.status = this.V_VIDEO_READY;
            if (this.onVideoStart) this.onVideoStart();
            this.stat_timer = setInterval(function() {
                this.getStats(this.getResult);
                if (!this._statLogNextAt) this._statLogNextAt = Date.now() + 6E4;
                if (Date.now() >= this._statLogNextAt) {
                    this._statLogNextAt = Date.now() + 6E4;
                    this.getStats(this._logStats)
                }
            }.bind(this), 1E3), LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("stat_timer set", this.stat_timer)
        }.bind(this);
        this.onconnectionstatechangeHandle = function(a) {
            switch (this.pconnection.connectionState) {
                case "connected":
                    this.status != this.V_VIDEO_READY && (this.status = this.V_CONNECT_OK);
                    this.log("The connection has become fully connected");
                    if (this.onConnected) this.onConnected();
                    break;
                case "failed":
                    this.status = this.V_CONNECT_FAIL;
                    this.log("One or more transports has terminated unexpectedly or in an error");
                    if (this.onDisconnected) this.onDisconnected();
                    break;
                case "disconnected":
                case "closed":
                    if (this.status = this.V_CONNECT_DIS, this.log("The connection has been closed"), this.onClose) this.onClose()
            }
        }.bind(this);
        this.handleTrack = function(a) {
            this.log("Track");
            try {
                var b = [];
                var c = 0;
                for (var d = 0; d < a.streams.length; d++) {
                    var h = a.streams[d];
                    b.push({
                        v: h.getVideoTracks ? h.getVideoTracks().length : 0,
                        a: h.getAudioTracks ? h.getAudioTracks().length : 0
                    });
                    c++
                }
                LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("handleTrack", {streams: c, tracks: b})
            } catch (k) {}
            for (var c = this.getSrcObjectString(), b = 0; b < a.streams.length; b++) {
                if (a.streams[b].getVideoTracks().length) {
                    var h = new MediaStream(a.streams[b].getVideoTracks());
                    if (this.video && this.video[c] !== h) try {
                        this.video[c] = h
                    } catch (k) {
                        this.video.src = window.URL.createObjectURL(h)
                    }
                }
                a.streams[b].getAudioTracks().length && this.audio && this.audio[c] !== a.streams[b] && (this.audio[c] = a.streams[b])
            }
        }.bind(this);
        VideoGame.prototype.requestFullScreen = function(a) {
            navigator.keyboard && navigator.keyboard.lock();
            a = a ? a : document.documentElement;
            return a.requestFullscreen ? a.requestFullscreen() : a.webkitRequestFullScreen ? a.webkitRequestFullScreen() : a.mozRequestFullScreen ? a.mozRequestFullScreen() : a.msRequestFullscreen ? a.msRequestFullscreen() : Error("不支持全屏")
        };
        VideoGame.prototype.exitRequestFullscreen = function() {
            this.isRequestFullscreen() && (document.exitFullscreen ? document.exitFullscreen() : document.mozCancelFullScreen ? document.mozCancelFullScreen() : document.webkitExitFullscreen && document.webkitExitFullscreen())
        };
        VideoGame.prototype.isRequestFullscreen = function() {
            return !!(document.fullscreenElement || document.webkitFullscreenElement || document.mozFullScreenElement)
        };
        this.getResult = function(a) {
            for (var c = 0; c < a.length; ++c) {
                var b = a[c];
                if ("ssrc" == b.type && "video" == b.mediaType) {
                    var h = b.packetsLost / (b.packetsLost + b.packetsReceived) * 100,
                        h = h.toFixed(2),
                        b = parseInt(b.googCurrentDelayMs);
                    200 < b || 3 < h ? this.delay_high_count++ : (150 > b || 1 > h) && 0 > --this.delay_high_count && (this.delay_high_count = 0);
                    LGAIR_UI.set_tips("延时:" + b + "ms  丢包率:" + h + "% Count:" + this.delay_high_count);
                    8 < this.delay_high_count && (this.delay_high_count = 0, LGAIR_UI.add_tips_msg("您的网络太差，建议您切换网络再试"))
                }
            }
        }.bind(this);
        this.getStats = function(a) {
            if (!this.pconnection || !this.pconnection.getStats) return;
            this.pconnection.getStats().then(function(c) {
                var b = [];
                c.forEach(function(a) {
                    b.push(a)
                });
                a(b)
            }).catch(function(b) {
                LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("getStats error", b)
            })
        }.bind(this)
        ;
        this._logStats = function(a) {
            try {
                var b = {
                    framesDecoded: null,
                    framesDropped: null,
                    bytesReceived: null,
                    jitterBufferDelay: null,
                    jitterBufferEmittedCount: null,
                    totalDecodeTime: null
                };
                for (var c = 0; c < a.length; c++) {
                    var d = a[c];
                    if ("inbound-rtp" === d.type && "video" === d.kind) {
                        b.framesDecoded = d.framesDecoded;
                        b.framesDropped = d.framesDropped;
                        b.bytesReceived = d.bytesReceived;
                        b.jitterBufferDelay = d.jitterBufferDelay;
                        b.jitterBufferEmittedCount = d.jitterBufferEmittedCount;
                        b.totalDecodeTime = d.totalDecodeTime
                    }
                    if ("ssrc" === d.type && "video" === d.mediaType) {
                        b.framesDecoded = null == b.framesDecoded ? d.framesDecoded : b.framesDecoded;
                        b.framesDropped = null == b.framesDropped ? d.framesDropped : b.framesDropped;
                        b.bytesReceived = null == b.bytesReceived ? d.bytesReceived : b.bytesReceived
                    }
                }
                var e = this.video || null;
                var f = {
                    readyState: e ? e.readyState : null,
                    width: e ? e.videoWidth : null,
                    height: e ? e.videoHeight : null,
                    currentTime: e ? Math.round(100 * e.currentTime) / 100 : null,
                    bufferedEnd: null
                };
                if (e && e.buffered && e.buffered.length) {
                    f.bufferedEnd = Math.round(100 * e.buffered.end(e.buffered.length - 1)) / 100
                }
                LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("rtc-stats", {rtc: b, video: f})
            } catch (g) {
                LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("rtc-stats error", g)
            }
        }.bind(this)
    },
    LGAIR_UI = {
        logo_url: "public/images/loading.png",
        icon_url: "public/images/108.png",
        ad_url: "public/images/loading.png",
        ad_goto_url: "#",
        logo_showTime: 1,
        logo_startTime: 0,
        ad_showTime: 1E6,
        ad_minTime: 2,
        video_state: 0,
        adTimer: null,
        logoTimer: null,
        auto_timer_hander: null,
        msg_timer_hander: null,
        reconnectTimer: null,
        reconnectStart: 0,
        debug_clickTime: 0,
        debug_clickCount: 0,
        tipIndex: 0,
        TIP_STATE: 1,
        TIP_CONSOLE: 2
    };
LGAIR_UI.tipMode = LGAIR_UI.TIP_STATE;
LGAIR_UI.screen_locked = !1;
LGAIR_UI._debugCounts = LGAIR_UI._debugCounts || {};
LGAIR_UI._leakLog = function(a, b) {
    try {
        console.log("[lgair][leak]", a, void 0 !== b ? b : "");
    } catch (c) {}
};
LGAIR_UI.create_style = function() {
    if (!document.styleSheets.h5lgairad_css) {
        var a = document.createElement("style"),
            b = ".splash {\t\t\t\tposition:absolute;\t\t\t\ttop:0;\t\t\t\tleft:0;\t\t\t\tz-index:200001;\t\t\t\theight:100vh;\t\t\t\twidth:100vw;\t\t\t\tdisplay:none;\t\t\t}\t\t@media screen and (orientation: landscape) {   \t\t\t/*横屏 css*/ \t\t\t.splash {\t\t\t\ttop:100%;\t\t\t\tleft:0;\t\t\t\theight:100vw;\t\t\t\twidth:100vh;\t\t\t\ttransform:rotate(-90deg);\t\t\t\ttransform-origin:0% 0%\t\t\t}\t\t}  \t\t.pageLogo {\t\t\tposition:absolute;\t\t\ttop:0;\t\t\tleft:0;\t\t\theight:100%;\t\t\twidth:100%;\t\t\tbackground:#dfdfdf;\t\t\tz-index:200001;\t\t} \t\t.pageLogo img {\t\t\tposition:relative;\t\t\ttop:0;\t\t\tleft:0;\t\t\twidth:100%;\t\t\theight:100%;\t\t\tz-index:100;\t\t} \t\t.pageAd {\t\t\tposition:absolute;\t\t\ttop:0;\t\t\tleft:0;\t\t\theight:100%;\t\t\twidth:100%;\t\t\tbackground:#000;\t\t\tz-index:200000;\t\t} \t\t.adContainer {\t\t\tposition:relative;\t\t\ttop:0;\t\t\tleft:0;\t\t\twidth:100%;\t\t\theight:100%;\t\t\tz-index:100;\t\t\tbackground:#000;\t\t\tfont-size: 0;\t\t} \t\t#lgimgad {\t\t\tposition:relative;\t\t\ttop:0;\t\t\tleft:0;\t\t\twidth:100%;\t\t\theight:100%;\t\t\tz-index:100;\t\t} \t\t.pageAd .adTimer {\t\t\tposition:absolute;\t\t\ttop:3vh;\t\t\tleft:80vw;\t\t\tbackground:#000;\t\t\tcolor:#fff;\t\t\tz-index:101;\t\t} \t\t.pageAd .adText {\t\t\tposition:absolute;\t\t\tbottom:10%;\t\t\tcolor:#111;\t\t\tz-index:101;\t\t\twidth:100%;\t\t\ttext-align: center;\t\t}";
        a.type = "text/css";
        a.id = "h5lgairad_css";
        a.styleSheet ? a.styleSheet.cssText = b : a.innerHTML = b;
        document.getElementsByTagName("head")[0].appendChild(a)
    }
    document.styleSheets.h5lgair_guidui_css || (a = document.createElement("style"), b = "\t\t#uicontainer {\t\t\tz-index:100;\t\t\tbackground: transparent;\t\t\ttext-align: center;\t\t\tmargin-top: 0%;\t\t\tmargin:0;\t\t\twidth:calc(100vw - 42px);\t\t\theight:calc(100vh - 42px);\t\t\tpointer-events: none;\t\t}\t\t.modal-container {\t\t\tposition: absolute;\t\t\tz-index: 5;\t\t\tbox-shadow: none;\t\t\ttext-align: center;\t\t\tbackground: transparent;\t\t}\t\t.modal-container .modal-content {\t\t\tposition: relative;\t\t\tmax-width: 800px;\t\t\tbox-shadow: none;\t\t\tmargin: auto;\t\t\toutline: none;\t\t\tbackground: transparent;\t\t}\t\t.modal-container .modal-overlay {\t\t\tposition: fixed;\t\t\tdisplay: flex;\t\t\talign-items: flex-start;\t\t\ttop: 0px;\t\t\tleft: 0px;\t\t\tright: 0px;\t\t\tbottom: 0px;\t\t\toverflow-y: auto;\t\t\toverflow-x: hidden;\t\t\tz-index: 1000;\t\t\tbox-shadow: none;\t\t\tbackground: rgba(0, 0, 0, 0.75);\t\t}\t\t.modal-container .modal-overlay-none {\t\t\tposition: fixed;\t\t\tdisplay: flex;\t\t\talign-items: flex-start;\t\t\ttop: 0px;\t\t\tleft: 0px;\t\t\tright: 0px;\t\t\tbottom: 0px;\t\t\toverflow-y: auto;\t\t\toverflow-x: hidden;\t\t\tz-index: 1000;\t\t\tbox-shadow: none;\t\t\tpointer-events: none;\t\t}\t\t.tipButton {\t\t\tbackground-color: rgb(255, 255, 255);\t\t\tcursor: pointer;\t\t\tuser-select: none;\t\t\tline-height: 60px;\t\t\theight: 60px;\t\t\tfont-size: 16px;\t\t\tborder-width: initial;\t\t\tborder-style: none;\t\t\tborder-color: initial;\t\t\tborder-image: initial;\t\t\toutline: none;\t\t\tpadding: 0px 30px;\t\t\tborder-radius: 35px;\t\t}\t\t.tipMsg {\t\t\tbackground-color: rgba(255, 255, 255,0.75);\t\t\tcursor: pointer;\t\t\tuser-select: none;\t\t\tline-height: 36px;\t\t\theight: 36px;\t\t\tfont-size: 16px;\t\t\tborder-width: initial;\t\t\tborder-style: none;\t\t\tborder-color: initial;\t\t\tborder-image: initial;\t\t\toutline: none;\t\t\tpadding: 0px 30px;\t\t\tborder-radius: 35px;\t\t}\t\tbutton {\t\t\t-webkit-appearance: button;\t\t\t-webkit-writing-mode: horizontal-tb !important;\t\t\ttext-rendering: auto;\t\t\tcolor: buttontext;\t\t\tletter-spacing: normal;\t\t\tword-spacing: normal;\t\t\ttext-transform: none;\t\t\ttext-indent: 0px;\t\t\ttext-shadow: none;\t\t\tdisplay: inline-block;\t\t\ttext-align: center;\t\t\talign-items: flex-start;\t\t\tcursor: default;\t\t\tbackground-color: buttonface;\t\t\tbox-sizing: border-box;\t\t\tmargin: 0em;\t\t\tfont: 400 13.3333px Arial;\t\t\tpadding: 1px 6px;\t\t\tborder-width: 2px;\t\t\tborder-style: outset;\t\t\tborder-color: buttonface;\t\t\tborder-image: initial;\t\t}", a.type = "text/css", a.id = "h5lgair_guidui_css", a.styleSheet ? a.styleSheet.cssText = b : a.innerHTML = b, document.getElementsByTagName("head")[0].appendChild(a));
    document.styleSheets.h5lgairio_css || (a = document.createElement("style"), b = ".io_div {position: absolute;top: 0;left: 0;margin:5px;width: 90%;height: 32px;z-index: 200011;display:none;} input{width:80%;line-height:32px;border:1px solid red;}.paste_btn{position: absolute;z-index: 200010;display:none;} .btn{margin-left:10px;border-radius: 7px;background: #8AC007;padding:7px;height: 20px;}", a.type = "text/css", a.id = "h5lgairio_css", a.styleSheet ? a.styleSheet.cssText = b : a.innerHTML = b, document.getElementsByTagName("head")[0].appendChild(a));
    document.styleSheets.h5lgair_assistive_css || (a = document.createElement("style"), b = '.assistDiv {\t\t\t\tposition:absolute;\t\t\t\ttop:0;\t\t\t\tleft:0;\t\t\t\tz-index:50;\t\t\t\twidth:100vw;\t\t\t}\t\t@media screen and (orientation: landscape) {   \t\t\t/*横屏 css*/ \t\t\t.assistDiv {\t\t\t\ttop:100%;\t\t\t\tleft:0;\t\t\t\twidth:100vh;\t\t\t\ttransform:rotate(-90deg);\t\t\t\ttransform-origin:0% 0%;\t\t\t}\t\t}  #drager {   position: fixed;   border-radius: 50%;   width: 32px;   height: 32px;   background-color: rgba(0, 0, 0, 0.7);   z-index: 10000;   cursor: pointer;   top: 0px;   left: 0px;   border-radius: 30%;   padding: 6px; }  #drager>div {   border-radius: 50%;   width: 100%;   height: 100%;   background-color: rgba(255, 255, 255, 0.7);   transition: all 0.2s;  -webkit-transition: all 0.2s;  -moz-transition: all 0.2s;  -o-transition: all 0.2s; } #drager:hover>div{   background-color: rgba(255, 255, 255, 0.9); } \t.tip {\t\tposition:absolute;\t\tleft:0px;\t\ttop:0px;\t\twidth:100%;\t\theight:20px;\t\tz-index:1000000;\t\tbackground-color:rgba(0, 0, 0, 0.4);\t\tcolor:#fff;\t\tfont-size: 13px;\t\tdisplay:none;\t\tpointer-events: none;\t}\t.tip2 {\t\tposition: fixed;\t\tleft: 50%;\t\ttransform: translate(-50%, 0%);\t\tz-index: 10;\t\ttext-align: center;\t\twidth: 180px;\t\tbackground-color: rgba(0, 0, 0, 0.8);\t\tcolor: rgb(255, 255, 255);\t\tdisplay: flex;\t\tflex-direction: row;\t\tjustify-content: space-around;\t\tfont-size: 13px;\t\tpadding: 2px 10px;\t\tborder-radius: 5px;\t\tpointer-events: none;\t}\t.console {\t\tposition:absolute;\t\tleft:0px;\t\ttop:0px;\t\twidth:100%;\t\theight:100px;\t\tz-index:1000000;\t\tbackground-color:rgba(0, 0, 0, 0.4);\t\tcolor:#fff;\t\tfont-size: 13px;\t\tdisplay:none;\t\toverflow:auto;\t}\t.console p {\t\tmargin: 0px;\t\tpadding-top: 0px;\t\tpadding-right: 0px;\t\tpadding-left: 0px;\t}', a.type = "text/css", a.id = "h5lgair_assistive_css", a.styleSheet ? a.styleSheet.cssText = b : a.innerHTML = b, document.getElementsByTagName("head")[0].appendChild(a))
};
LGAIR_UI.create_logo = function(a) {
    var b = LGAIR_UI.logo_showTime,
        c = LGAIR_UI.logo_url,
        f = LGAIR_UI.icon_url;
    a = a || {};
    void 0 != a.imgurl && "" != a.imgurl && (c = a.imgurl);
    void 0 != a.icon_url && "" != a.icon_url && (f = a.icon_url);
    void 0 != a.stime && (b = a.stime);
    var d = document.getElementById("lgairlogo");
    if (null == d) {
        d = document.createElement("div");
        d.setAttribute("id", "lgairlogo");
        d.setAttribute("class", "pageLogo");
        d.innerHTML = '<img src="' + c + '" id="lgimglogo"></img>';
        document.getElementById("splash").appendChild(d);
        d = document.getElementById("lgairlogo");
        d.style.display = "block";
        $("#splash").show();
        var c = document.getElementById("splash"),
            d = "50%",
            h = "25%";
        void 0 != a.icon_x && "" != a.icon_x && (d = a.icon_x, isNaN(Number(d)) || (d += "%"));
        void 0 != a.icon_y && "" != a.icon_y && (h = a.icon_y, isNaN(Number(h)) || (h += "%"));
        void 0 != a.icon_size && "" != a.icon_size && (h = a.icon_size);
        LGAIR_UI.create_progress(d, h, 80, f, "努力加载中...", c);
        LGAIR_UI.progress_auto(.8, 100)
    }
    null == LGAIR_UI.logoTimer && (LGAIR_UI.logo_startTime = Date.parse(new Date), LGAIR_UI.logoTimer = setTimeout(LGAIR_UI.close_logo, 1E3 * b), LGAIR_UI._debugCounts.logoTimerSet = (LGAIR_UI._debugCounts.logoTimerSet || 0) + 1, LGAIR_UI._leakLog("logoTimer set", LGAIR_UI.logoTimer))
};
LGAIR_UI.close_logo = function(a) {
    if (a || !(Date.parse(new Date) < LGAIR_UI.logo_startTime + LGAIR_UI.logo_showTime))
        if (1 == LGAIR_UI.video_state || a) LGAIR_UI.logoTimer && (clearInterval(LGAIR_UI.logoTimer), LGAIR_UI._leakLog("logoTimer cleared", LGAIR_UI.logoTimer), LGAIR_UI.logoTimer = null), LGAIR_UI.auto_timer_hander && (clearInterval(LGAIR_UI.auto_timer_hander), LGAIR_UI._leakLog("auto_timer_hander cleared", LGAIR_UI.auto_timer_hander), LGAIR_UI.auto_timer_hander = null), $("#lgairlogo").hide(), $("#splash").hide(), $("#progress_div").hide(), LGAIR_UI.progress_value(100)
};
LGAIR_UI.create_ad = function(a) {
    var b = LGAIR_UI.ad_showTime,
        c = LGAIR_UI.ad_minTime,
        f = LGAIR_UI.ad_url,
        d = LGAIR_UI.ad_goto_url,
        h = 0,
        k = "";
    LGAIR_UI.close_logo(!0);
    a = a || {};
    void 0 != a.url && (d = a.url);
    void 0 != a.stime && (b = a.stime);
    void 0 != a.mtime && (c = a.mtime);
    void 0 != a.text && (k = a.text);
    void 0 != a.imgurl && "" != a.imgurl && (f = a.imgurl);
    void 0 != a.forcejump && a.forcejump ? window.location.href = d : (a = document.getElementById("lgairad"), null == a && (a = document.createElement("div"), a.setAttribute("id", "lgairad"), a.setAttribute("class", "pageAd"), a.innerHTML = '<div class="adContainer"><a href="' + d + '"><img id="lgimgad" src="' + f + '" alt=""></img></a></div><div class="adTimer" id="lgadTimer" showTime="' + b + '"></div><div class="adText" id="adText">' + k + "</div>", document.getElementById("splash").appendChild(a), a = document.getElementById("lgairad"), a.style.display = "block", $("#splash").show()), null == LGAIR_UI.adTimer && (h < c ? $("#lgadTimer").css({
        background: "gray"
    }) : $("#lgadTimer").css({
        background: "blue"
    }), LGAIR_UI.adTimer = setInterval(function() {
        b--;
        h++;
        0 <= b ? ($("#lgadTimer").html(""), h < c ? $("#lgadTimer").css({
            background: "gray"
        }) : $("#lgadTimer").css({
            background: "blue"
        })) : LGAIR_UI.close_ad()
    }, 1E3), LGAIR_UI._debugCounts.adTimerSet = (LGAIR_UI._debugCounts.adTimerSet || 0) + 1, LGAIR_UI._leakLog("adTimer set", LGAIR_UI.adTimer)), $("#lgairad").on("click", function() {
        h >= c && LGAIR_UI.close_ad()
    }))
};
LGAIR_UI.close_ad = function() {
    if (1 == LGAIR_UI.video_state && LGAIR_UI.adTimer) {
        clearInterval(LGAIR_UI.adTimer);
        LGAIR_UI._leakLog("adTimer cleared", LGAIR_UI.adTimer);
        LGAIR_UI.adTimer = null;
        var a = document.getElementById("lgairad");
        a && a.parentNode && a.parentNode.removeChild(a);
        var b = document.getElementById("splash");
        b && (b.style.display = "none")
    }
};
LGAIR_UI.stopAllTimers = function() {
    this.stopProgress();
    if (this._leakSamplerId) {
        clearInterval(this._leakSamplerId);
        this._leakSamplerId = null;
        this._leakLog("leak sampler cleared")
    }
    if (this.adTimer) {
        clearInterval(this.adTimer);
        this._leakLog("adTimer cleared(stopAllTimers)", this.adTimer);
        this.adTimer = null
    }
    if (this.auto_timer_hander) {
        clearInterval(this.auto_timer_hander);
        this._leakLog("auto_timer_hander cleared(stopAllTimers)", this.auto_timer_hander);
        this.auto_timer_hander = null
    }
    if (this.logoTimer) {
        clearTimeout(this.logoTimer);
        this._leakLog("logoTimer cleared(stopAllTimers)", this.logoTimer);
        this.logoTimer = null
    }
    var a = document.getElementById("lgairad");
    a && a.parentNode && a.parentNode.removeChild(a);
    var b = document.getElementById("splash");
    b && (b.style.display = "none")
};
LGAIR_UI.video_start = function(a) {
    LGAIR_UI.video_state = 1;
    LGAIR_UI.close_logo();
    LGAIR_UI.close_ad();
    if (!LGAIR_UI._leakSamplerId) {
        LGAIR_UI._leakSamplerId = setInterval(function() {
            try {
                var a = performance && performance.memory ? Math.round(performance.memory.usedJSHeapSize / 1048576) + "MB" : "n/a";
                var b = document.getElementsByTagName("*").length;
                var c = window.h5_lgair && window.h5_lgair.g_player ? window.h5_lgair.g_player : null;
                var d = 0, h = 0, k = "n/a";
                if (c && c.getSrcObjectString) {
                    k = c.getSrcObjectString();
                    if (c.video && c.video[k] && c.video[k].getTracks) d = c.video[k].getTracks().length;
                    if (c.audio && c.audio[k] && c.audio[k].getTracks) h = c.audio[k].getTracks().length
                }
                LGAIR_UI._leakLog("sample", {heap: a, nodes: b, videoTracks: d, audioTracks: h})
            } catch (c) {}
        }, 6E4);
        LGAIR_UI._leakLog("leak sampler set", LGAIR_UI._leakSamplerId)
    }
    LGAIR_UI.create_assistive_touch(a);
    LGAIR_UI.create_input();
    LGAIR_UI.create_paste_btn()
};
LGAIR_UI.video_end = function(a) {
    0 < LGAIR_UI.video_state && (LGAIR_UI.exitRequestFullscreen(), LGAIR_UI.video_state = 0, LGAIR_UI.create_ad(a), LGAIR_UI.hide_input())
};
LGAIR_UI.create_input = function() {
    LGAIR_UI._debugCounts.create_input = (LGAIR_UI._debugCounts.create_input || 0) + 1;
    LGAIR_UI._leakLog("create_input", LGAIR_UI._debugCounts.create_input);
    document.body.insertAdjacentHTML("afterbegin", '<div class="io_div" id="input_box"><input id="input_text" type="text" name="df" placeholder="请在这里输入"><span class="btn" id="btn_go">→</span></div>');
    $("#input_box").show();
    $("#input_box").css({
        top: -1E3,
        left: -1E3
    });
    $("#input_text").focus();
    $("#input_text").on("blur", function() {
        $("#input_text").focus()
    });
    $("#input_text").on("compositionstart", function(a) {
        a.target.isNeedPrevent = !0
    });
    $("#input_text").on("compositionend", function(a) {
        a.target.isNeedPrevent = !1;
        a = a.target.value.trim();
        0 < a.length && window.h5_lgair.send_input_txt(a);
        $("#input_text").val("")
    });
    $("#input_text").on("input", function(a) {
        a.target.isNeedPrevent || (a = a.target.value.trim(), 0 < a.length && window.h5_lgair.send_input_txt(a), $("#input_text").val(""))
    });
    $("#btn_go").click(function() {
        window.h5_lgair.on_input_end(2)
    })
};
LGAIR_UI.show_input = function(a) {};
LGAIR_UI.hide_input = function(a) {};
LGAIR_UI.get_input_val = function(a) {
    return $("#input_text").val()
};
LGAIR_UI.create_paste_btn = function() {
    LGAIR_UI._debugCounts.create_paste_btn = (LGAIR_UI._debugCounts.create_paste_btn || 0) + 1;
    LGAIR_UI._leakLog("create_paste_btn", LGAIR_UI._debugCounts.create_paste_btn);
    document.body.insertAdjacentHTML("afterbegin", '<spain class="btn paste_btn" id="paste_btn">粘&nbsp;贴</spain>');
    $("#paste_btn").click(function() {
        $("#paste_btn").offset();
        $("#paste_btn").hide();
        navigator.clipboard.readText().then(function(a) {
            window.h5_lgair.send_input_txt(a)
        })
    });
    $(document).keydown(function(a) {
        a.ctrlKey && 86 == a.keyCode || (13 == a.which ? window.h5_lgair.sendCmdEvent("VK_RETURN") : 8 == a.which ? window.h5_lgair.sendCmdEvent("VK_BACK") : 32 == a.which && window.h5_lgair.send_input_txt(String.fromCharCode(a.keyCode)))
    })
};
LGAIR_UI.show_paste_btn = function(a, b) {};
LGAIR_UI.hide_paste_btn = function() {};
LGAIR_UI.set_tips = function(a, b) {
    void 0 == b && (b = LGAIR_UI.TIP_STATE);
    b & LGAIR_UI.TIP_STATE && $("#tip").text(a);
    if (b & LGAIR_UI.TIP_CONSOLE) {
        var c = $("#console");
        c.prepend("<p>[" + this.tipIndex + "]" + a + "</p>");
        this.tipIndex++;
        var d = c.children("p");
        200 < d.length && d.slice(200).remove()
    }
};
LGAIR_UI.create_assistive_touch = function(a) {
    LGAIR_UI._debugCounts.create_assistive_touch = (LGAIR_UI._debugCounts.create_assistive_touch || 0) + 1;
    LGAIR_UI._leakLog("create_assistive_touch", LGAIR_UI._debugCounts.create_assistive_touch);
    function b(a) {
        k = document.documentElement.clientWidth;
        m = document.documentElement.clientHeight;
        a || (a = window.event);
        d = a.clientX - parseInt(e.style.left);
        h = a.clientY - parseInt(e.style.top);
        document.onmousemove = c
    }

    function c(a) {
        null == a && (a = window.event);
        0 >= a.clientY - h ? e.style.top = "0px" : a.clientY - h > m - parseInt(e.clientHeight) ? e.style.top = m - parseInt(e.clientHeight) + "px" : e.style.top = a.clientY - h + "px";
        0 >= a.clientX - d ? e.style.left = "0px" : a.clientX - d > k - parseInt(e.clientWidth) ? e.style.left = k - parseInt(e.clientWidth) + "px" : e.style.left = a.clientX - d + "px"
    }

    function f(a) {
        var c = {
            transform: "translate(0px, 0px) scale(1)"
        };
        a && (c = LGAIR_UI.isRequestFullscreen() ? {
            transform: "translate(0px, 0px) scale(1) rotate(-90deg)"
        } : {
            transform: "translate(0px, 0px) scale(1) rotate(90deg)"
        });
        $(".assistivetouch-menu-content").css(c)
    }
    // 移除悬浮按钮功能，直接显示右侧侧栏
    // 不再创建drager元素，只添加必要的tip和console元素
    document.getElementById("assistDiv").insertAdjacentHTML("beforeend", '<div id="tip" class="tip"></div><div id="console" class="console"></div>');
    0 < (a & 2) && ($("#assistivetouch-menu-content").css({
        height: "170px"
    }), $("#ctr_audio").css({
        transform: "translate(-80px, -40px)"
    }), $("#ctr_rotateMenu").css({
        transform: "translate(0px, -40px)"
    }), $("#ctr_fullscreen").css({
        transform: "translate(80px, -40px)"
    }), $("#ctr_goClean").hide(), $("#ctr_goHome").hide(), $("#ctr_goBack").hide(), $("#ctr_share").css({
        transform: "translate(-80px, 40px)"
    }), $("#ctr_tips").css({
        transform: "translate(0px, 40px)"
    }), $("#ctr_quality").css({
        transform: "translate(80px, 40px)"
    }));
    // 移除悬浮按钮相关的事件监听器和变量声明
    // 由于已经移除了悬浮按钮功能，这里不再需要触摸事件处理
    $(".assistivetouch-menu-content").on("click", function() {
        event.stopPropagation()
    });
    $("#ctr_audio").on("click", function() {
        1 == $("#ctr_audio").attr("state") ? (document.getElementById("gaudio").play(), document.getElementById("gaudio").muted = !1, $("#ctr_audio").attr("state", 0), $("#s_audio").html('<path d="M64 192v128h85.334L256 431.543V80.458L149.334 192H64zm288 64c0-38.399-21.333-72.407-53.333-88.863v176.636C330.667 328.408 352 294.4 352 256zM298.667 64v44.978C360.531 127.632 405.334 186.882 405.334 256c0 69.119-44.803 128.369-106.667 147.022V448C384 428.254 448 349.257 448 256c0-93.256-64-172.254-149.333-192z"></path>')) : (document.getElementById("gaudio").pause(), document.getElementById("gaudio").muted = !0, $("#ctr_audio").attr("state", 1), $("#s_audio").html('<path d="M405.5 256c0 22.717-4.883 44.362-13.603 63.855l31.88 31.88C439.283 323.33 448 290.653 448 256c0-93.256-64-172.254-149-192v44.978C361 127.632 405.5 186.882 405.5 256zM256 80.458l-51.021 52.48L256 183.957zM420.842 396.885L91.116 67.157l-24 24 90.499 90.413-8.28 10.43H64v128h85.334L256 431.543V280l94.915 94.686C335.795 387.443 318 397.213 299 403.022V448c31-7.172 58.996-22.163 82.315-42.809l39.61 39.693 24-24.043-24.002-24.039-.081.083z"></path><path d="M352.188 256c0-38.399-21.188-72.407-53.188-88.863v59.82l50.801 50.801A100.596 100.596 0 0 0 352.188 256z"></path>'))
    });
    $("#ctr_close").on("click", function() {
    });
    $("#ctr_exit").on("click", function() {});
    $("#ctr_tips").on("click", function() {
        var a = Date.parse(new Date);
        1E3 > a - LGAIR_UI.debug_clickTime && (LGAIR_UI.debug_clickCount++, 5 <= LGAIR_UI.debug_clickCount && (LGAIR_UI.debug_clickCount = 0, this.tipMode = this.tipMode & LGAIR_UI.TIP_CONSOLE ? this.tipMode ^ LGAIR_UI.TIP_CONSOLE : this.tipMode | LGAIR_UI.TIP_CONSOLE));
        LGAIR_UI.debug_clickTime = a;
        1 == $("#ctr_tips").attr("state") ? ($("#ctr_tips").attr("state", 0), $("#tip").hide(), $("#tips_lb").text("显示状态"), $("#s_tips").html('<path d="M255.8 112c-80.4 0-143.8 50.6-219.6 133.3-5.5 6.1-5.6 15.2-.1 21.3C101 338.3 158.2 400 255.8 400c96.4 0 168.7-77.7 220.1-134 5.3-5.8 5.6-14.6.5-20.7C424 181.8 351.5 112 255.8 112zm4.4 233.9c-53 2.4-96.6-41.2-94.1-94.1 2.1-46.2 39.5-83.6 85.7-85.7 53-2.4 96.6 41.2 94.1 94.1-2.1 46.2-39.5 83.6-85.7 85.7z"></path><path d="M256 209c0-6 1.1-11.7 3.1-16.9-1 0-2-.1-3.1-.1-36.9 0-66.6 31.4-63.8 68.9 2.4 31.3 27.6 56.5 58.9 58.9 37.5 2.8 68.9-26.9 68.9-63.8 0-1.3-.1-2.6-.1-3.9-5.6 2.5-11.7 3.9-18.2 3.9-25.2 0-45.7-21.1-45.7-47z"></path>'), this.tipMode & LGAIR_UI.TIP_CONSOLE && $("#console").hide()) : ($("#ctr_tips").attr("state", 1), $("#tip").show(), $("#tips_lb").text("关闭状态"), $("#s_tips").html('<path d="M88.3 68.1c-5.6-5.5-14.6-5.5-20.1.1-5.5 5.5-5.5 14.5 0 20l355.5 355.7c3.7 3.7 9 4.9 13.7 3.6 2.4-.6 4.6-1.9 6.4-3.7 5.5-5.5 5.5-14.5 0-20L88.3 68.1zM260.2 345.9c-53 2.4-96.6-41.2-94.1-94.1.6-12.2 3.6-23.8 8.6-34.3L121.3 164c-27.7 21.4-55.4 48.9-85.1 81.3-5.5 6.1-5.6 15.2-.1 21.3C101 338.3 158.2 400 255.8 400c29.7 0 57.1-7.4 82.3-19.2l-43.5-43.5c-10.6 5-22.2 8-34.4 8.6zM475.8 266c5.3-5.8 5.6-14.6.5-20.7C424 181.8 351.5 112 255.8 112c-29.1 0-56 6.6-82 19l43.7 43.7c10.5-5 22.1-8.1 34.3-8.6 53-2.4 96.6 41.2 94.1 94.1-.6 12.2-3.6 23.8-8.6 34.3l53.5 53.5c33-25.3 61.3-55.9 85-82z"></path><path d="M192.2 260.9c2.4 31.3 27.6 56.5 58.9 58.9 8.2.6 16.1-.3 23.4-2.6l-79.8-79.8c-2.2 7.4-3.1 15.3-2.5 23.5zM320 256c0-1.3-.1-2.6-.1-3.9-5.6 2.5-11.7 3.9-18.2 3.9-1.1 0-2.1 0-3.1-.1l18.6 18.7c1.8-5.9 2.8-12.2 2.8-18.6zM256 209c0-6 1.1-11.7 3.1-16.9-1 0-2-.1-3.1-.1-6.4 0-12.6 1-18.5 2.8l18.7 18.7c-.1-1.5-.2-3-.2-4.5z"></path>'), this.tipMode & LGAIR_UI.TIP_CONSOLE && $("#console").show())
    });
    $("#ctr_goHome").on("click", function() {
        window.h5_lgair.sendCmdEvent("goHome")
    });
    $("#ctr_goBack").on("click", function() {
        window.h5_lgair.sendCmdEvent("goBack")
    });
    $("#ctr_goClean").on("click", function() {
        window.h5_lgair.sendCmdEvent("goClean")
    });
    $("#ctr_share").on("click", function() {});
    $("#ctr_quality").on("click", function() {
        1 == $("#ctr_quality").attr("state") ? ($("#ctr_quality").attr("state", 2), $("#quality_lb").text("高清"), window.h5_lgair.sendMessage({
            id: "custom",
            data: btoa(JSON.stringify({
                cmd: "quality",
                value: 2
            }))
        })) : ($("#ctr_quality").attr("state", 1), $("#quality_lb").text("标清"), window.h5_lgair.sendMessage({
            id: "custom",
            data: btoa(JSON.stringify({
                cmd: "quality",
                value: 1
            }))
        }))
    });
    $("#ctr_fullscreen").on("click", function() {
        if (1 == $("#ctr_fullscreen").attr("state")) $("#ctr_fullscreen").attr("state", 0), LGAIR_UI.exitRequestFullscreen(), $("#fullscreen_lb").text("全屏");
        else {
            $("#ctr_fullscreen").attr("state", 1);
            $("#fullscreen_lb").text("退出全屏");
            var a = document.getElementById("root");
            LGAIR_UI.requestFullScreen(a)
        }
        f(!1)
    });
    $("#ctr_rotateMenu").on("click", function() {
        1 == $("#ctr_rotateMenu").attr("state") ? ($("#ctr_rotateMenu").attr("state", 0), f(!1)) : ($("#ctr_rotateMenu").attr("state", 1), f(!0))
    });
    $(".assistivetouch-menu-main").on("click", function(e) {
        e.stopPropagation()
    });
    $("#drager").on("click", function() {
        window.h5_lgair.be_Rotate && f(!1);
        $(".assistivetouch-menu-main").show();
        $(".assistivetouch-menu-main").css("pointer-events", "auto")
    });
};
LGAIR_UI.create_guid = function() {
    LGAIR_UI._debugCounts.create_guid = (LGAIR_UI._debugCounts.create_guid || 0) + 1;
    LGAIR_UI._leakLog("create_guid", LGAIR_UI._debugCounts.create_guid);
    document.getElementById("uicontainer").innerHTML = '\t<div id="main_ui">\t\t<div class="modal-container">\t\t\t<div class="modal-overlay">\t\t\t\t<div id="modal_mask" class="modal-content">\t\t\t\t\t<button id="btn_driver" class="tipButton">游戏连接成功, 点击开始游戏</button>\t\t\t\t</div>\t\t\t</div>\t\t</div></div>';
    $("#btn_driver").on("click", function() {
        $("#main_ui").hide();
        window.h5_lgair.play()
    })
};
LGAIR_UI.lock_screen = function(a, b) {
    if (!LGAIR_UI.screen_locked)
        if (LGAIR_UI._debugCounts.lock_screen = (LGAIR_UI._debugCounts.lock_screen || 0) + 1, LGAIR_UI._leakLog("lock_screen", LGAIR_UI._debugCounts.lock_screen), LGAIR_UI.screen_locked = !0, null == document.getElementById("main_ui2") ? document.getElementById("uicontainer").insertAdjacentHTML("beforeend", '\t<div id="main_ui2">\t\t<div class="modal-container">\t\t\t<div class="modal-overlay">\t\t\t\t<div id="modal_mask2" class="modal-content">\t\t\t\t\t<button id="btn_driver2" class="tipButton"></button>\t\t\t\t</div>\t\t\t</div>\t\t</div></div>') : $("#main_ui2").show(), $("#btn_driver2").text(a), void 0 != b) $("#btn_driver2").on("click", function() {
            $("#main_ui2").hide();
            b
        });
        else $("#btn_driver2").on("click", function() {})
};
LGAIR_UI.unlock_screen = function() {
    LGAIR_UI.screen_locked && (LGAIR_UI.screen_locked = !1, $("#main_ui2").hide())
};
LGAIR_UI.add_tips_msg = function(a) {
    null == document.getElementById("main_ui_tips") ? document.getElementById("uicontainer").insertAdjacentHTML("afterbegin", '\t<div id="main_ui_tips">\t\t<div class="modal-container">\t\t\t<div class="modal-overlay-none">\t\t\t\t<div id="tips_msg_mask" class="modal-content" style="margin-top:100px;">\t\t\t\t\t<div id="tips_msg_div" class="tipMsg"></div>\t\t\t\t</div>\t\t\t</div>\t\t</div></div>') : $("#main_ui_tips").show();
    $("#tips_msg_div").text(a);
    $("#tips_msg_div").show();
    $("#tips_msg_div").fadeOut(3E3, function() {
        $("#main_ui_tips").hide()
    })
};
LGAIR_UI.show_reconnect = function() {
    var a = document.getElementById("reconnect_tip");
    a || (document.body.insertAdjacentHTML("beforeend", '<div id="reconnect_tip" style="position:fixed;left:50%;bottom:16px;transform:translateX(-50%);z-index:200010;font-size:12px;color:#fff;background:rgba(0,0,0,0.6);padding:4px 10px;border-radius:12px;pointer-events:none;">正在重连中...0 s</div>'), a = document.getElementById("reconnect_tip"));
    a.style.display = "block";
    LGAIR_UI.reconnectStart = Date.now();
    LGAIR_UI.reconnectTimer && clearInterval(LGAIR_UI.reconnectTimer);
    LGAIR_UI.reconnectTimer = setInterval(function() {
        var b = Math.floor((Date.now() - LGAIR_UI.reconnectStart) / 1E3);
        a.textContent = "正在重连中..." + b + " s"
    }, 1E3)
};
LGAIR_UI.hide_reconnect = function() {
    LGAIR_UI.reconnectTimer && (clearInterval(LGAIR_UI.reconnectTimer), LGAIR_UI.reconnectTimer = null);
    var a = document.getElementById("reconnect_tip");
    a && (a.style.display = "none")
};
LGAIR_UI._currentProgressCleanup = null;
LGAIR_UI.create_progress = function(a, b, c, f, d, h) {
    var k = c - 12,
        m = -.5 * c;
    a = '<div id="progress_div" style="margin-left:' + m + "px;margin-top:" + m + "px;position:absolute;left:" + a + ";top:" + b + ';z-index:300001;"><img src="' + f + '" style="width:' + k + "px;height:" + k + 'px;left:6px;top:6px;border:0px;border-radius:50%;position:absolute;"></img><canvas pValue=0 id="progress_canvas" width="' + c + 'px" height="' + c + 'px" style="left:0px;top:0px;position:absolute;z-index:1;"></canvas><span id="pro_title" style="left:0px;top:' + (c + 10) + "px;width:" + c + 'px;text-align:center;display:block;word-break: keep-all;font-size:12px;position:absolute;z-index:1;"></span></div>';
    void 0 != h ? h.insertAdjacentHTML("beforeend", a) : document.body.insertAdjacentHTML("beforeend", a);
    $("#progress_div").show();
    void 0 != d && (document.getElementById("pro_title").innerText = d);
    var e = document.getElementById("progress_canvas"),
        l = e.getContext("2d"),
        r = c / 2,
        p = c / 2,
        n = 2 * Math.PI / 100,
        u = (c - 2) / 2;
    speed = .1;
    var stopAnimation = !1,
        localAnimationIds = [];
    (function y() {
        if (stopAnimation) {
            localAnimationIds = [];
            return;
        }
        var a = window.requestAnimationFrame(y, e);
        localAnimationIds.push(a);
        if (50 < localAnimationIds.length) {
            var b = localAnimationIds.shift();
            b && cancelAnimationFrame(b)
        }
        l.clearRect(0, 0, e.width, e.height);
        var c = parseFloat(e.getAttribute("pValue"));
        l.save();
        l.lineWidth = 2;
        l.beginPath();
        l.strokeStyle = "#00b8e0";
        l.arc(r, p, u, 0, 2 * Math.PI, !1);
        l.stroke();
        l.closePath();
        l.lineWidth = 3;
        l.save();
        l.beginPath();
        l.strokeStyle = "#f8f8f8";
        l.arc(r, p, u, -Math.PI / 2, -(Math.PI / 2 - c * n), !1);
        l.stroke();
        l.closePath();
        l.fillStyle = "#fff";
        l.font = "15px Arial";
        l.textAlign = "center";
        l.textBaseline = "middle";
        l.fillText(c.toFixed(0) + "%", r, p)
    })();
    LGAIR_UI._currentProgressCleanup = function() {
        stopAnimation = !0;
        localAnimationIds.forEach(function(a) {
            a && cancelAnimationFrame(a)
        });
        localAnimationIds = [];
        var a = document.getElementById("progress_div");
        a && a.parentNode && a.parentNode.removeChild(a)
    };
    return LGAIR_UI._currentProgressCleanup
};
LGAIR_UI.stopProgress = function() {
    if (LGAIR_UI._currentProgressCleanup) {
        LGAIR_UI._currentProgressCleanup();
        LGAIR_UI._currentProgressCleanup = null
    }
};
LGAIR_UI.progress_value = function(a, b) {
    100 < a && (a = 100);
    0 < a && document.getElementById("progress_canvas").setAttribute("pValue", a);
    void 0 != b && (document.getElementById("pro_title").innerText = b)
};
LGAIR_UI.progress_auto = function(a, b) {
    void 0 == b && (b = 100);
    0 > b && (b = 100);
    var c = 0;
    LGAIR_UI.auto_timer_hander = setInterval(function() {
        100 > c ? (99 < c || (c = 95 < c ? c + .01 * a : 90 < c ? c + .03 * a : 75 < c ? c + .1 * a : 60 < c ? c + .5 * a : 40 < c ? c + .8 * a : c + a), 100 <= LGAIR_UI.progress_value(c) && (clearInterval(LGAIR_UI.auto_timer_hander), LGAIR_UI.auto_timer_hander = null)) : (clearInterval(LGAIR_UI.auto_timer_hander), LGAIR_UI.auto_timer_hander = null)
    }, b), LGAIR_UI._debugCounts.autoTimerSet = (LGAIR_UI._debugCounts.autoTimerSet || 0) + 1, LGAIR_UI._leakLog("auto_timer_hander set", LGAIR_UI.auto_timer_hander)
};
LGAIR_UI.requestFullScreen = function(a) {
    navigator.keyboard && navigator.keyboard.lock();
    a = a ? a : document.documentElement;
    return a.requestFullscreen ? a.requestFullscreen() : a.webkitRequestFullScreen ? a.webkitRequestFullScreen() : a.mozRequestFullScreen ? a.mozRequestFullScreen() : a.msRequestFullscreen ? a.msRequestFullscreen() : Error("不支持全屏")
};
LGAIR_UI.exitRequestFullscreen = function() {
    this.isRequestFullscreen() && (document.exitFullscreen ? document.exitFullscreen() : document.mozCancelFullScreen ? document.mozCancelFullScreen() : document.webkitExitFullscreen && document.webkitExitFullscreen())
};
LGAIR_UI.isRequestFullscreen = function() {
    return !!(document.fullscreenElement || document.webkitFullscreenElement || document.mozFullScreenElement)
};
var LGAIR = function(a) {
    this.MSG_TYPE_START = 4278190081;
    this.MSG_TYPE_VIDEO = 4278190082;
    this.MSG_TYPE_AUDIO = 4278190085;
    this.MSG_TYPE_CTRL = 4278190086;
    this.MSG_TYPE_MESSAGE = 4278190087;
    this.MSG_INPUT_CTRL = 4278190088;
    this.VF_MPEG = 0;
    this.VF_H264 = 1;
    this.SDL_EVENT_MSGTYPE_NULL = 0;
    this.SDL_EVENT_MSGTYPE_KEYBOARD = 1;
    this.SDL_EVENT_MSGTYPE_MOUSEKEY = 2;
    this.SDL_EVENT_MSGTYPE_MOUSEMOTION = 3;
    this.SDL_EVENT_MSGTYPE_MOUSEWHEEL = 4;
    this.SDL_EVENT_MSGTYPE_KEYTEXT = 5;
    this.SDL_EVENT_MSGTYPE_KEYCMD = 6;
    this.SDL_EVENT_MSGTYPE_MOUSECLICK = 7;
    this.SDL_EVENT_MSGTYPE_POINTSCLICK = 8;
    this.SDL_EVENT_MSGTYPE_DEVCMD = 9;
    this.SDL_EVENT_MSGTYPE_CLIPTEXT = 10;
    this.ACTION_DOWN = 0;
    this.ACTION_UP = 1;
    this.ACTION_MOVE = 2;
    this.ACTION_POINTER_DOWN = 5;
    this.ACTION_POINTER_UP = 6;
    this.LOG_NONE = 0;
    this.LOG_ERROR = 1;
    this.LOG_WARNING = 2;
    this.LOG_DEBUG = 4;
    this.LOG_DATA = 8;
    this.LOG_VIDEO_DATA = 16;
    this.g_player = null;
    this.Canvas_left = this.Canvas_top = this.Canvas_height = this.Canvas_width = 0;
    this.Canvas_id = "videoCanvas";
    this.Canvas_id_ = "#" + this.Canvas_id;
    this.mobile_force = this.be_Rotate = !1;
    this.tid = 0;
    this.icon_size = this.icon_y = this.icon_x = this.icon_url = this.end_url = this.fail_url = "";
    this.touch_show = 1;
    this.vmod = LGAIR.VF_MPEG;
    this.log_index = this.loger_report_level = 0;
    this.clipMsg = this.handle_id = "";
    this.disable_click = 0;
    this.g_mouse_down = !1;
    this.touch_down_time = 0;
    this.g_input_mode = !1;
    this.TouchesList = [];
    this.last_rpc_time = 0;
    this.ctr_socket = null;
    this._eventHandlers = {};
    this.socket_protocl = "lgcloud";
    this.socket_addr = "ws://192.168.0.164:9080/lg?token=123";
    this.send_buffer = new ArrayBuffer(40);
    this.buffer_view = new DataView(this.send_buffer, 0, 28);
    this.buffer_view_sys = new DataView(this.send_buffer, 0, 20);
    this.buffer_view_multi = new DataView(this.send_buffer);
    this.g_mctrl_socket = null;
    this.send_txt_buffer = new Uint8Array(1024);
    this._config = a || {};
    void 0 != this._config.id ? this.Canvas_id = this._config.id : this._config.id = this.Canvas_id;
    void 0 != this._config.url && (this.socket_addr = this._config.url);
    void 0 != this._config.socket && (this.ctr_socket = this._config.socket);
    void 0 != this._config.vmod && (this.vmod = this._config.vmod);
    void 0 != this._config.log && (this.loger_report_level = this._config.log);
    void 0 != this._config.handle && (this.handle_id = this._config.handle);
    void 0 != this._config.clipMsg && (this.clipMsg = this._config.clipMsg);
    void 0 != this._config.mbm && (this.mobile_force = this._config.mbm);
    void 0 != this._config.tid && (this.tid = this._config.tid);
    void 0 != this._config.furl && (this.fail_url = this._config.furl);
    void 0 != this._config.eurl && (this.end_url = this._config.eurl);
    void 0 != this._config.ico && (this.icon_url = this._config.ico);
    void 0 != this._config.icox && (this.icon_x = this._config.icox);
    void 0 != this._config.icoy && (this.icon_y = this._config.icoy);
    void 0 != this._config.icosize && (this.icon_size = this._config.icosize);
    void 0 != this._config.tshow && (this.touch_show = this._config.tshow);
    this._webrtcRestartMs = void 0 != this._config.webrtcRestartMs ? this._config.webrtcRestartMs : 72E5;
    this._webrtcRestartTimer = null;
    this._webrtcHotkeyHandler = null;
    this._isRestarting = !1
};
LGAIR.prototype.start = function(a) {
    for (var b in a) this._config[b] = a[b];
    void 0 != this._config.useVideo ? (this.g_player = new VideoGame(this._config), this.g_player.start(), this.g_player.onVideoStart = function() {
        this.bind_event();
        this.on_play_start();
        LGAIR_UI.set_tips("tid:[" + this.tid + "]", LGAIR_UI.TIP_CONSOLE)
    }.bind(this), this.g_player.onDataChannelReady = function(a) {
        "" != this.clipMsg && this.send_input_txt(this.clipMsg, this.SDL_EVENT_MSGTYPE_CLIPTEXT)
    }.bind(this), this.g_player.onDatarecv = function(a) {
        this.fun_decode_data(a.data)
    }.bind(this), this.g_player.onClose = function(a) {
        this.fun_log("webRTC onClose 关闭连接...", this.LOG_DEBUG);
        this.on_play_end({
            text: "感谢体验..."
        }, 0)
    }.bind(this), this.g_player.onDisconnected = function() {
        this.fun_log("webRTC onDisconnected 错误...", this.LOG_ERROR);
        this.on_play_end({
            text: "连接断开，【轻触】重试..."
        }, 1)
    }.bind(this)) : this.fun_init_socket();
    void 0 != this._config.useMctrl && this._config.useMctrl && (this.g_mctrl_socket || (this.g_mctrl_socket = new MCtrlConn(this._config.token), this.g_mctrl_socket.start()));
    if (void 0 != this._config.useVideo) {
        this._webrtcRestartMs = void 0 != this._config.webrtcRestartMs ? this._config.webrtcRestartMs : this._webrtcRestartMs;
        this._webrtcRestartTimer && (clearInterval(this._webrtcRestartTimer), this._webrtcRestartTimer = null);
        0 < this._webrtcRestartMs && (this._webrtcRestartTimer = setInterval(function() {
            var a = !1;
            "undefined" !== typeof document && (a = !!document.hidden || "hidden" === document.visibilityState);
            !a && "undefined" !== typeof window && (a = !window.hasFocus());
            a && this._restartWebRTC && this._restartWebRTC()
        }.bind(this), this._webrtcRestartMs), LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("webrtc restart timer set", this._webrtcRestartTimer))
    }
    if (!this._webrtcHotkeyHandler) {
        this._webrtcHotkeyHandler = function(a) {
            if (a && a.ctrlKey && a.shiftKey && a.altKey && 75 === a.keyCode) {
                a.preventDefault();
                this._restartWebRTC && this._restartWebRTC()
            }
        }.bind(this);
        window.addEventListener("keydown", this._webrtcHotkeyHandler);
        LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("webrtc hotkey bound", "Ctrl+Shift+Alt+K")
    }
};
LGAIR.prototype._restartWebRTC = function() {
    try {
        LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("webrtc restart begin");
        this._isRestarting = !0;
        LGAIR_UI && LGAIR_UI.show_reconnect && LGAIR_UI.show_reconnect();
        this.unbind_event();
        this.g_player && (this.g_player.end(), this.g_player = null);
        try {
            var a = document.getElementById(this._config.id || "videoCanvas");
            a && a.parentNode && a.parentNode.removeChild(a);
            var b = document.getElementById("gaudio");
            b && b.parentNode && b.parentNode.removeChild(b)
        } catch (c) {
            LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("media element rebuild error", c)
        }
        this.g_player = new VideoGame(this._config);
        this.g_player.start();
        this.g_player.onVideoStart = function() {
            this.bind_event();
            this.on_play_start();
            LGAIR_UI.set_tips("tid:[" + this.tid + "]", LGAIR_UI.TIP_CONSOLE)
            this._isRestarting = !1;
            LGAIR_UI && LGAIR_UI.hide_reconnect && LGAIR_UI.hide_reconnect()
        }.bind(this);
        this.g_player.onDataChannelReady = function(a) {
            "" != this.clipMsg && this.send_input_txt(this.clipMsg, this.SDL_EVENT_MSGTYPE_CLIPTEXT)
        }.bind(this);
        this.g_player.onDatarecv = function(a) {
            this.fun_decode_data(a.data)
        }.bind(this);
        this.g_player.onClose = function(a) {
            this.fun_log("webRTC onClose 关闭连接...", this.LOG_DEBUG);
            this.on_play_end({
                text: "感谢体验..."
            }, 0)
        }.bind(this);
        this.g_player.onDisconnected = function() {
            this.fun_log("webRTC onDisconnected 错误...", this.LOG_ERROR);
            this.on_play_end({
                text: "连接断开，【轻触】重试..."
            }, 1)
        }.bind(this);
        LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("webrtc restart end")
    } catch (a) {
        this._isRestarting = !1;
        LGAIR_UI && LGAIR_UI.hide_reconnect && LGAIR_UI.hide_reconnect();
        LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("webrtc restart error", a)
    }
};
LGAIR.prototype.changeScreen = function(a) {
    // WebRTC 模式：视频流分辨率随设备旋转自动变化，#container 和 video 样式由 play.html CSS 控制
    // 无需任何 JS 干预布局，be_Rotate 永远为 false
    this.be_Rotate = !1;
    this.fun_log("changeScreen (WebRTC no-rotate mode), be_Rotate=false", this.LOG_DEBUG)
};
LGAIR.prototype.trans_xy = function(a, b) {
    var videoEl = document.getElementById(this.Canvas_id);
    var rect = videoEl.getBoundingClientRect();

    // 渲染区域尺寸（CSS像素）
    var renderW = rect.width;
    var renderH = rect.height;
    this.Canvas_top = Math.round(rect.top);
    this.Canvas_left = Math.round(rect.left);

    // 视频流原始分辨率（设备屏幕分辨率）
    var srcW = (videoEl.videoWidth && videoEl.videoWidth > 0) ? videoEl.videoWidth : renderW;
    var srcH = (videoEl.videoHeight && videoEl.videoHeight > 0) ? videoEl.videoHeight : renderH;

    // object-fit:contain 下视频实际显示区域（letterbox）
    var renderRatio = renderW / renderH;
    var srcRatio = srcW / srcH;
    var drawW, drawH, drawLeft, drawTop;
    if (srcRatio > renderRatio) {
        // 视频比容器更宽，两侧有黑边（上下letterbox不对，左右占满）
        drawW = renderW;
        drawH = renderW / srcRatio;
        drawLeft = 0;
        drawTop = (renderH - drawH) / 2;
    } else {
        // 视频比容器更高，左右有黑边
        drawH = renderH;
        drawW = renderH * srcRatio;
        drawLeft = (renderW - drawW) / 2;
        drawTop = 0;
    }

    // 鼠标相对于视频实际绘制区域的偏移
    var relX = (a - rect.left) - drawLeft;
    var relY = (b - rect.top) - drawTop;

    // 限制在绘制区域内
    relX = Math.max(0, Math.min(relX, drawW - 1));
    relY = Math.max(0, Math.min(relY, drawH - 1));

    // 按视频原始分辨率比例缩放，得到设备坐标
    var devX = Math.round(relX * srcW / drawW);
    var devY = Math.round(relY * srcH / drawH);

    // sw/sh 使用视频流原始分辨率，服务端据此做坐标归一化
    this.Canvas_width = srcW;
    this.Canvas_height = srcH;

    return { x: devX, y: devY };
};
LGAIR.prototype.IsPC = function() {
    if (this.mobile_force) return !1;
    for (var a = navigator.userAgent, b = "Android;iPhone;SymbianOS;Windows Phone;iPad;iPod".split(";"), c = !0, f = 0; f < b.length; f++)
        if (-1 != a.indexOf(b[f])) {
            c = !1;
            break
        }
    return c
};
LGAIR.prototype.fun_log = function(a, b) {
    var c = this.LOG_NONE;
    "undefined" != typeof b && (c = b);
    this.loger_report_level > this.LOG_NONE && this.loger_report_level & c && (++this.log_index, $.ajax({
        type: "POST",
        url: "logger.php?handle=" + this.handle_id,
        data: JSON.stringify({
            id: this.log_index,
            act: "jslog",
            msg: a
        }),
        datatype: "json",
        success: function(a) {},
        error: function(a) {}
    }))
};
LGAIR.prototype.app_call = function(a) {
    return window.LGAIR_APP ? window.LGAIR_APP.app_call(a) : "OUTAPP"
};
LGAIR.prototype.fun_rpc = function(a) {
    this.fun_log("SendRPC:" + JSON.stringify(a));
    $.ajax({
        type: "POST",
        url: "api_rpc.php?handle=" + this.handle_id,
        data: JSON.stringify(a),
        datatype: "json",
        success: function(a) {},
        error: function(a) {}
    })
};
LGAIR.prototype.send_ball_rpc = function() {};
LGAIR.prototype.save_data = function(a, b) {
    this.loger_report_level & b && $.ajax({
        url: "data.php?handle=" + this.handle_id,
        type: "POST",
        data: a,
        dataType: "JSON",
        processData: !1,
        contentType: !1,
        success: function(b) {
            console.log("send data ok size:" + a.length)
        }
    })
};
LGAIR.getAllParam = function() {
    var a = location.search;
    void 0 != window.super_args && (a = window.super_args);
    var b = null;
    if (-1 != a.indexOf("?"))
        for (var b = {}, a = a.substr(1).split("&"), c = 0; c < a.length; c++) {
            var f = a[c].split("=");
            if (2 < f.length) {
                var d = a[c].split(f[0] + "=");
                b[f[0]] = d[1]
            } else b[f[0]] = f[1]
        }
    return b
};
LGAIR.getKeyArgStr = function(a) {
    var b = "",
        c = this.getAllParam();
    if (null != c) {
        var f = [];
        if (0 < c.map.length && 0 < a.length) {
            for (b = 0; b < a.length; b++) "undefined" != typeof c[a[b]] && f.push(a[b] + "=" + c[a[b]]);
            b = f.join("&")
        }
    }
    return b
};
LGAIR.ip2long = function(a) {
    var b = !1;
    a.match(/^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$/) && (a = a.split("."), b = a[0] * Math.pow(256, 3) + a[1] * Math.pow(256, 2) + a[2] * Math.pow(256, 1) + a[3] * Math.pow(256, 0));
    return b << 0
};
LGAIR.long2ip = function(a) {
    a >>>= 0;
    var b = !1;
    !isNaN(a) && (0 <= a || 4294967295 >= a) && (b = Math.floor(a / Math.pow(256, 3)) + "." + Math.floor(a % Math.pow(256, 3) / Math.pow(256, 2)) + "." + Math.floor(a % Math.pow(256, 3) % Math.pow(256, 2) / Math.pow(256, 1)) + "." + Math.floor(a % Math.pow(256, 3) % Math.pow(256, 2) % Math.pow(256, 1) / Math.pow(256, 0)));
    return b
};
LGAIR.string2buffer = function(a) {
    for (var b = "", c = 0, c = 0; c < a.length; c++) b = "" === b ? a.charCodeAt(c).toString(16) : b + ("," + a.charCodeAt(c).toString(16));
    return new Uint8Array(b.match(/[\da-f]{2}/gi).map(function(a) {
        return parseInt(a, 16)
    }))
};
LGAIR.Utf8ArrayToStr = function(a) {
    var b, c, f, d, h, k;
    b = "";
    f = a.length;
    for (c = 0; c < f;) switch (d = a[c++], d >> 4) {
        case 0:
        case 1:
        case 2:
        case 3:
        case 4:
        case 5:
        case 6:
        case 7:
            b += String.fromCharCode(d);
            break;
        case 12:
        case 13:
            h = a[c++];
            b += String.fromCharCode((d & 31) << 6 | h & 63);
            break;
        case 14:
            h = a[c++], k = a[c++], b += String.fromCharCode((d & 15) << 12 | (h & 63) << 6 | (k & 63) << 0)
    }
    return b
};
LGAIR.str2utf8 = function(a) {
    for (var b, c = "", f = 0; f < a.length; f++) b = a.charCodeAt(f), 128 > b ? c += a.charAt(f) : 2048 > b ? (c += String.fromCharCode(192 | b >> 6 & 31), c += String.fromCharCode(128 | b >> 0 & 63)) : 65536 > b ? (c += String.fromCharCode(224 | b >> 12 & 15), c += String.fromCharCode(128 | b >> 6 & 63), c += String.fromCharCode(128 | b >> 0 & 63)) : this.fun_log("不是UCS-2字符集");
    return c
};
LGAIR.prototype.sendTouchStartEvent = function(a, b) {
    var c = this.trans_xy(a, b);
    this.g_mouse_down = !0;
    this.touch_down_time = Date.parse(new Date);
    LGAIR_UI.hide_paste_btn();
    this.fun_send_data({
        type: this.SDL_EVENT_MSGTYPE_MOUSEKEY,
        data: {
            press: 1,
            mx: c.x,
            my: c.y,
            btn: 1,
            sw: this.Canvas_width,
            sh: this.Canvas_height
        }
    })
};
LGAIR.prototype.sendTouchEndEvent = function(a, b) {
    var c = this.trans_xy(a, b);
    this.g_mouse_down = !1;
    this.fun_send_data({
        type: this.SDL_EVENT_MSGTYPE_MOUSEKEY,
        data: {
            press: 0,
            mx: c.x,
            my: c.y,
            btn: 1,
            sw: this.Canvas_width,
            sh: this.Canvas_height
        }
    });
    this.LongTouchEvent(Date.parse(new Date) - this.touch_down_time, a, b)
};
LGAIR.prototype.sendTouchMoveEvent = function(a, b) {
    var c = this.trans_xy(a, b);
    this.g_mouse_down && this.fun_send_data({
        type: this.SDL_EVENT_MSGTYPE_MOUSEMOTION,
        data: {
            press: 1,
            mx: c.x,
            my: c.y,
            btn: 1,
            sw: this.Canvas_width,
            sh: this.Canvas_height
        }
    })
};
LGAIR.prototype.setTouch = function(a) {
    var b = this.trans_xy(a.pageX, a.pageY),
        c = this.findTouchIndexById(a.identifier),
        f = -1; - 1 == c ? (this.TouchesList.push({
        x: b.x,
        y: b.y,
        id: a.identifier
    }), f = this.TouchesList.length - 1) : (this.TouchesList[c].x = b.x, this.TouchesList[c].y = b.y);
    return f
};
LGAIR.prototype.clearTouch = function(a) {
    a = this.findTouchIndexById(a.identifier); - 1 != a && (this.TouchesList[a].x = 0, this.TouchesList[a].y = 0);
    return a
};
LGAIR.prototype.findTouchIndexById = function(a) {
    for (var b = 0; b < this.TouchesList.length; b++)
        if (this.TouchesList[b].id == a) return b;
    return -1
};
LGAIR.prototype.sendTouchsEvent = function(a, b) {
    var c = -1,
        f;
    if (a == this.ACTION_MOVE)
        for (var d = 0; d < b.length; d++) this.setTouch(b[d]);
    else if (a == this.ACTION_POINTER_DOWN)
        for (d = 0; d < b.length; d++) f = this.setTouch(b[d]), -1 < f && -1 == c && (c = f);
    else if (a == this.ACTION_POINTER_UP)
        for (d = 0; d < b.length; d++) f = this.clearTouch(b[d]), -1 < f && -1 == c && (c = f); - 1 == c && (c = 0);
    d = this.TouchesList.length;
    a == this.ACTION_POINTER_UP && 1 == d && (d = 0);
    this.fun_send_data_multi({
        type: this.SDL_EVENT_MSGTYPE_POINTSCLICK,
        data: {
            points: d,
            action: a,
            nid: c,
            p: this.TouchesList,
            sw: this.Canvas_width,
            sh: this.Canvas_height
        }
    });
    if (a == this.ACTION_POINTER_UP)
        for (d = 0; d < b.length; d++) c = this.findTouchIndexById(b[d].identifier), -1 != c && this.TouchesList.splice(c, 1)
};
LGAIR.prototype.LongTouchEvent = function(a, b, c) {
    3 < a && (this.fun_log("LongTouchEvent body W:" + $(this.Canvas_id_).attr("width") + " H:" + $(this.Canvas_id_).attr("height"), this.LOG_DEBUG), document.getElementById(this.Canvas_id).getBoundingClientRect(), this.fun_log("LongTouchEvent canvas W:" + this.Canvas_width + " H:" + this.Canvas_height, this.LOG_DEBUG));
    2 <= a && LGAIR_UI.show_paste_btn(b, c)
};
LGAIR.prototype.sendCmdEvent = function(a) {
    var b, c, f;
    b = this.SDL_EVENT_MSGTYPE_KEYCMD;
    switch (a) {
        case "goClean":
            c = 257;
            f = 7;
            break;
        case "goHome":
            c = 257;
            f = 6;
            break;
        case "goBack":
            c = 257;
            f = 5;
            break;
        case "volUp":
            c = 257;
            f = 8;
            break;
        case "volDown":
            c = 257;
            f = 9;
            break;
        case "goClean":
            c = 257;
            f = 7;
            break;
        case "VK_BACK":
        case "VK_DELETE":
            c = 258;
            f = 1;
            break;
        case "VK_RETURN":
            c = 258;
            f = 2;
            break;
        case "VK_RIGHT":
            c = 258;
            f = 3;
            break;
        case "VK_LEFT":
            c = 258;
            f = 4;
            break;
        case "VK_UP":
            c = 258;
            f = 5;
            break;
        case "VK_DOWN":
            c = 258, f = 6
    }
    var d = 0;
    this.buffer_view.setUint32(d, this.MSG_TYPE_CTRL);
    d += 4;
    this.buffer_view.setUint32(d, 20);
    d += 4;
    this.buffer_view_sys.setUint16(d, 12);
    d += 2;
    this.buffer_view_sys.setUint8(d, b);
    d++;
    this.buffer_view_sys.setUint8(d, 0);
    d++;
    this.buffer_view_sys.setUint32(d, c);
    this.buffer_view_sys.setUint32(d + 4, f);
    this.g_player.sendData(this.buffer_view_sys);
    this.g_mctrl_socket && this.g_mctrl_socket.sendData(this.buffer_view_sys);
    this.fun_log("发送系统控制成功：" + a)
};
LGAIR.prototype.send_input_txt = function(a, b) {
    this.fun_log("send_input_txt:" + a);
    void 0 == b && (b = this.SDL_EVENT_MSGTYPE_KEYTEXT);
    var c, f, d, h, k = LGAIR.string2buffer(LGAIR.str2utf8(a));
    this.fun_log("txt_buf:" + LGAIR.Utf8ArrayToStr(k));
    for (var m = 0, e = k.byteLength, l = 50, r = 1; m < e; ++r) {
        m + l >= e ? (l = e - m, c = 0) : c = 1;
        var p = new Uint8Array(k.buffer, m, l);
        f = b;
        d = l;
        h = 6 + l;
        var n = 0,
            u = h + 8,
            q = new DataView(this.send_txt_buffer.buffer, 0, u);
        q.setUint32(n, this.MSG_TYPE_CTRL);
        n += 4;
        q.setUint32(n, u);
        n += 4;
        q.setUint16(n, h);
        n += 2;
        q.setUint8(n, f);
        n++;
        q.setUint8(n, c);
        n++;
        q.setUint16(n, d);
        n += 2;
        this.send_txt_buffer.set(p, n);
        this.g_player.sendData(q);
        this.g_mctrl_socket && this.g_mctrl_socket.sendData(q);
        this.fun_log("发送文本消息成功：" + p);
        m += l
    }
};
LGAIR.prototype.fun_init_socket = function() {
    if (!(this.ctr_socket instanceof WebSocket)) try {
        null != this.ctr_socket && (this.ctr_socket.close(), this.ctr_socket = null), this.ctr_socket = new WebSocket(this.socket_addr), this.ctr_socket.onopen = function(a) {
            this.fun_log("连接成功...", this.LOG_DEBUG)
        }.bind(this), this.ctr_socket.onmessage = function(a) {
            this.fun_decode_data(a.data)
        }.bind(this), this.ctr_socket.onclose = function(a) {
            this.ctr_socket = null;
            this.fun_log("关闭连接...", this.LOG_DEBUG);
            this.on_play_end({
                text: "感谢体验..."
            }, 0)
        }.bind(this), this.ctr_socket.onerror = function(a) {
            this.fun_log("websocket 错误...", this.LOG_ERROR);
            this.on_play_end({
                text: "发生错误，【轻触】重试..."
            }, 1)
        }.bind(this)
    } catch (a) {
        this.fun_log("有错误发生", this.LOG_ERROR)
    }
};
LGAIR.prototype.fun_send_data = function(a) {
    this.send_ball_rpc();
    a = a || {};
    var b = JSON.stringify(a);
    if ("" == b) this.fun_log("待发送数据为空！");
    else try {
        if (null != this.g_player.sendData) {
            this.fun_log("发送数据：" + b);
            var c = 0;
            this.buffer_view.setUint32(c, this.MSG_TYPE_CTRL);
            c += 4;
            this.buffer_view.setUint32(c, 28);
            c += 4;
            this.buffer_view.setUint16(c, 20);
            c += 2;
            this.buffer_view.setUint8(c++, a.type);
            this.buffer_view.setUint8(c++, 0);
            this.buffer_view.setUint8(c++, a.data.press);
            this.buffer_view.setUint8(c++, a.data.btn);
            this.buffer_view.setUint8(c++, 0);
            this.buffer_view.setUint8(c++, 0);
            this.buffer_view.setUint16(c, a.data.mx);
            c += 2;
            this.buffer_view.setUint16(c, a.data.my);
            c += 2;
            this.buffer_view.setUint16(c, 0);
            c += 2;
            this.buffer_view.setUint16(c, 0);
            c += 2;
            this.buffer_view.setUint16(c, a.data.sw);
            this.buffer_view.setUint16(c + 2, a.data.sh);
            this.g_player.sendData(this.buffer_view);
            this.g_mctrl_socket && this.g_mctrl_socket.sendData(this.buffer_view);
            this.fun_log("发送数据成功：" + b)
        }
    } catch (f) {
        this.fun_log("发送数据出错：" + b, this.LOG_ERROR)
    }
};
LGAIR.prototype.fun_send_data_multi = function(a) {
    this.send_ball_rpc();
    a = a || {};
    var b = JSON.stringify(a);
    if ("" == b) this.fun_log("待发送数据为空！");
    else try {
        if (null != this.g_player.sendData) {
            this.fun_log("发送数据：" + b);
            var c = 0;
            this.buffer_view_multi.setUint32(c, this.MSG_TYPE_CTRL);
            c += 4;
            this.buffer_view_multi.setUint32(c, 40);
            c += 4;
            this.buffer_view_multi.setUint16(c, 32);
            c += 2;
            this.buffer_view_multi.setUint8(c++, a.type);
            this.buffer_view_multi.setUint8(c++, 0);
            this.buffer_view_multi.setUint16(c, a.data.points);
            for (var c = c + 2, f = 5 < a.data.points ? 5 : a.data.points, d = 0, d = 0; d < f; ++d) this.buffer_view_multi.setUint16(c, a.data.p[d].x), c += 2;
            for (; 5 > d; ++d) this.buffer_view_multi.setUint16(c, 0), c += 2;
            for (d = 0; d < f; ++d) this.buffer_view_multi.setUint16(c, a.data.p[d].y), c += 2;
            for (; 5 > d; ++d) this.buffer_view_multi.setUint16(c, 0), c += 2;
            this.buffer_view_multi.setUint8(c, a.data.nid);
            c += 1;
            this.buffer_view_multi.setUint8(c, a.data.action);
            c += 1;
            this.buffer_view_multi.setUint16(c, a.data.sw);
            this.buffer_view_multi.setUint16(c + 2, a.data.sh);
            new Uint8Array(this.send_buffer);
            this.g_player.sendData(this.buffer_view_multi);
            this.g_mctrl_socket && this.g_mctrl_socket.sendData(this.buffer_view_multi);
            this.fun_log("发送多点触控数据成功：" + b)
        }
    } catch (h) {
        this.fun_log("发送多点触控数据出错：" + b, this.LOG_ERROR)
    }
};
LGAIR.prototype.fun_decode_data = function(a) {
    var b = new Uint8Array(a);
    a = new DataView(a);
    var c = 0,
        f = a.getUint32(c),
        c = c + 4,
        d = a.getUint32(c),
        c = c + 4;
    switch (f) {
        case this.MSG_TYPE_START:
            b = a.getUint32(c);
            a = a.getUint32(c + 4);
            this.fun_log("Video start W:" + b + " H:" + a + " vmode:" + this.vmod, this.LOG_DEBUG);
            "" != this.clipMsg && this.send_input_txt(this.clipMsg, this.SDL_EVENT_MSGTYPE_CLIPTEXT);
            break;
        case this.MSG_TYPE_VIDEO:
            this.save_data(b.subarray(c));
            this.g_player.decode(b.subarray(c, d));
            break;
        case this.MSG_TYPE_MESSAGE:
            this.on_play_end({
                text: "有用户正在体验，请稍候..."
            });
            this.fun_close_socket();
            break;
        case this.MSG_INPUT_CTRL:
            if (c += 3, b = a.getUint8(c), a = a.getUint8(c + 1), this.fun_log("Input msg m_type=" + b + " m_val=" + a), 2 == b)
                if (0 == a) this.on_input_end();
                else if (1 == a) this.on_input_start()
    }
};
LGAIR.prototype.fun_close_socket = function() {
    try {
        null != this.ctr_socket && (this.ctr_socket.close(), this.ctr_socket = null), this.fun_log("关闭socket成功", this.LOG_DEBUG)
    } catch (a) {
        this.fun_log("关闭socket出错", this.LOG_ERROR)
    }
};
LGAIR.prototype.unbind_event = function() {
    this.fun_log("unbind_event 清理事件监听器", this.LOG_DEBUG);
    LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("unbind_event", this._eventHandlers);
    var a = "onorientationchange" in window ? "orientationchange" : "resize";
    $(window).off(a, this._eventHandlers.resizeHandler);
    var b;
    "undefined" !== typeof document.hidden ? b = "visibilitychange" : "undefined" !== typeof document.msHidden ? b = "msvisibilitychange" : "undefined" !== typeof document.webkitHidden && (b = "webkitvisibilitychange");
    $(window).off(b, this._eventHandlers.visibilityHandler);
    document.removeEventListener("qbrowserVisibilityChange", this._eventHandlers.qbrowserHandler);
    $(this.Canvas_id_).off("touchstart", this._eventHandlers.touchStartHandler);
    $(this.Canvas_id_).off("touchend", this._eventHandlers.touchEndHandler);
    $(this.Canvas_id_).off("touchmove", this._eventHandlers.touchMoveHandler);
    $(this.Canvas_id_).off("mousedown", this._eventHandlers.mouseDownHandler);
    $(this.Canvas_id_).off("mouseup", this._eventHandlers.mouseUpHandler);
    $(this.Canvas_id_).off("mousemove", this._eventHandlers.mouseMoveHandler);
    this._eventHandlers = {}
};
LGAIR.prototype.bind_event = function() {
    var a = this;
    LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("bind_event", this._eventHandlers);
    this._eventHandlers.resizeHandler = this.changeScreen.bind(this);
    var b = "onorientationchange" in window ? "orientationchange" : "resize";
    $(window).on(b, this._eventHandlers.resizeHandler);
    var c, d;
    "undefined" !== typeof document.hidden ? (c = "hidden", d = "visibilitychange") : "undefined" !== typeof document.msHidden ? (c = "msHidden", d = "msvisibilitychange") : "undefined" !== typeof document.webkitHidden && (c = "webkitHidden", d = "webkitvisibilitychange");
    this._eventHandlers.visibilityHandler = function() {
        a.fun_log("event:" + d + " state:" + document.visibilityState, a.LOG_DEBUG);
        document[c] || "hidden" == document.visibilityState ? a.g_player && a.g_player.pause() : a.g_player && a.g_player.play()
    };
    $(window).on(d, this._eventHandlers.visibilityHandler);
    this._eventHandlers.qbrowserHandler = function(b) {
        b.hidden ? a.g_player && a.g_player.pause() : a.g_player && a.g_player.play()
    };
    document.addEventListener("qbrowserVisibilityChange", this._eventHandlers.qbrowserHandler);
    0 == this.disable_click && (this._eventHandlers.mouseDownHandler = function(b) {
        var c = b.clientX,
            d = b.clientY;
        1 == b.which && a.sendTouchStartEvent(c, d)
    }, this._eventHandlers.mouseUpHandler = function(b) {
        var c = b.clientX,
            d = b.clientY;
        1 == b.which && a.sendTouchEndEvent(c, d)
    }, this._eventHandlers.touchStartHandler = function(b) {
        b.preventDefault();
        a.sendTouchsEvent(a.ACTION_POINTER_DOWN, b.changedTouches);
        a.g_player.play()
    }, this._eventHandlers.touchEndHandler = function(b) {
        b.preventDefault();
        a.sendTouchsEvent(a.ACTION_POINTER_UP, b.changedTouches)
    }, this._eventHandlers.touchMoveHandler = function(b) {
        b.preventDefault();
        a.sendTouchsEvent(a.ACTION_MOVE, b.changedTouches)
    }, this._eventHandlers.mouseMoveHandler = function(b) {
        a.sendTouchMoveEvent(b.pageX, b.pageY)
    }, $(this.Canvas_id_).on("mousedown", this._eventHandlers.mouseDownHandler), $(this.Canvas_id_).on("mouseup", this._eventHandlers.mouseUpHandler), $(this.Canvas_id_).on("touchstart", this._eventHandlers.touchStartHandler), $(this.Canvas_id_).on("touchend", this._eventHandlers.touchEndHandler), $(this.Canvas_id_).on("touchmove", this._eventHandlers.touchMoveHandler), $(this.Canvas_id_).on("mousemove", this._eventHandlers.mouseMoveHandler));
    this.changeScreen(!0)
};
LGAIR.prototype.on_play_start = function() {
    this.fun_log("on_play_start：tid=" + this.tid, this.LOG_DEBUG);
    LGAIR_UI.video_start(this.touch_show)
};
LGAIR.prototype.on_play_end = function(a, b) {
    this.fun_log("on_play_end：" + b + " args:" + JSON.stringify(a), this.LOG_DEBUG);
    if (this._isRestarting) {
        LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("on_play_end ignored (restarting)");
        return
    }
    this.unbind_event();
    void 0 != b && (0 == b ? this.end_url ? (a.forcejump = !0, a.url = this.end_url) : a.url = window.location.href : 1 == b && (this.fail_url ? (a.forcejump = !0, a.url = this.fail_url) : a.url = window.location.href));
    LGAIR_UI.video_end(a);
    this.g_player.end();
    this.g_mctrl_socket && this.g_mctrl_socket.end();
    this.fun_close_socket()
};
LGAIR.prototype.on_input_start = function(a) {
    this.fun_log("on_input_start：stay text:" + a, this.LOG_DEBUG);
    LGAIR_UI.show_input(a);
    this.g_input_mode = !0
};
LGAIR.prototype.on_input_end = function(a) {
    this.fun_log("on_input_end:" + a, this.LOG_DEBUG);
    this.g_input_mode = !1;
    var b = LGAIR_UI.hide_input(2 == a ? !0 : !1);
    2 == a ? (0 < b.length && this.send_input_txt(b), this.sendCmdEvent("VK_RETURN")) : 1 == a && this.sendCmdEvent("goBack")
};
LGAIR.prototype.on_input_keydown = function(a) {
    if (13 == a.which) this.on_input_end(2);
    else 8 != a.which || LGAIR_UI.get_input_val() || (this.sendCmdEvent("VK_BACK"), this.fun_log("on_input_keydown:send del event", this.LOG_DEBUG))
};
LGAIR.prototype.sendMessage = function(a) {
    this.g_player.sendMessage(a)
};
LGAIR.prototype.play = function() {
    this.g_player.play()
};
LGAIR.prototype.destroy = function() {
    this.fun_log("LGAIR destroy 清理资源", this.LOG_DEBUG);
    LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("destroy");
    this.unbind_event();
    null != this.g_player && (this.g_player.end(), this.g_player = null);
    null != this.g_mctrl_socket && (this.g_mctrl_socket.end(), this.g_mctrl_socket = null);
    this.fun_close_socket();
    this.TouchesList = [];
    this.clipMsg = "";
    this.handle_id = "";
    this.on_play_end = null;
    this.on_play_start = null;
    if (window._h5LgairMessageHandler) {
        window.removeEventListener("message", window._h5LgairMessageHandler);
        LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("message handler removed");
        window._h5LgairMessageHandler = null
    }
    if (this._webrtcHotkeyHandler) {
        window.removeEventListener("keydown", this._webrtcHotkeyHandler);
        this._webrtcHotkeyHandler = null;
        LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("webrtc hotkey removed")
    }
    if (this._webrtcRestartTimer) {
        clearInterval(this._webrtcRestartTimer);
        this._webrtcRestartTimer = null;
        LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("webrtc restart timer cleared")
    }
    if (typeof LGAIR_UI.stopAllTimers === "function") {
        LGAIR_UI.stopAllTimers()
    }
    if (window.h5_lgair_wait_interval) {
        clearInterval(window.h5_lgair_wait_interval);
        LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("wait_interval cleared", window.h5_lgair_wait_interval);
        window.h5_lgair_wait_interval = null
    }
};
$(function() {
    var a = 0,
        b = Date.parse(new Date),
        c = 0,
        f = 2,
        d = "123",
        h = "",
        k = "",
        m = 0,
        e = "yun6.lgyouxi.cn",
        l = 8198,
        r = 8200,
        p = "ws://",
        n = 1,
        u = 0,
        q = 0,
        y = "cloud.lgair.cn",
        B = 0,
        D = "",
        E = "",
        x = "",
        F = "",
        G = "",
        H = "",
        C = "wsa.lgair.cn",
        I = 1,
        z = "",
        J = !0,
        g = LGAIR.getAllParam();
    if (null != g) {
        "undefined" != typeof g.v && "h264" == g.v && (c = 1);
        "undefined" != typeof g.q && (f = parseInt(g.q), 1 > f && (f = 1), 2 < f && (f = 2));
        "undefined" != typeof g.token && (d = g.token);
        "undefined" != typeof g.i && (h = g.i);
        "undefined" != typeof g.p && (l = g.p);
        "undefined" != typeof g.shost && (e = g.shost);
        "undefined" != typeof g.sport && (r = g.sport);
        if ("undefined" != typeof g.pf) {
            var t = parseInt(g.pf);
            if (0 == t || 1 == t) n = t
        }
        "undefined" != typeof g.dm && (t = parseInt(g.dm), 0 == t || 1 == t) && (q = t);
        "undefined" != typeof g.log && (a = g.log);
        "undefined" != typeof g.handle && (b = g.handle);
        "undefined" != typeof g.msg && (k = decodeURIComponent(g.msg));
        "undefined" != typeof g.mbm && (m = g.mbm);
        "undefined" != typeof g.dc && (u = g.dc);
        "undefined" != typeof g.line && (y = g.line);
        "undefined" != typeof g.tid && (B = g.tid);
        "undefined" != typeof g.furl && (D = decodeURIComponent(g.furl));
        "undefined" != typeof g.eurl && (E = decodeURIComponent(g.eurl));
        "undefined" != typeof g.ico && (x = decodeURIComponent(g.ico));
        "undefined" != typeof g.icox && (F = decodeURIComponent(g.icox));
        "undefined" != typeof g.icoy && (G = decodeURIComponent(g.icoy));
        "undefined" != typeof g.icosize && (H = decodeURIComponent(g.icosize));
        "undefined" != typeof g.tshow && (I = parseInt(g.tshow));
        "undefined" != typeof g.wssproxy && (C = g.wssproxy);
        "undefined" != typeof g.rtc_i && "undefined" != typeof g.rtc_p && (z = g.rtc_i + " " + g.rtc_p);
        "undefined" != typeof g.mctrl && (J = parseInt(g.mctrl));
        var R = 120;
        "undefined" != typeof g.wr && (R = parseInt(g.wr))
    }
    var v = document.documentElement.clientWidth,
        w = document.documentElement.clientHeight;
    v < w && (t = v, v = w, w = t);
    w = Math.round(1280 * w / v);
    v = 1280;
    "https:" === window.location.protocol && (p = "wss://");
    src_url = "ws://192.168.147.145:9006/lgcloud?user=hl&os=mobile";
    src_url = "ws://192.168.54.99:9016/lgcloud?user=hl&os=mobile";
    src_url = "ws://192.168.147.145:8200/lgcloud?user=hl&os=mobile";
    src_url = p + e + ":" + r + "/lgcloud?user=hl&os=mobile&token=" + d + "&type=" + c + "&quality=" + f + "&platform=" + n + "&dm=" + q + "&width=" + v + "&height=" + w;
    var A = new LGAIR({
        size: {
            width: 800,
            height: 480
        },
        id: "videoCanvas",
        url: src_url,
        vmod: c,
        log: a,
        handle: b,
        clipMsg: k,
        mbm: m,
        dc: u,
        useVideo: !0,
        tid: B,
        furl: D,
        eurl: E,
        ico: x,
        icox: x,
        icoy: x,
        icosize: x,
        tshow: I,
        useMctrl: J,
        token: d,
        webrtcRestartMs: 0 < R ? 6E4 * R : 0
    });
    window.h5_lgair = A;
    var K = !1;
    window.h5_lgair_wait_interval = null;
    if ("undefined" == typeof g.shost && "undefined" == typeof g.sport && "undefined" != typeof g.p) {
        $.ajax({
            type: "GET",
            url: window.location.protocol + "//" + y + "/get_addr.php?i=" + h + "&p=" + l,
            timeout: 5E3,
            datatype: "json",
            success: function(a) {
                a = JSON.parse(a);
                "" != a.i && "" != a.wp && (e = a.i, r = a.wp);
                "" != a.i && "" != a.p && (z = a.i + " " + a.p);
                void 0 != a.wssp && "" != a.wssp && (C = a.wssp)
            },
            error: function(a) {},
            complete: function() {
                K = !0
            }
        });
        window.h5_lgair_wait_interval = setInterval(function() {
            K && (src_url = "wss://" == p ? p + C + "/" + e + "/" + r : p + e + ":" + r, src_url = src_url + "/lgcloud?user=hl&os=mobile&token=" + d + "&type=" + c + "&quality=" + f + "&platform=" + n + "&dm=" + q + "&width=" + v + "&height=" + w, A.start({
                url: src_url,
                rtc_addr: z
            }), window.h5_lgair_wait_interval && (clearInterval(window.h5_lgair_wait_interval), window.h5_lgair_wait_interval = null))
        }, 200), LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("wait_interval set", window.h5_lgair_wait_interval)
    } else A.start({
        url: src_url,
        rtc_addr: z
    });
    LGAIR_UI.create_style();
    LGAIR_UI.create_logo({
        icon_url: x,
        icon_x: F,
        icon_y: G,
        icon_size: H
    });
    LGAIR_UI.progress_value(0, "加载中[" + B + "]");
    if (!window._h5LgairMessageHandler) {
        window._h5LgairMessageHandler = function(a) {
            if (a.data && "close" === a.data.type && A && A.on_play_end) {
                A.on_play_end({
                    text: "即将关闭..."
                })
            }
        };
        window.addEventListener("message", window._h5LgairMessageHandler, !1);
        LGAIR_UI && LGAIR_UI._leakLog && LGAIR_UI._leakLog("message handler added")
    }
});
