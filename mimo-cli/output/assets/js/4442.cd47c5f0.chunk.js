"use strict";(self.webpackChunkmimo_chat=self.webpackChunkmimo_chat||[]).push([[4442],{14442:(function(ie,I,t){t.d(I,{DX:function(){return F.DX}});var g=t(42461),J=t.n(g),o=t(87155),n=t(44012),B=t(81978),K=t(30759),O=t(21282),P=t(52346),A=t(73532),C=t(71636),T=t(50240),N=t(16945),b=t(83212),e=t(47440),F=t(71572);const V=5,_=(0,o.PA)(({open:v,onClose:H,onChange:X,maxDuration:G=600,supportedMimeTypes:re,readingText:Y})=>{const{t:u}=(0,B.Bd)(),f=(0,n.useRef)(v),l=(0,n.useRef)(!1),d=(0,n.useRef)(()=>{}),m=(0,n.useRef)(null),{recordingState:a,duration:R,isSilent:U,analyserNode:Q,audioDevices:Z,selectedDeviceId:$,setSelectedDeviceId:k,startStream:D,stopStream:E,start:p,stop:x,pause:q,resume:z}=(0,b.A)({maxDuration:G,supportedMimeTypes:re,silenceDetectionDelay:V,onComplete:r=>{X==null||X(r),d.current()},onError:r=>{const M=r instanceof Error?r.message:String(r);console.error("\u5F55\u5236\u5931\u8D25\uFF0C\u8BF7\u91CD\u8BD5\uFF0C\u9519\u8BEF\u539F\u56E0\uFF1A",M,r);const s=M.includes("\u4E0D\u652F\u6301\u89E3\u7801")||M.includes("AUDIO_DECODE_NOT_SUPPORTED")||M.includes("Unable to decode")||M.includes("EncodingError");C.A.error(u(s?"chat.audio.export_not_supported":"chat.audio.recording_failed")),d.current()}}),S=(0,n.useCallback)(()=>{l.current||(l.current=!0,E(),f.current&&H())},[H,E]),ee=(0,n.useCallback)(()=>{x(!1)},[x]);d.current=S,(0,n.useEffect)(()=>{const r=m.current;if(r)return Q&&a==="recording"?r.connectAnalyser(Q):r.disconnectAudio(),()=>{r.disconnectAudio()}},[Q,a]);const te=(0,n.useCallback)(()=>{x(!1),S()},[S,x]);return(0,n.useEffect)(()=>{f.current=v,v&&(l.current=!1)},[v]),(0,n.useEffect)(()=>(v&&D(),()=>{x(!1),E()}),[v]),(0,n.useEffect)(()=>{v&&$&&a==="idle"&&D()},[$]),(0,e.jsxs)(P.A,{open:v,onClose:te,title:u("chat.audio.title"),size:"md",allowClickOutside:!1,showCloseButton:!0,classNames:{container:"max-w-[37.5rem]"},children:[(0,e.jsxs)("div",{className:"flex flex-col items-center gap-2 px-6 pb-[2.625rem] pt-[2.625rem]",children:[Y&&(0,e.jsxs)("div",{className:"flex w-full flex-col items-center gap-2 pb-4",children:[(0,e.jsx)("p",{className:"select-none text-xs leading-5 text-mimo-text-h2-caption",children:u("chat.audio.reading_hint")}),(0,e.jsxs)("p",{className:"w-[65%] select-none text-center text-sm font-medium leading-5 text-mimo-text-h1-title",children:["\u201C",Y,"\u201D"]})]}),(0,e.jsx)(T.A,{ref:m,className:"h-[4.1875rem] w-[15.375rem]",forceIdle:a==="idle",speed:.3}),(0,e.jsx)("p",{className:"h-5 select-none text-xs leading-5 text-mimo-text-h3-placeholder",children:U?u("chat.audio.silence_detected"):a==="idle"?null:u("chat.audio.recording")})]}),(0,e.jsx)("div",{className:"flex items-center justify-between px-6 py-[0.8125rem]",children:a==="idle"?(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(A.Ay,{options:Z.map((r,M)=>({key:r.deviceId,value:r.deviceId,label:r.label||`${u("chat.audio.microphone")} ${M+1}`})),value:$,onChange:r=>k(r),renderTrigger:r=>(0,e.jsxs)("div",{className:"box-border flex cursor-pointer select-none items-center gap-1 rounded-md border border-mimo-line-border-card px-2 py-1 text-mimo-text-h1-title transition-colors",children:[(0,e.jsx)(K.A,{className:"h-4 w-4"}),(0,e.jsx)("span",{className:"max-w-48 truncate text-sm",children:(r==null?void 0:r.label)||u("chat.audio.select_microphone")})]}),matchTriggerWidth:!1,classNames:{trigger:"rounded-md hover:bg-mimo-fill-neutral-hover transition-colors cursor-pointer"}}),(0,e.jsx)(O.Ay,{size:"sm",variant:"outline",theme:"neutral",onClick:p,className:"box-border min-w-20 px-2 py-1 hover:!bg-mimo-btn-neutral-bg-hover",children:u("chat.audio.start_recording")})]}):(0,e.jsxs)(e.Fragment,{children:[(0,e.jsxs)("div",{className:"flex items-center gap-2",children:[(0,e.jsx)(O.K0,{onClick:a==="recording"?q:z,className:"h-6 w-6 !text-mimo-icon-n2",children:a==="recording"?(0,e.jsxs)("svg",{className:"h-4 w-4",fill:"currentColor",viewBox:"0 0 24 24",children:[(0,e.jsx)("rect",{x:"6",y:"4",width:"4",height:"16"}),(0,e.jsx)("rect",{x:"14",y:"4",width:"4",height:"16"})]}):(0,e.jsx)("svg",{className:"h-4 w-4",fill:"currentColor",viewBox:"0 0 24 24",children:(0,e.jsx)("path",{d:"M8 5v14l11-7z"})})}),(0,e.jsx)("span",{className:"font-medium text-mimo-text-h1-title",children:(0,N.a)(R)}),(0,e.jsxs)("span",{className:"text-mimo-text-h4-disable",children:["/ ",(0,N.a)(G)]})]}),(0,e.jsxs)("div",{className:"flex items-center gap-2",children:[U&&(0,e.jsx)(O.Ay,{size:"sm",variant:"outline",theme:"neutral",onClick:ee,className:"box-border min-w-20 px-2 py-1",children:u("chat.audio.re_record")}),(0,e.jsx)(O.Ay,{size:"sm",variant:"solid",theme:"error",onClick:()=>x(!0),className:"box-border min-w-20 px-2 py-1",children:u("chat.audio.stop_recording")})]})]})})]})});I.Ay=_}),30759:(function(ie,I,t){var g=t(44012),J,o;function n(){return n=Object.assign?Object.assign.bind():function(P){for(var A=1;A<arguments.length;A++){var C=arguments[A];for(var T in C)({}).hasOwnProperty.call(C,T)&&(P[T]=C[T])}return P},n.apply(null,arguments)}const B=(P,A)=>g.createElement("svg",n({xmlns:"http://www.w3.org/2000/svg",fill:"none",viewBox:"0 0 16 16",ref:A},P),J||(J=g.createElement("path",{fill:"currentColor",d:"M8 .667A3.333 3.333 0 0 0 4.667 4v3.333a3.333 3.333 0 1 0 6.666 0V4A3.333 3.333 0 0 0 8 .667m2 6.666a2 2 0 1 1-4 0V4a2 2 0 1 1 4 0z"})),o||(o=g.createElement("path",{fill:"currentColor",d:"M13.333 6a.667.667 0 0 0-.666.667v.666a4.667 4.667 0 0 1-9.334 0v-.666a.667.667 0 1 0-1.333 0v.666a6 6 0 0 0 5.333 5.964V15a.667.667 0 0 0 1.334 0v-1.703A6 6 0 0 0 14 7.333v-.666A.667.667 0 0 0 13.333 6"}))),K=(0,g.forwardRef)(B),O=(0,g.memo)(K);I.A=O}),50240:(function(ie,I,t){var g=t(82759),J=t.n(g),o=t(44012),n=t(81072),B=t(1627),K=t(73865),O=t(19704),P=t(47440);const A=(0,O.h)("AudioWaveVisualizer"),C=`
  uniform float uWidth;
  uniform float uLength;
  uniform float uAmplitude;
  uniform float uStep;
  uniform float uFrequency;
  uniform float uComplexity;
  uniform float uConcentration;

  varying vec3 vNormal;
  varying float horizontalPosition;
  varying vec3 vWorldPosition;

  vec2 randomGradient(vec2 p) {
    float angle = fract(sin(dot(p.xy, vec2(12.9898, 78.233))) * 43758.5453123) * 6.28318530718;
    return vec2(cos(angle), sin(angle));
  }

  float perlinNoise(vec2 st) {
    vec2 i = floor(st);
    vec2 f = fract(st);

    vec2 g00 = randomGradient(i);
    vec2 g10 = randomGradient(i + vec2(1.0, 0.0));
    vec2 g01 = randomGradient(i + vec2(0.0, 1.0));
    vec2 g11 = randomGradient(i + vec2(1.0, 1.0));

    vec2 p00 = f - vec2(0.0, 0.0);
    vec2 p10 = f - vec2(1.0, 0.0);
    vec2 p01 = f - vec2(0.0, 1.0);
    vec2 p11 = f - vec2(1.0, 1.0);

    float dot00 = dot(g00, p00);
    float dot10 = dot(g10, p10);
    float dot01 = dot(g01, p01);
    float dot11 = dot(g11, p11);

    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(dot00, dot10, u.x), mix(dot01, dot11, u.x), u.y);
  }

  float getOffset(vec2 _uv) {
    vec2 st = vec2(-_uv.x * uFrequency + uStep, _uv.y * uComplexity + uStep * 2.0);
    float scale = pow(1.0 - (cos(_uv.x * 2.0 * 3.1415926) + 1.0) * 0.5, uConcentration) * uAmplitude;
    return perlinNoise(st) * scale;
  }

  void main() {
    float differential = 0.1;
    vec3 newPosition = position;
    vec2 vUv = uv;
    horizontalPosition = uv.x;

    vec2 uvX = vUv + vec2(differential / uLength, 0.0);
    vec2 uvY = vUv + vec2(0.0, differential / uWidth);
    float offsetDfX = getOffset(uvX);
    float offsetDfY = getOffset(uvY);
    float offset = getOffset(vUv);

    vec3 dfX = vec3(differential, offsetDfX - offset, 0.0);
    vec3 dfY = vec3(0.0, offsetDfY - offset, differential);
    vNormal = cross(dfY * 100.0, dfX * 100.0);

    newPosition.z += offset;

    vec4 worldPosition = modelMatrix * vec4(newPosition, 1.0);
    vWorldPosition = worldPosition.xyz;
    vWorldPosition.y = offset;

    gl_Position = projectionMatrix * modelViewMatrix * vec4(newPosition, 1.0);
  }
`,T=`
  uniform float uWidth;
  uniform vec3 uLeftColor;
  uniform vec3 uRightColor;
  uniform vec3 uCameraPosition;
  uniform float uAlphaBase;
  uniform float uAlphaRim;

  varying vec3 vNormal;
  varying float horizontalPosition;
  varying vec3 vWorldPosition;

  void main() {
    float depthValue = 1.0 - (vWorldPosition.z / uWidth + 0.5);
    vec3 viewDirection = normalize(vWorldPosition - uCameraPosition);
    vec3 normalizedNormal = normalize(vNormal);
    float sinTheta = length(cross(normalizedNormal, viewDirection));

    vec3 color = mix(uLeftColor, uRightColor, horizontalPosition);
    gl_FragColor = vec4(color, (uAlphaBase + pow(sinTheta, 10.0) * uAlphaRim) * depthValue);
  }
`,N=80,b=10,e=10,F="#f56bff",V="#80d9ff",_={uAmplitude:20,uConcentration:1,uComplexity:4},v={uAmplitude:28,uComplexity:6},H=.3,X=.04,G=(0,o.forwardRef)(({className:re,forceIdle:Y=!1,speed:u=.3,leftColor:f,rightColor:l,alphaBase:d,alphaRim:m,idleAmplitude:a,activeAmplitude:R,simulateActive:U=!1},Q)=>{const Z=(0,o.useRef)(null),$=(0,o.useRef)(null),k=(0,o.useRef)(null),D=(0,o.useRef)(null),E=(0,o.useRef)(null),p=(0,o.useRef)(null),x=(0,o.useRef)(20),q=(0,o.useRef)({..._}),z=(0,o.useRef)(null),S=(0,o.useRef)({leftColor:f!=null?f:F,rightColor:l!=null?l:V,alphaBase:d!=null?d:.1,alphaRim:m!=null?m:.22}),ee=(0,o.useRef)(a!=null?a:_.uAmplitude),te=(0,o.useRef)(R!=null?R:v.uAmplitude),r=(0,o.useRef)(U);(0,o.useEffect)(()=>{S.current={leftColor:f!=null?f:F,rightColor:l!=null?l:V,alphaBase:d!=null?d:.1,alphaRim:m!=null?m:.22};const s=k.current;s&&(s.uniforms.uLeftColor.value=new n.Q1f(f!=null?f:F),s.uniforms.uRightColor.value=new n.Q1f(l!=null?l:V),s.uniforms.uAlphaBase.value=d!=null?d:.1,s.uniforms.uAlphaRim.value=m!=null?m:.22)},[f,l,d,m]),(0,o.useEffect)(()=>{ee.current=a!=null?a:_.uAmplitude},[a]),(0,o.useEffect)(()=>{te.current=R!=null?R:v.uAmplitude},[R]),(0,o.useEffect)(()=>{r.current=U},[U]);const M=()=>{if(!D.current)return 0;const s=new Uint8Array(D.current.fftSize);D.current.getByteTimeDomainData(s);let c=0;for(const i of s){const y=Math.abs(i-128)/128;y>c&&(c=y)}return c};return(0,o.useEffect)(()=>{const s=Z.current;if(!s)return;const c=s.parentElement;if(!c)return;let i;try{i=new B.JeP({canvas:s,antialias:!0,alpha:!0})}catch(L){A.warn("WebGL \u4E0A\u4E0B\u6587\u521B\u5EFA\u5931\u8D25\uFF0CAudioWaveVisualizer \u8DF3\u8FC7\u6E32\u67D3\uFF1A",L);return}i.setPixelRatio(window.devicePixelRatio),i.setClearColor(0,0),$.current=i;const y=new n.Z58,h=new n.ubm(60,1,.1,500);h.position.set(0,1,e),h.rotation.set(-.1,0,0);const j=S.current,w=new n.BKk({uniforms:{uWidth:{value:b},uLength:{value:N},uAmplitude:{value:_.uAmplitude},uStep:{value:x.current},uFrequency:{value:1},uComplexity:{value:_.uComplexity},uConcentration:{value:_.uConcentration},uLeftColor:{value:new n.Q1f(j.leftColor)},uRightColor:{value:new n.Q1f(j.rightColor)},uCameraPosition:{value:h.position.clone()},uAlphaBase:{value:j.alphaBase},uAlphaRim:{value:j.alphaRim}},vertexShader:C,fragmentShader:T,side:n.$EB,transparent:!0});k.current=w;const _e=N*10,Ee=b*10,ae=new n.bdM(N,b,_e,Ee),ue=new n.eaF(ae,w);ue.rotation.set(-Math.PI/2,0,0),y.add(ue);const le=()=>{const L=c.clientWidth,ne=c.clientHeight;i.setSize(L,ne);const oe=L/ne;h.fov=2*Math.atan(Math.tan(162*Math.PI/360)/oe)*180/Math.PI,h.aspect=oe,h.updateProjectionMatrix()};le();const de=new ResizeObserver(le);de.observe(c);let fe=performance.now();const me=()=>{z.current=requestAnimationFrame(me);const L=performance.now(),ne=(L-fe)/1e3;fe=L,x.current+=ne*u*5,w.uniforms.uStep.value=x.current;const oe=Y?0:r.current?1:M(),xe={..._,uAmplitude:ee.current},he={..._,...v,uAmplitude:te.current},ve=oe>X?he:xe,se=q.current,ce={...se};for(const W of Object.keys(se))W in ve&&(ce[W]=(1-H)*se[W]+H*ve[W]);q.current=ce;for(const[W,ge]of Object.entries(ce))w.uniforms[W]&&(w.uniforms[W].value=ge);w.needsUpdate=!0,i.render(y,h)};return me(),()=>{z.current!==null&&(cancelAnimationFrame(z.current),z.current=null),de.disconnect(),ae.dispose(),w.dispose(),i.dispose()}},[Y,u]),(0,o.useImperativeHandle)(Q,()=>({connectAudio:s=>{var h,j;(h=p.current)==null||h.disconnect(),(j=E.current)==null||j.close().catch(()=>{});const c=new AudioContext,i=c.createAnalyser();i.fftSize=512;const y=c.createMediaElementSource(s);y.connect(i),i.connect(c.destination),E.current=c,D.current=i,p.current=y},connectAnalyser:s=>{var c,i;(c=p.current)==null||c.disconnect(),(i=E.current)==null||i.close().catch(()=>{}),E.current=null,p.current=null,D.current=s},disconnectAudio:()=>{var s,c;(s=p.current)==null||s.disconnect(),(c=E.current)==null||c.close().catch(()=>{}),D.current=null,E.current=null,p.current=null}})),(0,P.jsx)("div",{className:(0,K.cn)("relative overflow-hidden",re),children:(0,P.jsx)("canvas",{ref:Z,className:"block h-full w-full"})})});G.displayName="AudioWaveVisualizer",I.A=G})}]);

//# sourceMappingURL=4442.cd47c5f0.chunk.js.map