"use strict";(self.webpackChunkmimo_chat=self.webpackChunkmimo_chat||[]).push([[1627],{1627:(function(bf,Fo,Xi){Xi.d(Fo,{JeP:function(){return Pr}});var Tf=Xi(42461),qr=Xi(82759),f=Xi(81072);/**
 * @license
 * Copyright 2010-2026 Three.js Authors
 * SPDX-License-Identifier: MIT
 */function Oo(){let o=null,p=!1,h=null,x=null;function E(S,I){h(S,I),x=o.requestAnimationFrame(E)}return{start:function(){p!==!0&&h!==null&&(x=o.requestAnimationFrame(E),p=!0)},stop:function(){o.cancelAnimationFrame(x),p=!1},setAnimationLoop:function(S){h=S},setContext:function(S){o=S}}}function Bo(o){const p=new WeakMap;function h(V,j){const $=V.array,ht=V.usage,K=$.byteLength,F=o.createBuffer();o.bindBuffer(j,F),o.bufferData(j,$,ht),V.onUploadCallback();let G;if($ instanceof Float32Array)G=o.FLOAT;else if(typeof Float16Array<"u"&&$ instanceof Float16Array)G=o.HALF_FLOAT;else if($ instanceof Uint16Array)V.isFloat16BufferAttribute?G=o.HALF_FLOAT:G=o.UNSIGNED_SHORT;else if($ instanceof Int16Array)G=o.SHORT;else if($ instanceof Uint32Array)G=o.UNSIGNED_INT;else if($ instanceof Int32Array)G=o.INT;else if($ instanceof Int8Array)G=o.BYTE;else if($ instanceof Uint8Array)G=o.UNSIGNED_BYTE;else if($ instanceof Uint8ClampedArray)G=o.UNSIGNED_BYTE;else throw new Error("THREE.WebGLAttributes: Unsupported buffer data format: "+$);return{buffer:F,type:G,bytesPerElement:$.BYTES_PER_ELEMENT,version:V.version,size:K}}function x(V,j,$){const ht=j.array,K=j.updateRanges;if(o.bindBuffer($,V),K.length===0)o.bufferSubData($,0,ht);else{K.sort((G,X)=>G.start-X.start);let F=0;for(let G=1;G<K.length;G++){const X=K[F],nt=K[G];nt.start<=X.start+X.count+1?X.count=Math.max(X.count,nt.start+nt.count-X.start):(++F,K[F]=nt)}K.length=F+1;for(let G=0,X=K.length;G<X;G++){const nt=K[G];o.bufferSubData($,nt.start*ht.BYTES_PER_ELEMENT,ht,nt.start,nt.count)}j.clearUpdateRanges()}j.onUploadCallback()}function E(V){return V.isInterleavedBufferAttribute&&(V=V.data),p.get(V)}function S(V){V.isInterleavedBufferAttribute&&(V=V.data);const j=p.get(V);j&&(o.deleteBuffer(j.buffer),p.delete(V))}function I(V,j){if(V.isInterleavedBufferAttribute&&(V=V.data),V.isGLBufferAttribute){const ht=p.get(V);(!ht||ht.version<V.version)&&p.set(V,{buffer:V.buffer,type:V.type,bytesPerElement:V.elementSize,version:V.version});return}const $=p.get(V);if($===void 0)p.set(V,h(V,j));else if($.version<V.version){if($.size!==V.array.byteLength)throw new Error("THREE.WebGLAttributes: The size of the buffer attribute's array buffer does not match the original size. Resizing buffer attributes is not supported.");x($.buffer,V,j),$.version=V.version}}return{get:E,remove:S,update:I}}var zo=`#ifdef USE_ALPHAHASH
	if ( diffuseColor.a < getAlphaHashThreshold( vPosition ) ) discard;
#endif`,Vo=`#ifdef USE_ALPHAHASH
	const float ALPHA_HASH_SCALE = 0.05;
	float hash2D( vec2 value ) {
		return fract( 1.0e4 * sin( 17.0 * value.x + 0.1 * value.y ) * ( 0.1 + abs( sin( 13.0 * value.y + value.x ) ) ) );
	}
	float hash3D( vec3 value ) {
		return hash2D( vec2( hash2D( value.xy ), value.z ) );
	}
	float getAlphaHashThreshold( vec3 position ) {
		float maxDeriv = max(
			length( dFdx( position.xyz ) ),
			length( dFdy( position.xyz ) )
		);
		float pixScale = 1.0 / ( ALPHA_HASH_SCALE * maxDeriv );
		vec2 pixScales = vec2(
			exp2( floor( log2( pixScale ) ) ),
			exp2( ceil( log2( pixScale ) ) )
		);
		vec2 alpha = vec2(
			hash3D( floor( pixScales.x * position.xyz ) ),
			hash3D( floor( pixScales.y * position.xyz ) )
		);
		float lerpFactor = fract( log2( pixScale ) );
		float x = ( 1.0 - lerpFactor ) * alpha.x + lerpFactor * alpha.y;
		float a = min( lerpFactor, 1.0 - lerpFactor );
		vec3 cases = vec3(
			x * x / ( 2.0 * a * ( 1.0 - a ) ),
			( x - 0.5 * a ) / ( 1.0 - a ),
			1.0 - ( ( 1.0 - x ) * ( 1.0 - x ) / ( 2.0 * a * ( 1.0 - a ) ) )
		);
		float threshold = ( x < ( 1.0 - a ) )
			? ( ( x < a ) ? cases.x : cases.y )
			: cases.z;
		return clamp( threshold , 1.0e-6, 1.0 );
	}
#endif`,Ah=`#ifdef USE_ALPHAMAP
	diffuseColor.a *= texture2D( alphaMap, vAlphaMapUv ).g;
#endif`,Eh=`#ifdef USE_ALPHAMAP
	uniform sampler2D alphaMap;
#endif`,ko=`#ifdef USE_ALPHATEST
	#ifdef ALPHA_TO_COVERAGE
	diffuseColor.a = smoothstep( alphaTest, alphaTest + fwidth( diffuseColor.a ), diffuseColor.a );
	if ( diffuseColor.a == 0.0 ) discard;
	#else
	if ( diffuseColor.a < alphaTest ) discard;
	#endif
#endif`,Go=`#ifdef USE_ALPHATEST
	uniform float alphaTest;
#endif`,Ho=`#ifdef USE_AOMAP
	float ambientOcclusion = ( texture2D( aoMap, vAoMapUv ).r - 1.0 ) * aoMapIntensity + 1.0;
	reflectedLight.indirectDiffuse *= ambientOcclusion;
	#if defined( USE_CLEARCOAT ) 
		clearcoatSpecularIndirect *= ambientOcclusion;
	#endif
	#if defined( USE_SHEEN ) 
		sheenSpecularIndirect *= ambientOcclusion;
	#endif
	#if defined( USE_ENVMAP ) && defined( STANDARD )
		float dotNV = saturate( dot( geometryNormal, geometryViewDir ) );
		reflectedLight.indirectSpecular *= computeSpecularOcclusion( dotNV, ambientOcclusion, material.roughness );
	#endif
#endif`,As=`#ifdef USE_AOMAP
	uniform sampler2D aoMap;
	uniform float aoMapIntensity;
#endif`,Yr=`#ifdef USE_BATCHING
	#if ! defined( GL_ANGLE_multi_draw )
	#define gl_DrawID _gl_DrawID
	uniform int _gl_DrawID;
	#endif
	uniform highp sampler2D batchingTexture;
	uniform highp usampler2D batchingIdTexture;
	mat4 getBatchingMatrix( const in float i ) {
		int size = textureSize( batchingTexture, 0 ).x;
		int j = int( i ) * 4;
		int x = j % size;
		int y = j / size;
		vec4 v1 = texelFetch( batchingTexture, ivec2( x, y ), 0 );
		vec4 v2 = texelFetch( batchingTexture, ivec2( x + 1, y ), 0 );
		vec4 v3 = texelFetch( batchingTexture, ivec2( x + 2, y ), 0 );
		vec4 v4 = texelFetch( batchingTexture, ivec2( x + 3, y ), 0 );
		return mat4( v1, v2, v3, v4 );
	}
	float getIndirectIndex( const in int i ) {
		int size = textureSize( batchingIdTexture, 0 ).x;
		int x = i % size;
		int y = i / size;
		return float( texelFetch( batchingIdTexture, ivec2( x, y ), 0 ).r );
	}
#endif
#ifdef USE_BATCHING_COLOR
	uniform sampler2D batchingColorTexture;
	vec4 getBatchingColor( const in float i ) {
		int size = textureSize( batchingColorTexture, 0 ).x;
		int j = int( i );
		int x = j % size;
		int y = j / size;
		return texelFetch( batchingColorTexture, ivec2( x, y ), 0 );
	}
#endif`,Wo=`#ifdef USE_BATCHING
	mat4 batchingMatrix = getBatchingMatrix( getIndirectIndex( gl_DrawID ) );
#endif`,Xo=`vec3 transformed = vec3( position );
#ifdef USE_ALPHAHASH
	vPosition = vec3( position );
#endif`,Ks=`vec3 objectNormal = vec3( normal );
#ifdef USE_TANGENT
	vec3 objectTangent = vec3( tangent.xyz );
#endif`,qo=`float G_BlinnPhong_Implicit( ) {
	return 0.25;
}
float D_BlinnPhong( const in float shininess, const in float dotNH ) {
	return RECIPROCAL_PI * ( shininess * 0.5 + 1.0 ) * pow( dotNH, shininess );
}
vec3 BRDF_BlinnPhong( const in vec3 lightDir, const in vec3 viewDir, const in vec3 normal, const in vec3 specularColor, const in float shininess ) {
	vec3 halfDir = normalize( lightDir + viewDir );
	float dotNH = saturate( dot( normal, halfDir ) );
	float dotVH = saturate( dot( viewDir, halfDir ) );
	vec3 F = F_Schlick( specularColor, 1.0, dotVH );
	float G = G_BlinnPhong_Implicit( );
	float D = D_BlinnPhong( shininess, dotNH );
	return F * ( G * D );
} // validated`,Yo=`#ifdef USE_IRIDESCENCE
	const mat3 XYZ_TO_REC709 = mat3(
		 3.2404542, -0.9692660,  0.0556434,
		-1.5371385,  1.8760108, -0.2040259,
		-0.4985314,  0.0415560,  1.0572252
	);
	vec3 Fresnel0ToIor( vec3 fresnel0 ) {
		vec3 sqrtF0 = sqrt( fresnel0 );
		return ( vec3( 1.0 ) + sqrtF0 ) / ( vec3( 1.0 ) - sqrtF0 );
	}
	vec3 IorToFresnel0( vec3 transmittedIor, float incidentIor ) {
		return pow2( ( transmittedIor - vec3( incidentIor ) ) / ( transmittedIor + vec3( incidentIor ) ) );
	}
	float IorToFresnel0( float transmittedIor, float incidentIor ) {
		return pow2( ( transmittedIor - incidentIor ) / ( transmittedIor + incidentIor ));
	}
	vec3 evalSensitivity( float OPD, vec3 shift ) {
		float phase = 2.0 * PI * OPD * 1.0e-9;
		vec3 val = vec3( 5.4856e-13, 4.4201e-13, 5.2481e-13 );
		vec3 pos = vec3( 1.6810e+06, 1.7953e+06, 2.2084e+06 );
		vec3 var = vec3( 4.3278e+09, 9.3046e+09, 6.6121e+09 );
		vec3 xyz = val * sqrt( 2.0 * PI * var ) * cos( pos * phase + shift ) * exp( - pow2( phase ) * var );
		xyz.x += 9.7470e-14 * sqrt( 2.0 * PI * 4.5282e+09 ) * cos( 2.2399e+06 * phase + shift[ 0 ] ) * exp( - 4.5282e+09 * pow2( phase ) );
		xyz /= 1.0685e-7;
		vec3 rgb = XYZ_TO_REC709 * xyz;
		return rgb;
	}
	vec3 evalIridescence( float outsideIOR, float eta2, float cosTheta1, float thinFilmThickness, vec3 baseF0 ) {
		vec3 I;
		float iridescenceIOR = mix( outsideIOR, eta2, smoothstep( 0.0, 0.03, thinFilmThickness ) );
		float sinTheta2Sq = pow2( outsideIOR / iridescenceIOR ) * ( 1.0 - pow2( cosTheta1 ) );
		float cosTheta2Sq = 1.0 - sinTheta2Sq;
		if ( cosTheta2Sq < 0.0 ) {
			return vec3( 1.0 );
		}
		float cosTheta2 = sqrt( cosTheta2Sq );
		float R0 = IorToFresnel0( iridescenceIOR, outsideIOR );
		float R12 = F_Schlick( R0, 1.0, cosTheta1 );
		float T121 = 1.0 - R12;
		float phi12 = 0.0;
		if ( iridescenceIOR < outsideIOR ) phi12 = PI;
		float phi21 = PI - phi12;
		vec3 baseIOR = Fresnel0ToIor( clamp( baseF0, 0.0, 0.9999 ) );		vec3 R1 = IorToFresnel0( baseIOR, iridescenceIOR );
		vec3 R23 = F_Schlick( R1, 1.0, cosTheta2 );
		vec3 phi23 = vec3( 0.0 );
		if ( baseIOR[ 0 ] < iridescenceIOR ) phi23[ 0 ] = PI;
		if ( baseIOR[ 1 ] < iridescenceIOR ) phi23[ 1 ] = PI;
		if ( baseIOR[ 2 ] < iridescenceIOR ) phi23[ 2 ] = PI;
		float OPD = 2.0 * iridescenceIOR * thinFilmThickness * cosTheta2;
		vec3 phi = vec3( phi21 ) + phi23;
		vec3 R123 = clamp( R12 * R23, 1e-5, 0.9999 );
		vec3 r123 = sqrt( R123 );
		vec3 Rs = pow2( T121 ) * R23 / ( vec3( 1.0 ) - R123 );
		vec3 C0 = R12 + Rs;
		I = C0;
		vec3 Cm = Rs - T121;
		for ( int m = 1; m <= 2; ++ m ) {
			Cm *= r123;
			vec3 Sm = 2.0 * evalSensitivity( float( m ) * OPD, float( m ) * phi );
			I += Cm * Sm;
		}
		return max( I, vec3( 0.0 ) );
	}
#endif`,Zo=`#ifdef USE_BUMPMAP
	uniform sampler2D bumpMap;
	uniform float bumpScale;
	vec2 dHdxy_fwd() {
		vec2 dSTdx = dFdx( vBumpMapUv );
		vec2 dSTdy = dFdy( vBumpMapUv );
		float Hll = bumpScale * texture2D( bumpMap, vBumpMapUv ).x;
		float dBx = bumpScale * texture2D( bumpMap, vBumpMapUv + dSTdx ).x - Hll;
		float dBy = bumpScale * texture2D( bumpMap, vBumpMapUv + dSTdy ).x - Hll;
		return vec2( dBx, dBy );
	}
	vec3 perturbNormalArb( vec3 surf_pos, vec3 surf_norm, vec2 dHdxy, float faceDirection ) {
		vec3 vSigmaX = normalize( dFdx( surf_pos.xyz ) );
		vec3 vSigmaY = normalize( dFdy( surf_pos.xyz ) );
		vec3 vN = surf_norm;
		vec3 R1 = cross( vSigmaY, vN );
		vec3 R2 = cross( vN, vSigmaX );
		float fDet = dot( vSigmaX, R1 ) * faceDirection;
		vec3 vGrad = sign( fDet ) * ( dHdxy.x * R1 + dHdxy.y * R2 );
		return normalize( abs( fDet ) * surf_norm - vGrad );
	}
#endif`,$o=`#if NUM_CLIPPING_PLANES > 0
	vec4 plane;
	#ifdef ALPHA_TO_COVERAGE
		float distanceToPlane, distanceGradient;
		float clipOpacity = 1.0;
		#pragma unroll_loop_start
		for ( int i = 0; i < UNION_CLIPPING_PLANES; i ++ ) {
			plane = clippingPlanes[ i ];
			distanceToPlane = - dot( vClipPosition, plane.xyz ) + plane.w;
			distanceGradient = fwidth( distanceToPlane ) / 2.0;
			clipOpacity *= smoothstep( - distanceGradient, distanceGradient, distanceToPlane );
			if ( clipOpacity == 0.0 ) discard;
		}
		#pragma unroll_loop_end
		#if UNION_CLIPPING_PLANES < NUM_CLIPPING_PLANES
			float unionClipOpacity = 1.0;
			#pragma unroll_loop_start
			for ( int i = UNION_CLIPPING_PLANES; i < NUM_CLIPPING_PLANES; i ++ ) {
				plane = clippingPlanes[ i ];
				distanceToPlane = - dot( vClipPosition, plane.xyz ) + plane.w;
				distanceGradient = fwidth( distanceToPlane ) / 2.0;
				unionClipOpacity *= 1.0 - smoothstep( - distanceGradient, distanceGradient, distanceToPlane );
			}
			#pragma unroll_loop_end
			clipOpacity *= 1.0 - unionClipOpacity;
		#endif
		diffuseColor.a *= clipOpacity;
		if ( diffuseColor.a == 0.0 ) discard;
	#else
		#pragma unroll_loop_start
		for ( int i = 0; i < UNION_CLIPPING_PLANES; i ++ ) {
			plane = clippingPlanes[ i ];
			if ( dot( vClipPosition, plane.xyz ) > plane.w ) discard;
		}
		#pragma unroll_loop_end
		#if UNION_CLIPPING_PLANES < NUM_CLIPPING_PLANES
			bool clipped = true;
			#pragma unroll_loop_start
			for ( int i = UNION_CLIPPING_PLANES; i < NUM_CLIPPING_PLANES; i ++ ) {
				plane = clippingPlanes[ i ];
				clipped = ( dot( vClipPosition, plane.xyz ) > plane.w ) && clipped;
			}
			#pragma unroll_loop_end
			if ( clipped ) discard;
		#endif
	#endif
#endif`,wh=`#if NUM_CLIPPING_PLANES > 0
	varying vec3 vClipPosition;
	uniform vec4 clippingPlanes[ NUM_CLIPPING_PLANES ];
#endif`,Qs=`#if NUM_CLIPPING_PLANES > 0
	varying vec3 vClipPosition;
#endif`,Jo=`#if NUM_CLIPPING_PLANES > 0
	vClipPosition = - mvPosition.xyz;
#endif`,Ko=`#if defined( USE_COLOR ) || defined( USE_COLOR_ALPHA )
	diffuseColor *= vColor;
#endif`,Qo=`#if defined( USE_COLOR ) || defined( USE_COLOR_ALPHA )
	varying vec4 vColor;
#endif`,jo=`#if defined( USE_COLOR ) || defined( USE_COLOR_ALPHA ) || defined( USE_INSTANCING_COLOR ) || defined( USE_BATCHING_COLOR )
	varying vec4 vColor;
#endif`,tl=`#if defined( USE_COLOR ) || defined( USE_COLOR_ALPHA ) || defined( USE_INSTANCING_COLOR ) || defined( USE_BATCHING_COLOR )
	vColor = vec4( 1.0 );
#endif
#ifdef USE_COLOR_ALPHA
	vColor *= color;
#elif defined( USE_COLOR )
	vColor.rgb *= color;
#endif
#ifdef USE_INSTANCING_COLOR
	vColor.rgb *= instanceColor.rgb;
#endif
#ifdef USE_BATCHING_COLOR
	vColor *= getBatchingColor( getIndirectIndex( gl_DrawID ) );
#endif`,el=`#define PI 3.141592653589793
#define PI2 6.283185307179586
#define PI_HALF 1.5707963267948966
#define RECIPROCAL_PI 0.3183098861837907
#define RECIPROCAL_PI2 0.15915494309189535
#define EPSILON 1e-6
#ifndef saturate
#define saturate( a ) clamp( a, 0.0, 1.0 )
#endif
#define whiteComplement( a ) ( 1.0 - saturate( a ) )
float pow2( const in float x ) { return x*x; }
vec3 pow2( const in vec3 x ) { return x*x; }
float pow3( const in float x ) { return x*x*x; }
float pow4( const in float x ) { float x2 = x*x; return x2*x2; }
float max3( const in vec3 v ) { return max( max( v.x, v.y ), v.z ); }
float average( const in vec3 v ) { return dot( v, vec3( 0.3333333 ) ); }
highp float rand( const in vec2 uv ) {
	const highp float a = 12.9898, b = 78.233, c = 43758.5453;
	highp float dt = dot( uv.xy, vec2( a,b ) ), sn = mod( dt, PI );
	return fract( sin( sn ) * c );
}
#ifdef HIGH_PRECISION
	float precisionSafeLength( vec3 v ) { return length( v ); }
#else
	float precisionSafeLength( vec3 v ) {
		float maxComponent = max3( abs( v ) );
		return length( v / maxComponent ) * maxComponent;
	}
#endif
struct IncidentLight {
	vec3 color;
	vec3 direction;
	bool visible;
};
struct ReflectedLight {
	vec3 directDiffuse;
	vec3 directSpecular;
	vec3 indirectDiffuse;
	vec3 indirectSpecular;
};
#ifdef USE_ALPHAHASH
	varying vec3 vPosition;
#endif
vec3 transformDirection( in vec3 dir, in mat4 matrix ) {
	return normalize( ( matrix * vec4( dir, 0.0 ) ).xyz );
}
vec3 inverseTransformDirection( in vec3 dir, in mat4 matrix ) {
	return normalize( ( vec4( dir, 0.0 ) * matrix ).xyz );
}
bool isPerspectiveMatrix( mat4 m ) {
	return m[ 2 ][ 3 ] == - 1.0;
}
vec2 equirectUv( in vec3 dir ) {
	float u = atan( dir.z, dir.x ) * RECIPROCAL_PI2 + 0.5;
	float v = asin( clamp( dir.y, - 1.0, 1.0 ) ) * RECIPROCAL_PI + 0.5;
	return vec2( u, v );
}
vec3 BRDF_Lambert( const in vec3 diffuseColor ) {
	return RECIPROCAL_PI * diffuseColor;
}
vec3 F_Schlick( const in vec3 f0, const in float f90, const in float dotVH ) {
	float fresnel = exp2( ( - 5.55473 * dotVH - 6.98316 ) * dotVH );
	return f0 * ( 1.0 - fresnel ) + ( f90 * fresnel );
}
float F_Schlick( const in float f0, const in float f90, const in float dotVH ) {
	float fresnel = exp2( ( - 5.55473 * dotVH - 6.98316 ) * dotVH );
	return f0 * ( 1.0 - fresnel ) + ( f90 * fresnel );
} // validated`,nl=`#ifdef ENVMAP_TYPE_CUBE_UV
	#define cubeUV_minMipLevel 4.0
	#define cubeUV_minTileSize 16.0
	float getFace( vec3 direction ) {
		vec3 absDirection = abs( direction );
		float face = - 1.0;
		if ( absDirection.x > absDirection.z ) {
			if ( absDirection.x > absDirection.y )
				face = direction.x > 0.0 ? 0.0 : 3.0;
			else
				face = direction.y > 0.0 ? 1.0 : 4.0;
		} else {
			if ( absDirection.z > absDirection.y )
				face = direction.z > 0.0 ? 2.0 : 5.0;
			else
				face = direction.y > 0.0 ? 1.0 : 4.0;
		}
		return face;
	}
	vec2 getUV( vec3 direction, float face ) {
		vec2 uv;
		if ( face == 0.0 ) {
			uv = vec2( direction.z, direction.y ) / abs( direction.x );
		} else if ( face == 1.0 ) {
			uv = vec2( - direction.x, - direction.z ) / abs( direction.y );
		} else if ( face == 2.0 ) {
			uv = vec2( - direction.x, direction.y ) / abs( direction.z );
		} else if ( face == 3.0 ) {
			uv = vec2( - direction.z, direction.y ) / abs( direction.x );
		} else if ( face == 4.0 ) {
			uv = vec2( - direction.x, direction.z ) / abs( direction.y );
		} else {
			uv = vec2( direction.x, direction.y ) / abs( direction.z );
		}
		return 0.5 * ( uv + 1.0 );
	}
	vec3 bilinearCubeUV( sampler2D envMap, vec3 direction, float mipInt ) {
		float face = getFace( direction );
		float filterInt = max( cubeUV_minMipLevel - mipInt, 0.0 );
		mipInt = max( mipInt, cubeUV_minMipLevel );
		float faceSize = exp2( mipInt );
		highp vec2 uv = getUV( direction, face ) * ( faceSize - 2.0 ) + 1.0;
		if ( face > 2.0 ) {
			uv.y += faceSize;
			face -= 3.0;
		}
		uv.x += face * faceSize;
		uv.x += filterInt * 3.0 * cubeUV_minTileSize;
		uv.y += 4.0 * ( exp2( CUBEUV_MAX_MIP ) - faceSize );
		uv.x *= CUBEUV_TEXEL_WIDTH;
		uv.y *= CUBEUV_TEXEL_HEIGHT;
		#ifdef texture2DGradEXT
			return texture2DGradEXT( envMap, uv, vec2( 0.0 ), vec2( 0.0 ) ).rgb;
		#else
			return texture2D( envMap, uv ).rgb;
		#endif
	}
	#define cubeUV_r0 1.0
	#define cubeUV_m0 - 2.0
	#define cubeUV_r1 0.8
	#define cubeUV_m1 - 1.0
	#define cubeUV_r4 0.4
	#define cubeUV_m4 2.0
	#define cubeUV_r5 0.305
	#define cubeUV_m5 3.0
	#define cubeUV_r6 0.21
	#define cubeUV_m6 4.0
	float roughnessToMip( float roughness ) {
		float mip = 0.0;
		if ( roughness >= cubeUV_r1 ) {
			mip = ( cubeUV_r0 - roughness ) * ( cubeUV_m1 - cubeUV_m0 ) / ( cubeUV_r0 - cubeUV_r1 ) + cubeUV_m0;
		} else if ( roughness >= cubeUV_r4 ) {
			mip = ( cubeUV_r1 - roughness ) * ( cubeUV_m4 - cubeUV_m1 ) / ( cubeUV_r1 - cubeUV_r4 ) + cubeUV_m1;
		} else if ( roughness >= cubeUV_r5 ) {
			mip = ( cubeUV_r4 - roughness ) * ( cubeUV_m5 - cubeUV_m4 ) / ( cubeUV_r4 - cubeUV_r5 ) + cubeUV_m4;
		} else if ( roughness >= cubeUV_r6 ) {
			mip = ( cubeUV_r5 - roughness ) * ( cubeUV_m6 - cubeUV_m5 ) / ( cubeUV_r5 - cubeUV_r6 ) + cubeUV_m5;
		} else {
			mip = - 2.0 * log2( 1.16 * roughness );		}
		return mip;
	}
	vec4 textureCubeUV( sampler2D envMap, vec3 sampleDir, float roughness ) {
		float mip = clamp( roughnessToMip( roughness ), cubeUV_m0, CUBEUV_MAX_MIP );
		float mipF = fract( mip );
		float mipInt = floor( mip );
		vec3 color0 = bilinearCubeUV( envMap, sampleDir, mipInt );
		if ( mipF == 0.0 ) {
			return vec4( color0, 1.0 );
		} else {
			vec3 color1 = bilinearCubeUV( envMap, sampleDir, mipInt + 1.0 );
			return vec4( mix( color0, color1, mipF ), 1.0 );
		}
	}
#endif`,il=`vec3 transformedNormal = objectNormal;
#ifdef USE_TANGENT
	vec3 transformedTangent = objectTangent;
#endif
#ifdef USE_BATCHING
	mat3 bm = mat3( batchingMatrix );
	transformedNormal /= vec3( dot( bm[ 0 ], bm[ 0 ] ), dot( bm[ 1 ], bm[ 1 ] ), dot( bm[ 2 ], bm[ 2 ] ) );
	transformedNormal = bm * transformedNormal;
	#ifdef USE_TANGENT
		transformedTangent = bm * transformedTangent;
	#endif
#endif
#ifdef USE_INSTANCING
	mat3 im = mat3( instanceMatrix );
	transformedNormal /= vec3( dot( im[ 0 ], im[ 0 ] ), dot( im[ 1 ], im[ 1 ] ), dot( im[ 2 ], im[ 2 ] ) );
	transformedNormal = im * transformedNormal;
	#ifdef USE_TANGENT
		transformedTangent = im * transformedTangent;
	#endif
#endif
transformedNormal = normalMatrix * transformedNormal;
#ifdef FLIP_SIDED
	transformedNormal = - transformedNormal;
#endif
#ifdef USE_TANGENT
	transformedTangent = ( modelViewMatrix * vec4( transformedTangent, 0.0 ) ).xyz;
	#ifdef FLIP_SIDED
		transformedTangent = - transformedTangent;
	#endif
#endif`,js=`#ifdef USE_DISPLACEMENTMAP
	uniform sampler2D displacementMap;
	uniform float displacementScale;
	uniform float displacementBias;
#endif`,tr=`#ifdef USE_DISPLACEMENTMAP
	transformed += normalize( objectNormal ) * ( texture2D( displacementMap, vDisplacementMapUv ).x * displacementScale + displacementBias );
#endif`,sl=`#ifdef USE_EMISSIVEMAP
	vec4 emissiveColor = texture2D( emissiveMap, vEmissiveMapUv );
	#ifdef DECODE_VIDEO_TEXTURE_EMISSIVE
		emissiveColor = sRGBTransferEOTF( emissiveColor );
	#endif
	totalEmissiveRadiance *= emissiveColor.rgb;
#endif`,rl=`#ifdef USE_EMISSIVEMAP
	uniform sampler2D emissiveMap;
#endif`,al="gl_FragColor = linearToOutputTexel( gl_FragColor );",ol=`vec4 LinearTransferOETF( in vec4 value ) {
	return value;
}
vec4 sRGBTransferEOTF( in vec4 value ) {
	return vec4( mix( pow( value.rgb * 0.9478672986 + vec3( 0.0521327014 ), vec3( 2.4 ) ), value.rgb * 0.0773993808, vec3( lessThanEqual( value.rgb, vec3( 0.04045 ) ) ) ), value.a );
}
vec4 sRGBTransferOETF( in vec4 value ) {
	return vec4( mix( pow( value.rgb, vec3( 0.41666 ) ) * 1.055 - vec3( 0.055 ), value.rgb * 12.92, vec3( lessThanEqual( value.rgb, vec3( 0.0031308 ) ) ) ), value.a );
}`,ll=`#ifdef USE_ENVMAP
	#ifdef ENV_WORLDPOS
		vec3 cameraToFrag;
		if ( isOrthographic ) {
			cameraToFrag = normalize( vec3( - viewMatrix[ 0 ][ 2 ], - viewMatrix[ 1 ][ 2 ], - viewMatrix[ 2 ][ 2 ] ) );
		} else {
			cameraToFrag = normalize( vWorldPosition - cameraPosition );
		}
		vec3 worldNormal = inverseTransformDirection( normal, viewMatrix );
		#ifdef ENVMAP_MODE_REFLECTION
			vec3 reflectVec = reflect( cameraToFrag, worldNormal );
		#else
			vec3 reflectVec = refract( cameraToFrag, worldNormal, refractionRatio );
		#endif
	#else
		vec3 reflectVec = vReflect;
	#endif
	#ifdef ENVMAP_TYPE_CUBE
		vec4 envColor = textureCube( envMap, envMapRotation * vec3( flipEnvMap * reflectVec.x, reflectVec.yz ) );
		#ifdef ENVMAP_BLENDING_MULTIPLY
			outgoingLight = mix( outgoingLight, outgoingLight * envColor.xyz, specularStrength * reflectivity );
		#elif defined( ENVMAP_BLENDING_MIX )
			outgoingLight = mix( outgoingLight, envColor.xyz, specularStrength * reflectivity );
		#elif defined( ENVMAP_BLENDING_ADD )
			outgoingLight += envColor.xyz * specularStrength * reflectivity;
		#endif
	#endif
#endif`,cl=`#ifdef USE_ENVMAP
	uniform float envMapIntensity;
	uniform float flipEnvMap;
	uniform mat3 envMapRotation;
	#ifdef ENVMAP_TYPE_CUBE
		uniform samplerCube envMap;
	#else
		uniform sampler2D envMap;
	#endif
#endif`,hl=`#ifdef USE_ENVMAP
	uniform float reflectivity;
	#if defined( USE_BUMPMAP ) || defined( USE_NORMALMAP ) || defined( PHONG ) || defined( LAMBERT )
		#define ENV_WORLDPOS
	#endif
	#ifdef ENV_WORLDPOS
		varying vec3 vWorldPosition;
		uniform float refractionRatio;
	#else
		varying vec3 vReflect;
	#endif
#endif`,ul=`#ifdef USE_ENVMAP
	#if defined( USE_BUMPMAP ) || defined( USE_NORMALMAP ) || defined( PHONG ) || defined( LAMBERT )
		#define ENV_WORLDPOS
	#endif
	#ifdef ENV_WORLDPOS
		
		varying vec3 vWorldPosition;
	#else
		varying vec3 vReflect;
		uniform float refractionRatio;
	#endif
#endif`,fl=`#ifdef USE_ENVMAP
	#ifdef ENV_WORLDPOS
		vWorldPosition = worldPosition.xyz;
	#else
		vec3 cameraToVertex;
		if ( isOrthographic ) {
			cameraToVertex = normalize( vec3( - viewMatrix[ 0 ][ 2 ], - viewMatrix[ 1 ][ 2 ], - viewMatrix[ 2 ][ 2 ] ) );
		} else {
			cameraToVertex = normalize( worldPosition.xyz - cameraPosition );
		}
		vec3 worldNormal = inverseTransformDirection( transformedNormal, viewMatrix );
		#ifdef ENVMAP_MODE_REFLECTION
			vReflect = reflect( cameraToVertex, worldNormal );
		#else
			vReflect = refract( cameraToVertex, worldNormal, refractionRatio );
		#endif
	#endif
#endif`,er=`#ifdef USE_FOG
	vFogDepth = - mvPosition.z;
#endif`,nr=`#ifdef USE_FOG
	varying float vFogDepth;
#endif`,ir=`#ifdef USE_FOG
	#ifdef FOG_EXP2
		float fogFactor = 1.0 - exp( - fogDensity * fogDensity * vFogDepth * vFogDepth );
	#else
		float fogFactor = smoothstep( fogNear, fogFar, vFogDepth );
	#endif
	gl_FragColor.rgb = mix( gl_FragColor.rgb, fogColor, fogFactor );
#endif`,qi=`#ifdef USE_FOG
	uniform vec3 fogColor;
	varying float vFogDepth;
	#ifdef FOG_EXP2
		uniform float fogDensity;
	#else
		uniform float fogNear;
		uniform float fogFar;
	#endif
#endif`,sr=`#ifdef USE_GRADIENTMAP
	uniform sampler2D gradientMap;
#endif
vec3 getGradientIrradiance( vec3 normal, vec3 lightDirection ) {
	float dotNL = dot( normal, lightDirection );
	vec2 coord = vec2( dotNL * 0.5 + 0.5, 0.0 );
	#ifdef USE_GRADIENTMAP
		return vec3( texture2D( gradientMap, coord ).r );
	#else
		vec2 fw = fwidth( coord ) * 0.5;
		return mix( vec3( 0.7 ), vec3( 1.0 ), smoothstep( 0.7 - fw.x, 0.7 + fw.x, coord.x ) );
	#endif
}`,rr=`#ifdef USE_LIGHTMAP
	uniform sampler2D lightMap;
	uniform float lightMapIntensity;
#endif`,ar=`LambertMaterial material;
material.diffuseColor = diffuseColor.rgb;
material.specularStrength = specularStrength;`,or=`varying vec3 vViewPosition;
struct LambertMaterial {
	vec3 diffuseColor;
	float specularStrength;
};
void RE_Direct_Lambert( const in IncidentLight directLight, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in LambertMaterial material, inout ReflectedLight reflectedLight ) {
	float dotNL = saturate( dot( geometryNormal, directLight.direction ) );
	vec3 irradiance = dotNL * directLight.color;
	reflectedLight.directDiffuse += irradiance * BRDF_Lambert( material.diffuseColor );
}
void RE_IndirectDiffuse_Lambert( const in vec3 irradiance, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in LambertMaterial material, inout ReflectedLight reflectedLight ) {
	reflectedLight.indirectDiffuse += irradiance * BRDF_Lambert( material.diffuseColor );
}
#define RE_Direct				RE_Direct_Lambert
#define RE_IndirectDiffuse		RE_IndirectDiffuse_Lambert`,Es=`uniform bool receiveShadow;
uniform vec3 ambientLightColor;
#if defined( USE_LIGHT_PROBES )
	uniform vec3 lightProbe[ 9 ];
#endif
vec3 shGetIrradianceAt( in vec3 normal, in vec3 shCoefficients[ 9 ] ) {
	float x = normal.x, y = normal.y, z = normal.z;
	vec3 result = shCoefficients[ 0 ] * 0.886227;
	result += shCoefficients[ 1 ] * 2.0 * 0.511664 * y;
	result += shCoefficients[ 2 ] * 2.0 * 0.511664 * z;
	result += shCoefficients[ 3 ] * 2.0 * 0.511664 * x;
	result += shCoefficients[ 4 ] * 2.0 * 0.429043 * x * y;
	result += shCoefficients[ 5 ] * 2.0 * 0.429043 * y * z;
	result += shCoefficients[ 6 ] * ( 0.743125 * z * z - 0.247708 );
	result += shCoefficients[ 7 ] * 2.0 * 0.429043 * x * z;
	result += shCoefficients[ 8 ] * 0.429043 * ( x * x - y * y );
	return result;
}
vec3 getLightProbeIrradiance( const in vec3 lightProbe[ 9 ], const in vec3 normal ) {
	vec3 worldNormal = inverseTransformDirection( normal, viewMatrix );
	vec3 irradiance = shGetIrradianceAt( worldNormal, lightProbe );
	return irradiance;
}
vec3 getAmbientLightIrradiance( const in vec3 ambientLightColor ) {
	vec3 irradiance = ambientLightColor;
	return irradiance;
}
float getDistanceAttenuation( const in float lightDistance, const in float cutoffDistance, const in float decayExponent ) {
	float distanceFalloff = 1.0 / max( pow( lightDistance, decayExponent ), 0.01 );
	if ( cutoffDistance > 0.0 ) {
		distanceFalloff *= pow2( saturate( 1.0 - pow4( lightDistance / cutoffDistance ) ) );
	}
	return distanceFalloff;
}
float getSpotAttenuation( const in float coneCosine, const in float penumbraCosine, const in float angleCosine ) {
	return smoothstep( coneCosine, penumbraCosine, angleCosine );
}
#if NUM_DIR_LIGHTS > 0
	struct DirectionalLight {
		vec3 direction;
		vec3 color;
	};
	uniform DirectionalLight directionalLights[ NUM_DIR_LIGHTS ];
	void getDirectionalLightInfo( const in DirectionalLight directionalLight, out IncidentLight light ) {
		light.color = directionalLight.color;
		light.direction = directionalLight.direction;
		light.visible = true;
	}
#endif
#if NUM_POINT_LIGHTS > 0
	struct PointLight {
		vec3 position;
		vec3 color;
		float distance;
		float decay;
	};
	uniform PointLight pointLights[ NUM_POINT_LIGHTS ];
	void getPointLightInfo( const in PointLight pointLight, const in vec3 geometryPosition, out IncidentLight light ) {
		vec3 lVector = pointLight.position - geometryPosition;
		light.direction = normalize( lVector );
		float lightDistance = length( lVector );
		light.color = pointLight.color;
		light.color *= getDistanceAttenuation( lightDistance, pointLight.distance, pointLight.decay );
		light.visible = ( light.color != vec3( 0.0 ) );
	}
#endif
#if NUM_SPOT_LIGHTS > 0
	struct SpotLight {
		vec3 position;
		vec3 direction;
		vec3 color;
		float distance;
		float decay;
		float coneCos;
		float penumbraCos;
	};
	uniform SpotLight spotLights[ NUM_SPOT_LIGHTS ];
	void getSpotLightInfo( const in SpotLight spotLight, const in vec3 geometryPosition, out IncidentLight light ) {
		vec3 lVector = spotLight.position - geometryPosition;
		light.direction = normalize( lVector );
		float angleCos = dot( light.direction, spotLight.direction );
		float spotAttenuation = getSpotAttenuation( spotLight.coneCos, spotLight.penumbraCos, angleCos );
		if ( spotAttenuation > 0.0 ) {
			float lightDistance = length( lVector );
			light.color = spotLight.color * spotAttenuation;
			light.color *= getDistanceAttenuation( lightDistance, spotLight.distance, spotLight.decay );
			light.visible = ( light.color != vec3( 0.0 ) );
		} else {
			light.color = vec3( 0.0 );
			light.visible = false;
		}
	}
#endif
#if NUM_RECT_AREA_LIGHTS > 0
	struct RectAreaLight {
		vec3 color;
		vec3 position;
		vec3 halfWidth;
		vec3 halfHeight;
	};
	uniform sampler2D ltc_1;	uniform sampler2D ltc_2;
	uniform RectAreaLight rectAreaLights[ NUM_RECT_AREA_LIGHTS ];
#endif
#if NUM_HEMI_LIGHTS > 0
	struct HemisphereLight {
		vec3 direction;
		vec3 skyColor;
		vec3 groundColor;
	};
	uniform HemisphereLight hemisphereLights[ NUM_HEMI_LIGHTS ];
	vec3 getHemisphereLightIrradiance( const in HemisphereLight hemiLight, const in vec3 normal ) {
		float dotNL = dot( normal, hemiLight.direction );
		float hemiDiffuseWeight = 0.5 * dotNL + 0.5;
		vec3 irradiance = mix( hemiLight.groundColor, hemiLight.skyColor, hemiDiffuseWeight );
		return irradiance;
	}
#endif`,dl=`#ifdef USE_ENVMAP
	vec3 getIBLIrradiance( const in vec3 normal ) {
		#ifdef ENVMAP_TYPE_CUBE_UV
			vec3 worldNormal = inverseTransformDirection( normal, viewMatrix );
			vec4 envMapColor = textureCubeUV( envMap, envMapRotation * worldNormal, 1.0 );
			return PI * envMapColor.rgb * envMapIntensity;
		#else
			return vec3( 0.0 );
		#endif
	}
	vec3 getIBLRadiance( const in vec3 viewDir, const in vec3 normal, const in float roughness ) {
		#ifdef ENVMAP_TYPE_CUBE_UV
			vec3 reflectVec = reflect( - viewDir, normal );
			reflectVec = normalize( mix( reflectVec, normal, pow4( roughness ) ) );
			reflectVec = inverseTransformDirection( reflectVec, viewMatrix );
			vec4 envMapColor = textureCubeUV( envMap, envMapRotation * reflectVec, roughness );
			return envMapColor.rgb * envMapIntensity;
		#else
			return vec3( 0.0 );
		#endif
	}
	#ifdef USE_ANISOTROPY
		vec3 getIBLAnisotropyRadiance( const in vec3 viewDir, const in vec3 normal, const in float roughness, const in vec3 bitangent, const in float anisotropy ) {
			#ifdef ENVMAP_TYPE_CUBE_UV
				vec3 bentNormal = cross( bitangent, viewDir );
				bentNormal = normalize( cross( bentNormal, bitangent ) );
				bentNormal = normalize( mix( bentNormal, normal, pow2( pow2( 1.0 - anisotropy * ( 1.0 - roughness ) ) ) ) );
				return getIBLRadiance( viewDir, bentNormal, roughness );
			#else
				return vec3( 0.0 );
			#endif
		}
	#endif
#endif`,pl=`ToonMaterial material;
material.diffuseColor = diffuseColor.rgb;`,ml=`varying vec3 vViewPosition;
struct ToonMaterial {
	vec3 diffuseColor;
};
void RE_Direct_Toon( const in IncidentLight directLight, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in ToonMaterial material, inout ReflectedLight reflectedLight ) {
	vec3 irradiance = getGradientIrradiance( geometryNormal, directLight.direction ) * directLight.color;
	reflectedLight.directDiffuse += irradiance * BRDF_Lambert( material.diffuseColor );
}
void RE_IndirectDiffuse_Toon( const in vec3 irradiance, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in ToonMaterial material, inout ReflectedLight reflectedLight ) {
	reflectedLight.indirectDiffuse += irradiance * BRDF_Lambert( material.diffuseColor );
}
#define RE_Direct				RE_Direct_Toon
#define RE_IndirectDiffuse		RE_IndirectDiffuse_Toon`,gl=`BlinnPhongMaterial material;
material.diffuseColor = diffuseColor.rgb;
material.specularColor = specular;
material.specularShininess = shininess;
material.specularStrength = specularStrength;`,_l=`varying vec3 vViewPosition;
struct BlinnPhongMaterial {
	vec3 diffuseColor;
	vec3 specularColor;
	float specularShininess;
	float specularStrength;
};
void RE_Direct_BlinnPhong( const in IncidentLight directLight, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in BlinnPhongMaterial material, inout ReflectedLight reflectedLight ) {
	float dotNL = saturate( dot( geometryNormal, directLight.direction ) );
	vec3 irradiance = dotNL * directLight.color;
	reflectedLight.directDiffuse += irradiance * BRDF_Lambert( material.diffuseColor );
	reflectedLight.directSpecular += irradiance * BRDF_BlinnPhong( directLight.direction, geometryViewDir, geometryNormal, material.specularColor, material.specularShininess ) * material.specularStrength;
}
void RE_IndirectDiffuse_BlinnPhong( const in vec3 irradiance, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in BlinnPhongMaterial material, inout ReflectedLight reflectedLight ) {
	reflectedLight.indirectDiffuse += irradiance * BRDF_Lambert( material.diffuseColor );
}
#define RE_Direct				RE_Direct_BlinnPhong
#define RE_IndirectDiffuse		RE_IndirectDiffuse_BlinnPhong`,xl=`PhysicalMaterial material;
material.diffuseColor = diffuseColor.rgb;
material.diffuseContribution = diffuseColor.rgb * ( 1.0 - metalnessFactor );
material.metalness = metalnessFactor;
vec3 dxy = max( abs( dFdx( nonPerturbedNormal ) ), abs( dFdy( nonPerturbedNormal ) ) );
float geometryRoughness = max( max( dxy.x, dxy.y ), dxy.z );
material.roughness = max( roughnessFactor, 0.0525 );material.roughness += geometryRoughness;
material.roughness = min( material.roughness, 1.0 );
#ifdef IOR
	material.ior = ior;
	#ifdef USE_SPECULAR
		float specularIntensityFactor = specularIntensity;
		vec3 specularColorFactor = specularColor;
		#ifdef USE_SPECULAR_COLORMAP
			specularColorFactor *= texture2D( specularColorMap, vSpecularColorMapUv ).rgb;
		#endif
		#ifdef USE_SPECULAR_INTENSITYMAP
			specularIntensityFactor *= texture2D( specularIntensityMap, vSpecularIntensityMapUv ).a;
		#endif
		material.specularF90 = mix( specularIntensityFactor, 1.0, metalnessFactor );
	#else
		float specularIntensityFactor = 1.0;
		vec3 specularColorFactor = vec3( 1.0 );
		material.specularF90 = 1.0;
	#endif
	material.specularColor = min( pow2( ( material.ior - 1.0 ) / ( material.ior + 1.0 ) ) * specularColorFactor, vec3( 1.0 ) ) * specularIntensityFactor;
	material.specularColorBlended = mix( material.specularColor, diffuseColor.rgb, metalnessFactor );
#else
	material.specularColor = vec3( 0.04 );
	material.specularColorBlended = mix( material.specularColor, diffuseColor.rgb, metalnessFactor );
	material.specularF90 = 1.0;
#endif
#ifdef USE_CLEARCOAT
	material.clearcoat = clearcoat;
	material.clearcoatRoughness = clearcoatRoughness;
	material.clearcoatF0 = vec3( 0.04 );
	material.clearcoatF90 = 1.0;
	#ifdef USE_CLEARCOATMAP
		material.clearcoat *= texture2D( clearcoatMap, vClearcoatMapUv ).x;
	#endif
	#ifdef USE_CLEARCOAT_ROUGHNESSMAP
		material.clearcoatRoughness *= texture2D( clearcoatRoughnessMap, vClearcoatRoughnessMapUv ).y;
	#endif
	material.clearcoat = saturate( material.clearcoat );	material.clearcoatRoughness = max( material.clearcoatRoughness, 0.0525 );
	material.clearcoatRoughness += geometryRoughness;
	material.clearcoatRoughness = min( material.clearcoatRoughness, 1.0 );
#endif
#ifdef USE_DISPERSION
	material.dispersion = dispersion;
#endif
#ifdef USE_IRIDESCENCE
	material.iridescence = iridescence;
	material.iridescenceIOR = iridescenceIOR;
	#ifdef USE_IRIDESCENCEMAP
		material.iridescence *= texture2D( iridescenceMap, vIridescenceMapUv ).r;
	#endif
	#ifdef USE_IRIDESCENCE_THICKNESSMAP
		material.iridescenceThickness = (iridescenceThicknessMaximum - iridescenceThicknessMinimum) * texture2D( iridescenceThicknessMap, vIridescenceThicknessMapUv ).g + iridescenceThicknessMinimum;
	#else
		material.iridescenceThickness = iridescenceThicknessMaximum;
	#endif
#endif
#ifdef USE_SHEEN
	material.sheenColor = sheenColor;
	#ifdef USE_SHEEN_COLORMAP
		material.sheenColor *= texture2D( sheenColorMap, vSheenColorMapUv ).rgb;
	#endif
	material.sheenRoughness = clamp( sheenRoughness, 0.0001, 1.0 );
	#ifdef USE_SHEEN_ROUGHNESSMAP
		material.sheenRoughness *= texture2D( sheenRoughnessMap, vSheenRoughnessMapUv ).a;
	#endif
#endif
#ifdef USE_ANISOTROPY
	#ifdef USE_ANISOTROPYMAP
		mat2 anisotropyMat = mat2( anisotropyVector.x, anisotropyVector.y, - anisotropyVector.y, anisotropyVector.x );
		vec3 anisotropyPolar = texture2D( anisotropyMap, vAnisotropyMapUv ).rgb;
		vec2 anisotropyV = anisotropyMat * normalize( 2.0 * anisotropyPolar.rg - vec2( 1.0 ) ) * anisotropyPolar.b;
	#else
		vec2 anisotropyV = anisotropyVector;
	#endif
	material.anisotropy = length( anisotropyV );
	if( material.anisotropy == 0.0 ) {
		anisotropyV = vec2( 1.0, 0.0 );
	} else {
		anisotropyV /= material.anisotropy;
		material.anisotropy = saturate( material.anisotropy );
	}
	material.alphaT = mix( pow2( material.roughness ), 1.0, pow2( material.anisotropy ) );
	material.anisotropyT = tbn[ 0 ] * anisotropyV.x + tbn[ 1 ] * anisotropyV.y;
	material.anisotropyB = tbn[ 1 ] * anisotropyV.x - tbn[ 0 ] * anisotropyV.y;
#endif`,vl=`uniform sampler2D dfgLUT;
struct PhysicalMaterial {
	vec3 diffuseColor;
	vec3 diffuseContribution;
	vec3 specularColor;
	vec3 specularColorBlended;
	float roughness;
	float metalness;
	float specularF90;
	float dispersion;
	#ifdef USE_CLEARCOAT
		float clearcoat;
		float clearcoatRoughness;
		vec3 clearcoatF0;
		float clearcoatF90;
	#endif
	#ifdef USE_IRIDESCENCE
		float iridescence;
		float iridescenceIOR;
		float iridescenceThickness;
		vec3 iridescenceFresnel;
		vec3 iridescenceF0;
		vec3 iridescenceFresnelDielectric;
		vec3 iridescenceFresnelMetallic;
	#endif
	#ifdef USE_SHEEN
		vec3 sheenColor;
		float sheenRoughness;
	#endif
	#ifdef IOR
		float ior;
	#endif
	#ifdef USE_TRANSMISSION
		float transmission;
		float transmissionAlpha;
		float thickness;
		float attenuationDistance;
		vec3 attenuationColor;
	#endif
	#ifdef USE_ANISOTROPY
		float anisotropy;
		float alphaT;
		vec3 anisotropyT;
		vec3 anisotropyB;
	#endif
};
vec3 clearcoatSpecularDirect = vec3( 0.0 );
vec3 clearcoatSpecularIndirect = vec3( 0.0 );
vec3 sheenSpecularDirect = vec3( 0.0 );
vec3 sheenSpecularIndirect = vec3(0.0 );
vec3 Schlick_to_F0( const in vec3 f, const in float f90, const in float dotVH ) {
    float x = clamp( 1.0 - dotVH, 0.0, 1.0 );
    float x2 = x * x;
    float x5 = clamp( x * x2 * x2, 0.0, 0.9999 );
    return ( f - vec3( f90 ) * x5 ) / ( 1.0 - x5 );
}
float V_GGX_SmithCorrelated( const in float alpha, const in float dotNL, const in float dotNV ) {
	float a2 = pow2( alpha );
	float gv = dotNL * sqrt( a2 + ( 1.0 - a2 ) * pow2( dotNV ) );
	float gl = dotNV * sqrt( a2 + ( 1.0 - a2 ) * pow2( dotNL ) );
	return 0.5 / max( gv + gl, EPSILON );
}
float D_GGX( const in float alpha, const in float dotNH ) {
	float a2 = pow2( alpha );
	float denom = pow2( dotNH ) * ( a2 - 1.0 ) + 1.0;
	return RECIPROCAL_PI * a2 / pow2( denom );
}
#ifdef USE_ANISOTROPY
	float V_GGX_SmithCorrelated_Anisotropic( const in float alphaT, const in float alphaB, const in float dotTV, const in float dotBV, const in float dotTL, const in float dotBL, const in float dotNV, const in float dotNL ) {
		float gv = dotNL * length( vec3( alphaT * dotTV, alphaB * dotBV, dotNV ) );
		float gl = dotNV * length( vec3( alphaT * dotTL, alphaB * dotBL, dotNL ) );
		float v = 0.5 / ( gv + gl );
		return v;
	}
	float D_GGX_Anisotropic( const in float alphaT, const in float alphaB, const in float dotNH, const in float dotTH, const in float dotBH ) {
		float a2 = alphaT * alphaB;
		highp vec3 v = vec3( alphaB * dotTH, alphaT * dotBH, a2 * dotNH );
		highp float v2 = dot( v, v );
		float w2 = a2 / v2;
		return RECIPROCAL_PI * a2 * pow2 ( w2 );
	}
#endif
#ifdef USE_CLEARCOAT
	vec3 BRDF_GGX_Clearcoat( const in vec3 lightDir, const in vec3 viewDir, const in vec3 normal, const in PhysicalMaterial material) {
		vec3 f0 = material.clearcoatF0;
		float f90 = material.clearcoatF90;
		float roughness = material.clearcoatRoughness;
		float alpha = pow2( roughness );
		vec3 halfDir = normalize( lightDir + viewDir );
		float dotNL = saturate( dot( normal, lightDir ) );
		float dotNV = saturate( dot( normal, viewDir ) );
		float dotNH = saturate( dot( normal, halfDir ) );
		float dotVH = saturate( dot( viewDir, halfDir ) );
		vec3 F = F_Schlick( f0, f90, dotVH );
		float V = V_GGX_SmithCorrelated( alpha, dotNL, dotNV );
		float D = D_GGX( alpha, dotNH );
		return F * ( V * D );
	}
#endif
vec3 BRDF_GGX( const in vec3 lightDir, const in vec3 viewDir, const in vec3 normal, const in PhysicalMaterial material ) {
	vec3 f0 = material.specularColorBlended;
	float f90 = material.specularF90;
	float roughness = material.roughness;
	float alpha = pow2( roughness );
	vec3 halfDir = normalize( lightDir + viewDir );
	float dotNL = saturate( dot( normal, lightDir ) );
	float dotNV = saturate( dot( normal, viewDir ) );
	float dotNH = saturate( dot( normal, halfDir ) );
	float dotVH = saturate( dot( viewDir, halfDir ) );
	vec3 F = F_Schlick( f0, f90, dotVH );
	#ifdef USE_IRIDESCENCE
		F = mix( F, material.iridescenceFresnel, material.iridescence );
	#endif
	#ifdef USE_ANISOTROPY
		float dotTL = dot( material.anisotropyT, lightDir );
		float dotTV = dot( material.anisotropyT, viewDir );
		float dotTH = dot( material.anisotropyT, halfDir );
		float dotBL = dot( material.anisotropyB, lightDir );
		float dotBV = dot( material.anisotropyB, viewDir );
		float dotBH = dot( material.anisotropyB, halfDir );
		float V = V_GGX_SmithCorrelated_Anisotropic( material.alphaT, alpha, dotTV, dotBV, dotTL, dotBL, dotNV, dotNL );
		float D = D_GGX_Anisotropic( material.alphaT, alpha, dotNH, dotTH, dotBH );
	#else
		float V = V_GGX_SmithCorrelated( alpha, dotNL, dotNV );
		float D = D_GGX( alpha, dotNH );
	#endif
	return F * ( V * D );
}
vec2 LTC_Uv( const in vec3 N, const in vec3 V, const in float roughness ) {
	const float LUT_SIZE = 64.0;
	const float LUT_SCALE = ( LUT_SIZE - 1.0 ) / LUT_SIZE;
	const float LUT_BIAS = 0.5 / LUT_SIZE;
	float dotNV = saturate( dot( N, V ) );
	vec2 uv = vec2( roughness, sqrt( 1.0 - dotNV ) );
	uv = uv * LUT_SCALE + LUT_BIAS;
	return uv;
}
float LTC_ClippedSphereFormFactor( const in vec3 f ) {
	float l = length( f );
	return max( ( l * l + f.z ) / ( l + 1.0 ), 0.0 );
}
vec3 LTC_EdgeVectorFormFactor( const in vec3 v1, const in vec3 v2 ) {
	float x = dot( v1, v2 );
	float y = abs( x );
	float a = 0.8543985 + ( 0.4965155 + 0.0145206 * y ) * y;
	float b = 3.4175940 + ( 4.1616724 + y ) * y;
	float v = a / b;
	float theta_sintheta = ( x > 0.0 ) ? v : 0.5 * inversesqrt( max( 1.0 - x * x, 1e-7 ) ) - v;
	return cross( v1, v2 ) * theta_sintheta;
}
vec3 LTC_Evaluate( const in vec3 N, const in vec3 V, const in vec3 P, const in mat3 mInv, const in vec3 rectCoords[ 4 ] ) {
	vec3 v1 = rectCoords[ 1 ] - rectCoords[ 0 ];
	vec3 v2 = rectCoords[ 3 ] - rectCoords[ 0 ];
	vec3 lightNormal = cross( v1, v2 );
	if( dot( lightNormal, P - rectCoords[ 0 ] ) < 0.0 ) return vec3( 0.0 );
	vec3 T1, T2;
	T1 = normalize( V - N * dot( V, N ) );
	T2 = - cross( N, T1 );
	mat3 mat = mInv * transpose( mat3( T1, T2, N ) );
	vec3 coords[ 4 ];
	coords[ 0 ] = mat * ( rectCoords[ 0 ] - P );
	coords[ 1 ] = mat * ( rectCoords[ 1 ] - P );
	coords[ 2 ] = mat * ( rectCoords[ 2 ] - P );
	coords[ 3 ] = mat * ( rectCoords[ 3 ] - P );
	coords[ 0 ] = normalize( coords[ 0 ] );
	coords[ 1 ] = normalize( coords[ 1 ] );
	coords[ 2 ] = normalize( coords[ 2 ] );
	coords[ 3 ] = normalize( coords[ 3 ] );
	vec3 vectorFormFactor = vec3( 0.0 );
	vectorFormFactor += LTC_EdgeVectorFormFactor( coords[ 0 ], coords[ 1 ] );
	vectorFormFactor += LTC_EdgeVectorFormFactor( coords[ 1 ], coords[ 2 ] );
	vectorFormFactor += LTC_EdgeVectorFormFactor( coords[ 2 ], coords[ 3 ] );
	vectorFormFactor += LTC_EdgeVectorFormFactor( coords[ 3 ], coords[ 0 ] );
	float result = LTC_ClippedSphereFormFactor( vectorFormFactor );
	return vec3( result );
}
#if defined( USE_SHEEN )
float D_Charlie( float roughness, float dotNH ) {
	float alpha = pow2( roughness );
	float invAlpha = 1.0 / alpha;
	float cos2h = dotNH * dotNH;
	float sin2h = max( 1.0 - cos2h, 0.0078125 );
	return ( 2.0 + invAlpha ) * pow( sin2h, invAlpha * 0.5 ) / ( 2.0 * PI );
}
float V_Neubelt( float dotNV, float dotNL ) {
	return saturate( 1.0 / ( 4.0 * ( dotNL + dotNV - dotNL * dotNV ) ) );
}
vec3 BRDF_Sheen( const in vec3 lightDir, const in vec3 viewDir, const in vec3 normal, vec3 sheenColor, const in float sheenRoughness ) {
	vec3 halfDir = normalize( lightDir + viewDir );
	float dotNL = saturate( dot( normal, lightDir ) );
	float dotNV = saturate( dot( normal, viewDir ) );
	float dotNH = saturate( dot( normal, halfDir ) );
	float D = D_Charlie( sheenRoughness, dotNH );
	float V = V_Neubelt( dotNV, dotNL );
	return sheenColor * ( D * V );
}
#endif
float IBLSheenBRDF( const in vec3 normal, const in vec3 viewDir, const in float roughness ) {
	float dotNV = saturate( dot( normal, viewDir ) );
	float r2 = roughness * roughness;
	float rInv = 1.0 / ( roughness + 0.1 );
	float a = -1.9362 + 1.0678 * roughness + 0.4573 * r2 - 0.8469 * rInv;
	float b = -0.6014 + 0.5538 * roughness - 0.4670 * r2 - 0.1255 * rInv;
	float DG = exp( a * dotNV + b );
	return saturate( DG );
}
vec3 EnvironmentBRDF( const in vec3 normal, const in vec3 viewDir, const in vec3 specularColor, const in float specularF90, const in float roughness ) {
	float dotNV = saturate( dot( normal, viewDir ) );
	vec2 fab = texture2D( dfgLUT, vec2( roughness, dotNV ) ).rg;
	return specularColor * fab.x + specularF90 * fab.y;
}
#ifdef USE_IRIDESCENCE
void computeMultiscatteringIridescence( const in vec3 normal, const in vec3 viewDir, const in vec3 specularColor, const in float specularF90, const in float iridescence, const in vec3 iridescenceF0, const in float roughness, inout vec3 singleScatter, inout vec3 multiScatter ) {
#else
void computeMultiscattering( const in vec3 normal, const in vec3 viewDir, const in vec3 specularColor, const in float specularF90, const in float roughness, inout vec3 singleScatter, inout vec3 multiScatter ) {
#endif
	float dotNV = saturate( dot( normal, viewDir ) );
	vec2 fab = texture2D( dfgLUT, vec2( roughness, dotNV ) ).rg;
	#ifdef USE_IRIDESCENCE
		vec3 Fr = mix( specularColor, iridescenceF0, iridescence );
	#else
		vec3 Fr = specularColor;
	#endif
	vec3 FssEss = Fr * fab.x + specularF90 * fab.y;
	float Ess = fab.x + fab.y;
	float Ems = 1.0 - Ess;
	vec3 Favg = Fr + ( 1.0 - Fr ) * 0.047619;	vec3 Fms = FssEss * Favg / ( 1.0 - Ems * Favg );
	singleScatter += FssEss;
	multiScatter += Fms * Ems;
}
vec3 BRDF_GGX_Multiscatter( const in vec3 lightDir, const in vec3 viewDir, const in vec3 normal, const in PhysicalMaterial material ) {
	vec3 singleScatter = BRDF_GGX( lightDir, viewDir, normal, material );
	float dotNL = saturate( dot( normal, lightDir ) );
	float dotNV = saturate( dot( normal, viewDir ) );
	vec2 dfgV = texture2D( dfgLUT, vec2( material.roughness, dotNV ) ).rg;
	vec2 dfgL = texture2D( dfgLUT, vec2( material.roughness, dotNL ) ).rg;
	vec3 FssEss_V = material.specularColorBlended * dfgV.x + material.specularF90 * dfgV.y;
	vec3 FssEss_L = material.specularColorBlended * dfgL.x + material.specularF90 * dfgL.y;
	float Ess_V = dfgV.x + dfgV.y;
	float Ess_L = dfgL.x + dfgL.y;
	float Ems_V = 1.0 - Ess_V;
	float Ems_L = 1.0 - Ess_L;
	vec3 Favg = material.specularColorBlended + ( 1.0 - material.specularColorBlended ) * 0.047619;
	vec3 Fms = FssEss_V * FssEss_L * Favg / ( 1.0 - Ems_V * Ems_L * Favg + EPSILON );
	float compensationFactor = Ems_V * Ems_L;
	vec3 multiScatter = Fms * compensationFactor;
	return singleScatter + multiScatter;
}
#if NUM_RECT_AREA_LIGHTS > 0
	void RE_Direct_RectArea_Physical( const in RectAreaLight rectAreaLight, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in PhysicalMaterial material, inout ReflectedLight reflectedLight ) {
		vec3 normal = geometryNormal;
		vec3 viewDir = geometryViewDir;
		vec3 position = geometryPosition;
		vec3 lightPos = rectAreaLight.position;
		vec3 halfWidth = rectAreaLight.halfWidth;
		vec3 halfHeight = rectAreaLight.halfHeight;
		vec3 lightColor = rectAreaLight.color;
		float roughness = material.roughness;
		vec3 rectCoords[ 4 ];
		rectCoords[ 0 ] = lightPos + halfWidth - halfHeight;		rectCoords[ 1 ] = lightPos - halfWidth - halfHeight;
		rectCoords[ 2 ] = lightPos - halfWidth + halfHeight;
		rectCoords[ 3 ] = lightPos + halfWidth + halfHeight;
		vec2 uv = LTC_Uv( normal, viewDir, roughness );
		vec4 t1 = texture2D( ltc_1, uv );
		vec4 t2 = texture2D( ltc_2, uv );
		mat3 mInv = mat3(
			vec3( t1.x, 0, t1.y ),
			vec3(    0, 1,    0 ),
			vec3( t1.z, 0, t1.w )
		);
		vec3 fresnel = ( material.specularColorBlended * t2.x + ( material.specularF90 - material.specularColorBlended ) * t2.y );
		reflectedLight.directSpecular += lightColor * fresnel * LTC_Evaluate( normal, viewDir, position, mInv, rectCoords );
		reflectedLight.directDiffuse += lightColor * material.diffuseContribution * LTC_Evaluate( normal, viewDir, position, mat3( 1.0 ), rectCoords );
		#ifdef USE_CLEARCOAT
			vec3 Ncc = geometryClearcoatNormal;
			vec2 uvClearcoat = LTC_Uv( Ncc, viewDir, material.clearcoatRoughness );
			vec4 t1Clearcoat = texture2D( ltc_1, uvClearcoat );
			vec4 t2Clearcoat = texture2D( ltc_2, uvClearcoat );
			mat3 mInvClearcoat = mat3(
				vec3( t1Clearcoat.x, 0, t1Clearcoat.y ),
				vec3(             0, 1,             0 ),
				vec3( t1Clearcoat.z, 0, t1Clearcoat.w )
			);
			vec3 fresnelClearcoat = material.clearcoatF0 * t2Clearcoat.x + ( material.clearcoatF90 - material.clearcoatF0 ) * t2Clearcoat.y;
			clearcoatSpecularDirect += lightColor * fresnelClearcoat * LTC_Evaluate( Ncc, viewDir, position, mInvClearcoat, rectCoords );
		#endif
	}
#endif
void RE_Direct_Physical( const in IncidentLight directLight, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in PhysicalMaterial material, inout ReflectedLight reflectedLight ) {
	float dotNL = saturate( dot( geometryNormal, directLight.direction ) );
	vec3 irradiance = dotNL * directLight.color;
	#ifdef USE_CLEARCOAT
		float dotNLcc = saturate( dot( geometryClearcoatNormal, directLight.direction ) );
		vec3 ccIrradiance = dotNLcc * directLight.color;
		clearcoatSpecularDirect += ccIrradiance * BRDF_GGX_Clearcoat( directLight.direction, geometryViewDir, geometryClearcoatNormal, material );
	#endif
	#ifdef USE_SHEEN
 
 		sheenSpecularDirect += irradiance * BRDF_Sheen( directLight.direction, geometryViewDir, geometryNormal, material.sheenColor, material.sheenRoughness );
 
 		float sheenAlbedoV = IBLSheenBRDF( geometryNormal, geometryViewDir, material.sheenRoughness );
 		float sheenAlbedoL = IBLSheenBRDF( geometryNormal, directLight.direction, material.sheenRoughness );
 
 		float sheenEnergyComp = 1.0 - max3( material.sheenColor ) * max( sheenAlbedoV, sheenAlbedoL );
 
 		irradiance *= sheenEnergyComp;
 
 	#endif
	reflectedLight.directSpecular += irradiance * BRDF_GGX_Multiscatter( directLight.direction, geometryViewDir, geometryNormal, material );
	reflectedLight.directDiffuse += irradiance * BRDF_Lambert( material.diffuseContribution );
}
void RE_IndirectDiffuse_Physical( const in vec3 irradiance, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in PhysicalMaterial material, inout ReflectedLight reflectedLight ) {
	vec3 diffuse = irradiance * BRDF_Lambert( material.diffuseContribution );
	#ifdef USE_SHEEN
		float sheenAlbedo = IBLSheenBRDF( geometryNormal, geometryViewDir, material.sheenRoughness );
		float sheenEnergyComp = 1.0 - max3( material.sheenColor ) * sheenAlbedo;
		diffuse *= sheenEnergyComp;
	#endif
	reflectedLight.indirectDiffuse += diffuse;
}
void RE_IndirectSpecular_Physical( const in vec3 radiance, const in vec3 irradiance, const in vec3 clearcoatRadiance, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in PhysicalMaterial material, inout ReflectedLight reflectedLight) {
	#ifdef USE_CLEARCOAT
		clearcoatSpecularIndirect += clearcoatRadiance * EnvironmentBRDF( geometryClearcoatNormal, geometryViewDir, material.clearcoatF0, material.clearcoatF90, material.clearcoatRoughness );
	#endif
	#ifdef USE_SHEEN
		sheenSpecularIndirect += irradiance * material.sheenColor * IBLSheenBRDF( geometryNormal, geometryViewDir, material.sheenRoughness ) * RECIPROCAL_PI;
 	#endif
	vec3 singleScatteringDielectric = vec3( 0.0 );
	vec3 multiScatteringDielectric = vec3( 0.0 );
	vec3 singleScatteringMetallic = vec3( 0.0 );
	vec3 multiScatteringMetallic = vec3( 0.0 );
	#ifdef USE_IRIDESCENCE
		computeMultiscatteringIridescence( geometryNormal, geometryViewDir, material.specularColor, material.specularF90, material.iridescence, material.iridescenceFresnelDielectric, material.roughness, singleScatteringDielectric, multiScatteringDielectric );
		computeMultiscatteringIridescence( geometryNormal, geometryViewDir, material.diffuseColor, material.specularF90, material.iridescence, material.iridescenceFresnelMetallic, material.roughness, singleScatteringMetallic, multiScatteringMetallic );
	#else
		computeMultiscattering( geometryNormal, geometryViewDir, material.specularColor, material.specularF90, material.roughness, singleScatteringDielectric, multiScatteringDielectric );
		computeMultiscattering( geometryNormal, geometryViewDir, material.diffuseColor, material.specularF90, material.roughness, singleScatteringMetallic, multiScatteringMetallic );
	#endif
	vec3 singleScattering = mix( singleScatteringDielectric, singleScatteringMetallic, material.metalness );
	vec3 multiScattering = mix( multiScatteringDielectric, multiScatteringMetallic, material.metalness );
	vec3 totalScatteringDielectric = singleScatteringDielectric + multiScatteringDielectric;
	vec3 diffuse = material.diffuseContribution * ( 1.0 - totalScatteringDielectric );
	vec3 cosineWeightedIrradiance = irradiance * RECIPROCAL_PI;
	vec3 indirectSpecular = radiance * singleScattering;
	indirectSpecular += multiScattering * cosineWeightedIrradiance;
	vec3 indirectDiffuse = diffuse * cosineWeightedIrradiance;
	#ifdef USE_SHEEN
		float sheenAlbedo = IBLSheenBRDF( geometryNormal, geometryViewDir, material.sheenRoughness );
		float sheenEnergyComp = 1.0 - max3( material.sheenColor ) * sheenAlbedo;
		indirectSpecular *= sheenEnergyComp;
		indirectDiffuse *= sheenEnergyComp;
	#endif
	reflectedLight.indirectSpecular += indirectSpecular;
	reflectedLight.indirectDiffuse += indirectDiffuse;
}
#define RE_Direct				RE_Direct_Physical
#define RE_Direct_RectArea		RE_Direct_RectArea_Physical
#define RE_IndirectDiffuse		RE_IndirectDiffuse_Physical
#define RE_IndirectSpecular		RE_IndirectSpecular_Physical
float computeSpecularOcclusion( const in float dotNV, const in float ambientOcclusion, const in float roughness ) {
	return saturate( pow( dotNV + ambientOcclusion, exp2( - 16.0 * roughness - 1.0 ) ) - 1.0 + ambientOcclusion );
}`,yl=`
vec3 geometryPosition = - vViewPosition;
vec3 geometryNormal = normal;
vec3 geometryViewDir = ( isOrthographic ) ? vec3( 0, 0, 1 ) : normalize( vViewPosition );
vec3 geometryClearcoatNormal = vec3( 0.0 );
#ifdef USE_CLEARCOAT
	geometryClearcoatNormal = clearcoatNormal;
#endif
#ifdef USE_IRIDESCENCE
	float dotNVi = saturate( dot( normal, geometryViewDir ) );
	if ( material.iridescenceThickness == 0.0 ) {
		material.iridescence = 0.0;
	} else {
		material.iridescence = saturate( material.iridescence );
	}
	if ( material.iridescence > 0.0 ) {
		material.iridescenceFresnelDielectric = evalIridescence( 1.0, material.iridescenceIOR, dotNVi, material.iridescenceThickness, material.specularColor );
		material.iridescenceFresnelMetallic = evalIridescence( 1.0, material.iridescenceIOR, dotNVi, material.iridescenceThickness, material.diffuseColor );
		material.iridescenceFresnel = mix( material.iridescenceFresnelDielectric, material.iridescenceFresnelMetallic, material.metalness );
		material.iridescenceF0 = Schlick_to_F0( material.iridescenceFresnel, 1.0, dotNVi );
	}
#endif
IncidentLight directLight;
#if ( NUM_POINT_LIGHTS > 0 ) && defined( RE_Direct )
	PointLight pointLight;
	#if defined( USE_SHADOWMAP ) && NUM_POINT_LIGHT_SHADOWS > 0
	PointLightShadow pointLightShadow;
	#endif
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_POINT_LIGHTS; i ++ ) {
		pointLight = pointLights[ i ];
		getPointLightInfo( pointLight, geometryPosition, directLight );
		#if defined( USE_SHADOWMAP ) && ( UNROLLED_LOOP_INDEX < NUM_POINT_LIGHT_SHADOWS ) && ( defined( SHADOWMAP_TYPE_PCF ) || defined( SHADOWMAP_TYPE_BASIC ) )
		pointLightShadow = pointLightShadows[ i ];
		directLight.color *= ( directLight.visible && receiveShadow ) ? getPointShadow( pointShadowMap[ i ], pointLightShadow.shadowMapSize, pointLightShadow.shadowIntensity, pointLightShadow.shadowBias, pointLightShadow.shadowRadius, vPointShadowCoord[ i ], pointLightShadow.shadowCameraNear, pointLightShadow.shadowCameraFar ) : 1.0;
		#endif
		RE_Direct( directLight, geometryPosition, geometryNormal, geometryViewDir, geometryClearcoatNormal, material, reflectedLight );
	}
	#pragma unroll_loop_end
#endif
#if ( NUM_SPOT_LIGHTS > 0 ) && defined( RE_Direct )
	SpotLight spotLight;
	vec4 spotColor;
	vec3 spotLightCoord;
	bool inSpotLightMap;
	#if defined( USE_SHADOWMAP ) && NUM_SPOT_LIGHT_SHADOWS > 0
	SpotLightShadow spotLightShadow;
	#endif
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_SPOT_LIGHTS; i ++ ) {
		spotLight = spotLights[ i ];
		getSpotLightInfo( spotLight, geometryPosition, directLight );
		#if ( UNROLLED_LOOP_INDEX < NUM_SPOT_LIGHT_SHADOWS_WITH_MAPS )
		#define SPOT_LIGHT_MAP_INDEX UNROLLED_LOOP_INDEX
		#elif ( UNROLLED_LOOP_INDEX < NUM_SPOT_LIGHT_SHADOWS )
		#define SPOT_LIGHT_MAP_INDEX NUM_SPOT_LIGHT_MAPS
		#else
		#define SPOT_LIGHT_MAP_INDEX ( UNROLLED_LOOP_INDEX - NUM_SPOT_LIGHT_SHADOWS + NUM_SPOT_LIGHT_SHADOWS_WITH_MAPS )
		#endif
		#if ( SPOT_LIGHT_MAP_INDEX < NUM_SPOT_LIGHT_MAPS )
			spotLightCoord = vSpotLightCoord[ i ].xyz / vSpotLightCoord[ i ].w;
			inSpotLightMap = all( lessThan( abs( spotLightCoord * 2. - 1. ), vec3( 1.0 ) ) );
			spotColor = texture2D( spotLightMap[ SPOT_LIGHT_MAP_INDEX ], spotLightCoord.xy );
			directLight.color = inSpotLightMap ? directLight.color * spotColor.rgb : directLight.color;
		#endif
		#undef SPOT_LIGHT_MAP_INDEX
		#if defined( USE_SHADOWMAP ) && ( UNROLLED_LOOP_INDEX < NUM_SPOT_LIGHT_SHADOWS )
		spotLightShadow = spotLightShadows[ i ];
		directLight.color *= ( directLight.visible && receiveShadow ) ? getShadow( spotShadowMap[ i ], spotLightShadow.shadowMapSize, spotLightShadow.shadowIntensity, spotLightShadow.shadowBias, spotLightShadow.shadowRadius, vSpotLightCoord[ i ] ) : 1.0;
		#endif
		RE_Direct( directLight, geometryPosition, geometryNormal, geometryViewDir, geometryClearcoatNormal, material, reflectedLight );
	}
	#pragma unroll_loop_end
#endif
#if ( NUM_DIR_LIGHTS > 0 ) && defined( RE_Direct )
	DirectionalLight directionalLight;
	#if defined( USE_SHADOWMAP ) && NUM_DIR_LIGHT_SHADOWS > 0
	DirectionalLightShadow directionalLightShadow;
	#endif
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_DIR_LIGHTS; i ++ ) {
		directionalLight = directionalLights[ i ];
		getDirectionalLightInfo( directionalLight, directLight );
		#if defined( USE_SHADOWMAP ) && ( UNROLLED_LOOP_INDEX < NUM_DIR_LIGHT_SHADOWS )
		directionalLightShadow = directionalLightShadows[ i ];
		directLight.color *= ( directLight.visible && receiveShadow ) ? getShadow( directionalShadowMap[ i ], directionalLightShadow.shadowMapSize, directionalLightShadow.shadowIntensity, directionalLightShadow.shadowBias, directionalLightShadow.shadowRadius, vDirectionalShadowCoord[ i ] ) : 1.0;
		#endif
		RE_Direct( directLight, geometryPosition, geometryNormal, geometryViewDir, geometryClearcoatNormal, material, reflectedLight );
	}
	#pragma unroll_loop_end
#endif
#if ( NUM_RECT_AREA_LIGHTS > 0 ) && defined( RE_Direct_RectArea )
	RectAreaLight rectAreaLight;
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_RECT_AREA_LIGHTS; i ++ ) {
		rectAreaLight = rectAreaLights[ i ];
		RE_Direct_RectArea( rectAreaLight, geometryPosition, geometryNormal, geometryViewDir, geometryClearcoatNormal, material, reflectedLight );
	}
	#pragma unroll_loop_end
#endif
#if defined( RE_IndirectDiffuse )
	vec3 iblIrradiance = vec3( 0.0 );
	vec3 irradiance = getAmbientLightIrradiance( ambientLightColor );
	#if defined( USE_LIGHT_PROBES )
		irradiance += getLightProbeIrradiance( lightProbe, geometryNormal );
	#endif
	#if ( NUM_HEMI_LIGHTS > 0 )
		#pragma unroll_loop_start
		for ( int i = 0; i < NUM_HEMI_LIGHTS; i ++ ) {
			irradiance += getHemisphereLightIrradiance( hemisphereLights[ i ], geometryNormal );
		}
		#pragma unroll_loop_end
	#endif
#endif
#if defined( RE_IndirectSpecular )
	vec3 radiance = vec3( 0.0 );
	vec3 clearcoatRadiance = vec3( 0.0 );
#endif`,Ml=`#if defined( RE_IndirectDiffuse )
	#ifdef USE_LIGHTMAP
		vec4 lightMapTexel = texture2D( lightMap, vLightMapUv );
		vec3 lightMapIrradiance = lightMapTexel.rgb * lightMapIntensity;
		irradiance += lightMapIrradiance;
	#endif
	#if defined( USE_ENVMAP ) && defined( ENVMAP_TYPE_CUBE_UV )
		#if defined( STANDARD ) || defined( LAMBERT ) || defined( PHONG )
			iblIrradiance += getIBLIrradiance( geometryNormal );
		#endif
	#endif
#endif
#if defined( USE_ENVMAP ) && defined( RE_IndirectSpecular )
	#ifdef USE_ANISOTROPY
		radiance += getIBLAnisotropyRadiance( geometryViewDir, geometryNormal, material.roughness, material.anisotropyB, material.anisotropy );
	#else
		radiance += getIBLRadiance( geometryViewDir, geometryNormal, material.roughness );
	#endif
	#ifdef USE_CLEARCOAT
		clearcoatRadiance += getIBLRadiance( geometryViewDir, geometryClearcoatNormal, material.clearcoatRoughness );
	#endif
#endif`,Sl=`#if defined( RE_IndirectDiffuse )
	#if defined( LAMBERT ) || defined( PHONG )
		irradiance += iblIrradiance;
	#endif
	RE_IndirectDiffuse( irradiance, geometryPosition, geometryNormal, geometryViewDir, geometryClearcoatNormal, material, reflectedLight );
#endif
#if defined( RE_IndirectSpecular )
	RE_IndirectSpecular( radiance, iblIrradiance, clearcoatRadiance, geometryPosition, geometryNormal, geometryViewDir, geometryClearcoatNormal, material, reflectedLight );
#endif`,Zr=`#if defined( USE_LOGARITHMIC_DEPTH_BUFFER )
	gl_FragDepth = vIsPerspective == 0.0 ? gl_FragCoord.z : log2( vFragDepth ) * logDepthBufFC * 0.5;
#endif`,bl=`#if defined( USE_LOGARITHMIC_DEPTH_BUFFER )
	uniform float logDepthBufFC;
	varying float vFragDepth;
	varying float vIsPerspective;
#endif`,lr=`#ifdef USE_LOGARITHMIC_DEPTH_BUFFER
	varying float vFragDepth;
	varying float vIsPerspective;
#endif`,Yi=`#ifdef USE_LOGARITHMIC_DEPTH_BUFFER
	vFragDepth = 1.0 + gl_Position.w;
	vIsPerspective = float( isPerspectiveMatrix( projectionMatrix ) );
#endif`,$r=`#ifdef USE_MAP
	vec4 sampledDiffuseColor = texture2D( map, vMapUv );
	#ifdef DECODE_VIDEO_TEXTURE
		sampledDiffuseColor = sRGBTransferEOTF( sampledDiffuseColor );
	#endif
	diffuseColor *= sampledDiffuseColor;
#endif`,Jr=`#ifdef USE_MAP
	uniform sampler2D map;
#endif`,Kr=`#if defined( USE_MAP ) || defined( USE_ALPHAMAP )
	#if defined( USE_POINTS_UV )
		vec2 uv = vUv;
	#else
		vec2 uv = ( uvTransform * vec3( gl_PointCoord.x, 1.0 - gl_PointCoord.y, 1 ) ).xy;
	#endif
#endif
#ifdef USE_MAP
	diffuseColor *= texture2D( map, uv );
#endif
#ifdef USE_ALPHAMAP
	diffuseColor.a *= texture2D( alphaMap, uv ).g;
#endif`,Qr=`#if defined( USE_POINTS_UV )
	varying vec2 vUv;
#else
	#if defined( USE_MAP ) || defined( USE_ALPHAMAP )
		uniform mat3 uvTransform;
	#endif
#endif
#ifdef USE_MAP
	uniform sampler2D map;
#endif
#ifdef USE_ALPHAMAP
	uniform sampler2D alphaMap;
#endif`,ws=`float metalnessFactor = metalness;
#ifdef USE_METALNESSMAP
	vec4 texelMetalness = texture2D( metalnessMap, vMetalnessMapUv );
	metalnessFactor *= texelMetalness.b;
#endif`,Ln=`#ifdef USE_METALNESSMAP
	uniform sampler2D metalnessMap;
#endif`,Cs=`#ifdef USE_INSTANCING_MORPH
	float morphTargetInfluences[ MORPHTARGETS_COUNT ];
	float morphTargetBaseInfluence = texelFetch( morphTexture, ivec2( 0, gl_InstanceID ), 0 ).r;
	for ( int i = 0; i < MORPHTARGETS_COUNT; i ++ ) {
		morphTargetInfluences[i] =  texelFetch( morphTexture, ivec2( i + 1, gl_InstanceID ), 0 ).r;
	}
#endif`,Mn=`#if defined( USE_MORPHCOLORS )
	vColor *= morphTargetBaseInfluence;
	for ( int i = 0; i < MORPHTARGETS_COUNT; i ++ ) {
		#if defined( USE_COLOR_ALPHA )
			if ( morphTargetInfluences[ i ] != 0.0 ) vColor += getMorph( gl_VertexID, i, 2 ) * morphTargetInfluences[ i ];
		#elif defined( USE_COLOR )
			if ( morphTargetInfluences[ i ] != 0.0 ) vColor += getMorph( gl_VertexID, i, 2 ).rgb * morphTargetInfluences[ i ];
		#endif
	}
#endif`,jr=`#ifdef USE_MORPHNORMALS
	objectNormal *= morphTargetBaseInfluence;
	for ( int i = 0; i < MORPHTARGETS_COUNT; i ++ ) {
		if ( morphTargetInfluences[ i ] != 0.0 ) objectNormal += getMorph( gl_VertexID, i, 1 ).xyz * morphTargetInfluences[ i ];
	}
#endif`,Ch=`#ifdef USE_MORPHTARGETS
	#ifndef USE_INSTANCING_MORPH
		uniform float morphTargetBaseInfluence;
		uniform float morphTargetInfluences[ MORPHTARGETS_COUNT ];
	#endif
	uniform sampler2DArray morphTargetsTexture;
	uniform ivec2 morphTargetsTextureSize;
	vec4 getMorph( const in int vertexIndex, const in int morphTargetIndex, const in int offset ) {
		int texelIndex = vertexIndex * MORPHTARGETS_TEXTURE_STRIDE + offset;
		int y = texelIndex / morphTargetsTextureSize.x;
		int x = texelIndex - y * morphTargetsTextureSize.x;
		ivec3 morphUV = ivec3( x, y, morphTargetIndex );
		return texelFetch( morphTargetsTexture, morphUV, 0 );
	}
#endif`,ta=`#ifdef USE_MORPHTARGETS
	transformed *= morphTargetBaseInfluence;
	for ( int i = 0; i < MORPHTARGETS_COUNT; i ++ ) {
		if ( morphTargetInfluences[ i ] != 0.0 ) transformed += getMorph( gl_VertexID, i, 0 ).xyz * morphTargetInfluences[ i ];
	}
#endif`,Rh=`float faceDirection = gl_FrontFacing ? 1.0 : - 1.0;
#ifdef FLAT_SHADED
	vec3 fdx = dFdx( vViewPosition );
	vec3 fdy = dFdy( vViewPosition );
	vec3 normal = normalize( cross( fdx, fdy ) );
#else
	vec3 normal = normalize( vNormal );
	#ifdef DOUBLE_SIDED
		normal *= faceDirection;
	#endif
#endif
#if defined( USE_NORMALMAP_TANGENTSPACE ) || defined( USE_CLEARCOAT_NORMALMAP ) || defined( USE_ANISOTROPY )
	#ifdef USE_TANGENT
		mat3 tbn = mat3( normalize( vTangent ), normalize( vBitangent ), normal );
	#else
		mat3 tbn = getTangentFrame( - vViewPosition, normal,
		#if defined( USE_NORMALMAP )
			vNormalMapUv
		#elif defined( USE_CLEARCOAT_NORMALMAP )
			vClearcoatNormalMapUv
		#else
			vUv
		#endif
		);
	#endif
	#if defined( DOUBLE_SIDED ) && ! defined( FLAT_SHADED )
		tbn[0] *= faceDirection;
		tbn[1] *= faceDirection;
	#endif
#endif
#ifdef USE_CLEARCOAT_NORMALMAP
	#ifdef USE_TANGENT
		mat3 tbn2 = mat3( normalize( vTangent ), normalize( vBitangent ), normal );
	#else
		mat3 tbn2 = getTangentFrame( - vViewPosition, normal, vClearcoatNormalMapUv );
	#endif
	#if defined( DOUBLE_SIDED ) && ! defined( FLAT_SHADED )
		tbn2[0] *= faceDirection;
		tbn2[1] *= faceDirection;
	#endif
#endif
vec3 nonPerturbedNormal = normal;`,Cn=`#ifdef USE_NORMALMAP_OBJECTSPACE
	normal = texture2D( normalMap, vNormalMapUv ).xyz * 2.0 - 1.0;
	#ifdef FLIP_SIDED
		normal = - normal;
	#endif
	#ifdef DOUBLE_SIDED
		normal = normal * faceDirection;
	#endif
	normal = normalize( normalMatrix * normal );
#elif defined( USE_NORMALMAP_TANGENTSPACE )
	vec3 mapN = texture2D( normalMap, vNormalMapUv ).xyz * 2.0 - 1.0;
	mapN.xy *= normalScale;
	normal = normalize( tbn * mapN );
#elif defined( USE_BUMPMAP )
	normal = perturbNormalArb( - vViewPosition, normal, dHdxy_fwd(), faceDirection );
#endif`,ea=`#ifndef FLAT_SHADED
	varying vec3 vNormal;
	#ifdef USE_TANGENT
		varying vec3 vTangent;
		varying vec3 vBitangent;
	#endif
#endif`,Ih=`#ifndef FLAT_SHADED
	varying vec3 vNormal;
	#ifdef USE_TANGENT
		varying vec3 vTangent;
		varying vec3 vBitangent;
	#endif
#endif`,Rs=`#ifndef FLAT_SHADED
	vNormal = normalize( transformedNormal );
	#ifdef USE_TANGENT
		vTangent = normalize( transformedTangent );
		vBitangent = normalize( cross( vNormal, vTangent ) * tangent.w );
	#endif
#endif`,Ph=`#ifdef USE_NORMALMAP
	uniform sampler2D normalMap;
	uniform vec2 normalScale;
#endif
#ifdef USE_NORMALMAP_OBJECTSPACE
	uniform mat3 normalMatrix;
#endif
#if ! defined ( USE_TANGENT ) && ( defined ( USE_NORMALMAP_TANGENTSPACE ) || defined ( USE_CLEARCOAT_NORMALMAP ) || defined( USE_ANISOTROPY ) )
	mat3 getTangentFrame( vec3 eye_pos, vec3 surf_norm, vec2 uv ) {
		vec3 q0 = dFdx( eye_pos.xyz );
		vec3 q1 = dFdy( eye_pos.xyz );
		vec2 st0 = dFdx( uv.st );
		vec2 st1 = dFdy( uv.st );
		vec3 N = surf_norm;
		vec3 q1perp = cross( q1, N );
		vec3 q0perp = cross( N, q0 );
		vec3 T = q1perp * st0.x + q0perp * st1.x;
		vec3 B = q1perp * st0.y + q0perp * st1.y;
		float det = max( dot( T, T ), dot( B, B ) );
		float scale = ( det == 0.0 ) ? 0.0 : inversesqrt( det );
		return mat3( T * scale, B * scale, N );
	}
#endif`,Is=`#ifdef USE_CLEARCOAT
	vec3 clearcoatNormal = nonPerturbedNormal;
#endif`,na=`#ifdef USE_CLEARCOAT_NORMALMAP
	vec3 clearcoatMapN = texture2D( clearcoatNormalMap, vClearcoatNormalMapUv ).xyz * 2.0 - 1.0;
	clearcoatMapN.xy *= clearcoatNormalScale;
	clearcoatNormal = normalize( tbn2 * clearcoatMapN );
#endif`,ia=`#ifdef USE_CLEARCOATMAP
	uniform sampler2D clearcoatMap;
#endif
#ifdef USE_CLEARCOAT_NORMALMAP
	uniform sampler2D clearcoatNormalMap;
	uniform vec2 clearcoatNormalScale;
#endif
#ifdef USE_CLEARCOAT_ROUGHNESSMAP
	uniform sampler2D clearcoatRoughnessMap;
#endif`,sa=`#ifdef USE_IRIDESCENCEMAP
	uniform sampler2D iridescenceMap;
#endif
#ifdef USE_IRIDESCENCE_THICKNESSMAP
	uniform sampler2D iridescenceThicknessMap;
#endif`,ra=`#ifdef OPAQUE
diffuseColor.a = 1.0;
#endif
#ifdef USE_TRANSMISSION
diffuseColor.a *= material.transmissionAlpha;
#endif
gl_FragColor = vec4( outgoingLight, diffuseColor.a );`,Zi=`vec3 packNormalToRGB( const in vec3 normal ) {
	return normalize( normal ) * 0.5 + 0.5;
}
vec3 unpackRGBToNormal( const in vec3 rgb ) {
	return 2.0 * rgb.xyz - 1.0;
}
const float PackUpscale = 256. / 255.;const float UnpackDownscale = 255. / 256.;const float ShiftRight8 = 1. / 256.;
const float Inv255 = 1. / 255.;
const vec4 PackFactors = vec4( 1.0, 256.0, 256.0 * 256.0, 256.0 * 256.0 * 256.0 );
const vec2 UnpackFactors2 = vec2( UnpackDownscale, 1.0 / PackFactors.g );
const vec3 UnpackFactors3 = vec3( UnpackDownscale / PackFactors.rg, 1.0 / PackFactors.b );
const vec4 UnpackFactors4 = vec4( UnpackDownscale / PackFactors.rgb, 1.0 / PackFactors.a );
vec4 packDepthToRGBA( const in float v ) {
	if( v <= 0.0 )
		return vec4( 0., 0., 0., 0. );
	if( v >= 1.0 )
		return vec4( 1., 1., 1., 1. );
	float vuf;
	float af = modf( v * PackFactors.a, vuf );
	float bf = modf( vuf * ShiftRight8, vuf );
	float gf = modf( vuf * ShiftRight8, vuf );
	return vec4( vuf * Inv255, gf * PackUpscale, bf * PackUpscale, af );
}
vec3 packDepthToRGB( const in float v ) {
	if( v <= 0.0 )
		return vec3( 0., 0., 0. );
	if( v >= 1.0 )
		return vec3( 1., 1., 1. );
	float vuf;
	float bf = modf( v * PackFactors.b, vuf );
	float gf = modf( vuf * ShiftRight8, vuf );
	return vec3( vuf * Inv255, gf * PackUpscale, bf );
}
vec2 packDepthToRG( const in float v ) {
	if( v <= 0.0 )
		return vec2( 0., 0. );
	if( v >= 1.0 )
		return vec2( 1., 1. );
	float vuf;
	float gf = modf( v * 256., vuf );
	return vec2( vuf * Inv255, gf );
}
float unpackRGBAToDepth( const in vec4 v ) {
	return dot( v, UnpackFactors4 );
}
float unpackRGBToDepth( const in vec3 v ) {
	return dot( v, UnpackFactors3 );
}
float unpackRGToDepth( const in vec2 v ) {
	return v.r * UnpackFactors2.r + v.g * UnpackFactors2.g;
}
vec4 pack2HalfToRGBA( const in vec2 v ) {
	vec4 r = vec4( v.x, fract( v.x * 255.0 ), v.y, fract( v.y * 255.0 ) );
	return vec4( r.x - r.y / 255.0, r.y, r.z - r.w / 255.0, r.w );
}
vec2 unpackRGBATo2Half( const in vec4 v ) {
	return vec2( v.x + ( v.y / 255.0 ), v.z + ( v.w / 255.0 ) );
}
float viewZToOrthographicDepth( const in float viewZ, const in float near, const in float far ) {
	return ( viewZ + near ) / ( near - far );
}
float orthographicDepthToViewZ( const in float depth, const in float near, const in float far ) {
	#ifdef USE_REVERSED_DEPTH_BUFFER
	
		return depth * ( far - near ) - far;
	#else
		return depth * ( near - far ) - near;
	#endif
}
float viewZToPerspectiveDepth( const in float viewZ, const in float near, const in float far ) {
	return ( ( near + viewZ ) * far ) / ( ( far - near ) * viewZ );
}
float perspectiveDepthToViewZ( const in float depth, const in float near, const in float far ) {
	
	#ifdef USE_REVERSED_DEPTH_BUFFER
		return ( near * far ) / ( ( near - far ) * depth - near );
	#else
		return ( near * far ) / ( ( far - near ) * depth - far );
	#endif
}`,gi=`#ifdef PREMULTIPLIED_ALPHA
	gl_FragColor.rgb *= gl_FragColor.a;
#endif`,aa=`vec4 mvPosition = vec4( transformed, 1.0 );
#ifdef USE_BATCHING
	mvPosition = batchingMatrix * mvPosition;
#endif
#ifdef USE_INSTANCING
	mvPosition = instanceMatrix * mvPosition;
#endif
mvPosition = modelViewMatrix * mvPosition;
gl_Position = projectionMatrix * mvPosition;`,oa=`#ifdef DITHERING
	gl_FragColor.rgb = dithering( gl_FragColor.rgb );
#endif`,la=`#ifdef DITHERING
	vec3 dithering( vec3 color ) {
		float grid_position = rand( gl_FragCoord.xy );
		vec3 dither_shift_RGB = vec3( 0.25 / 255.0, -0.25 / 255.0, 0.25 / 255.0 );
		dither_shift_RGB = mix( 2.0 * dither_shift_RGB, -2.0 * dither_shift_RGB, grid_position );
		return color + dither_shift_RGB;
	}
#endif`,Tl=`float roughnessFactor = roughness;
#ifdef USE_ROUGHNESSMAP
	vec4 texelRoughness = texture2D( roughnessMap, vRoughnessMapUv );
	roughnessFactor *= texelRoughness.g;
#endif`,ca=`#ifdef USE_ROUGHNESSMAP
	uniform sampler2D roughnessMap;
#endif`,ha=`#if NUM_SPOT_LIGHT_COORDS > 0
	varying vec4 vSpotLightCoord[ NUM_SPOT_LIGHT_COORDS ];
#endif
#if NUM_SPOT_LIGHT_MAPS > 0
	uniform sampler2D spotLightMap[ NUM_SPOT_LIGHT_MAPS ];
#endif
#ifdef USE_SHADOWMAP
	#if NUM_DIR_LIGHT_SHADOWS > 0
		#if defined( SHADOWMAP_TYPE_PCF )
			uniform sampler2DShadow directionalShadowMap[ NUM_DIR_LIGHT_SHADOWS ];
		#else
			uniform sampler2D directionalShadowMap[ NUM_DIR_LIGHT_SHADOWS ];
		#endif
		varying vec4 vDirectionalShadowCoord[ NUM_DIR_LIGHT_SHADOWS ];
		struct DirectionalLightShadow {
			float shadowIntensity;
			float shadowBias;
			float shadowNormalBias;
			float shadowRadius;
			vec2 shadowMapSize;
		};
		uniform DirectionalLightShadow directionalLightShadows[ NUM_DIR_LIGHT_SHADOWS ];
	#endif
	#if NUM_SPOT_LIGHT_SHADOWS > 0
		#if defined( SHADOWMAP_TYPE_PCF )
			uniform sampler2DShadow spotShadowMap[ NUM_SPOT_LIGHT_SHADOWS ];
		#else
			uniform sampler2D spotShadowMap[ NUM_SPOT_LIGHT_SHADOWS ];
		#endif
		struct SpotLightShadow {
			float shadowIntensity;
			float shadowBias;
			float shadowNormalBias;
			float shadowRadius;
			vec2 shadowMapSize;
		};
		uniform SpotLightShadow spotLightShadows[ NUM_SPOT_LIGHT_SHADOWS ];
	#endif
	#if NUM_POINT_LIGHT_SHADOWS > 0
		#if defined( SHADOWMAP_TYPE_PCF )
			uniform samplerCubeShadow pointShadowMap[ NUM_POINT_LIGHT_SHADOWS ];
		#elif defined( SHADOWMAP_TYPE_BASIC )
			uniform samplerCube pointShadowMap[ NUM_POINT_LIGHT_SHADOWS ];
		#endif
		varying vec4 vPointShadowCoord[ NUM_POINT_LIGHT_SHADOWS ];
		struct PointLightShadow {
			float shadowIntensity;
			float shadowBias;
			float shadowNormalBias;
			float shadowRadius;
			vec2 shadowMapSize;
			float shadowCameraNear;
			float shadowCameraFar;
		};
		uniform PointLightShadow pointLightShadows[ NUM_POINT_LIGHT_SHADOWS ];
	#endif
	#if defined( SHADOWMAP_TYPE_PCF )
		float interleavedGradientNoise( vec2 position ) {
			return fract( 52.9829189 * fract( dot( position, vec2( 0.06711056, 0.00583715 ) ) ) );
		}
		vec2 vogelDiskSample( int sampleIndex, int samplesCount, float phi ) {
			const float goldenAngle = 2.399963229728653;
			float r = sqrt( ( float( sampleIndex ) + 0.5 ) / float( samplesCount ) );
			float theta = float( sampleIndex ) * goldenAngle + phi;
			return vec2( cos( theta ), sin( theta ) ) * r;
		}
	#endif
	#if defined( SHADOWMAP_TYPE_PCF )
		float getShadow( sampler2DShadow shadowMap, vec2 shadowMapSize, float shadowIntensity, float shadowBias, float shadowRadius, vec4 shadowCoord ) {
			float shadow = 1.0;
			shadowCoord.xyz /= shadowCoord.w;
			shadowCoord.z += shadowBias;
			bool inFrustum = shadowCoord.x >= 0.0 && shadowCoord.x <= 1.0 && shadowCoord.y >= 0.0 && shadowCoord.y <= 1.0;
			bool frustumTest = inFrustum && shadowCoord.z <= 1.0;
			if ( frustumTest ) {
				vec2 texelSize = vec2( 1.0 ) / shadowMapSize;
				float radius = shadowRadius * texelSize.x;
				float phi = interleavedGradientNoise( gl_FragCoord.xy ) * PI2;
				shadow = (
					texture( shadowMap, vec3( shadowCoord.xy + vogelDiskSample( 0, 5, phi ) * radius, shadowCoord.z ) ) +
					texture( shadowMap, vec3( shadowCoord.xy + vogelDiskSample( 1, 5, phi ) * radius, shadowCoord.z ) ) +
					texture( shadowMap, vec3( shadowCoord.xy + vogelDiskSample( 2, 5, phi ) * radius, shadowCoord.z ) ) +
					texture( shadowMap, vec3( shadowCoord.xy + vogelDiskSample( 3, 5, phi ) * radius, shadowCoord.z ) ) +
					texture( shadowMap, vec3( shadowCoord.xy + vogelDiskSample( 4, 5, phi ) * radius, shadowCoord.z ) )
				) * 0.2;
			}
			return mix( 1.0, shadow, shadowIntensity );
		}
	#elif defined( SHADOWMAP_TYPE_VSM )
		float getShadow( sampler2D shadowMap, vec2 shadowMapSize, float shadowIntensity, float shadowBias, float shadowRadius, vec4 shadowCoord ) {
			float shadow = 1.0;
			shadowCoord.xyz /= shadowCoord.w;
			#ifdef USE_REVERSED_DEPTH_BUFFER
				shadowCoord.z -= shadowBias;
			#else
				shadowCoord.z += shadowBias;
			#endif
			bool inFrustum = shadowCoord.x >= 0.0 && shadowCoord.x <= 1.0 && shadowCoord.y >= 0.0 && shadowCoord.y <= 1.0;
			bool frustumTest = inFrustum && shadowCoord.z <= 1.0;
			if ( frustumTest ) {
				vec2 distribution = texture2D( shadowMap, shadowCoord.xy ).rg;
				float mean = distribution.x;
				float variance = distribution.y * distribution.y;
				#ifdef USE_REVERSED_DEPTH_BUFFER
					float hard_shadow = step( mean, shadowCoord.z );
				#else
					float hard_shadow = step( shadowCoord.z, mean );
				#endif
				
				if ( hard_shadow == 1.0 ) {
					shadow = 1.0;
				} else {
					variance = max( variance, 0.0000001 );
					float d = shadowCoord.z - mean;
					float p_max = variance / ( variance + d * d );
					p_max = clamp( ( p_max - 0.3 ) / 0.65, 0.0, 1.0 );
					shadow = max( hard_shadow, p_max );
				}
			}
			return mix( 1.0, shadow, shadowIntensity );
		}
	#else
		float getShadow( sampler2D shadowMap, vec2 shadowMapSize, float shadowIntensity, float shadowBias, float shadowRadius, vec4 shadowCoord ) {
			float shadow = 1.0;
			shadowCoord.xyz /= shadowCoord.w;
			#ifdef USE_REVERSED_DEPTH_BUFFER
				shadowCoord.z -= shadowBias;
			#else
				shadowCoord.z += shadowBias;
			#endif
			bool inFrustum = shadowCoord.x >= 0.0 && shadowCoord.x <= 1.0 && shadowCoord.y >= 0.0 && shadowCoord.y <= 1.0;
			bool frustumTest = inFrustum && shadowCoord.z <= 1.0;
			if ( frustumTest ) {
				float depth = texture2D( shadowMap, shadowCoord.xy ).r;
				#ifdef USE_REVERSED_DEPTH_BUFFER
					shadow = step( depth, shadowCoord.z );
				#else
					shadow = step( shadowCoord.z, depth );
				#endif
			}
			return mix( 1.0, shadow, shadowIntensity );
		}
	#endif
	#if NUM_POINT_LIGHT_SHADOWS > 0
	#if defined( SHADOWMAP_TYPE_PCF )
	float getPointShadow( samplerCubeShadow shadowMap, vec2 shadowMapSize, float shadowIntensity, float shadowBias, float shadowRadius, vec4 shadowCoord, float shadowCameraNear, float shadowCameraFar ) {
		float shadow = 1.0;
		vec3 lightToPosition = shadowCoord.xyz;
		vec3 bd3D = normalize( lightToPosition );
		vec3 absVec = abs( lightToPosition );
		float viewSpaceZ = max( max( absVec.x, absVec.y ), absVec.z );
		if ( viewSpaceZ - shadowCameraFar <= 0.0 && viewSpaceZ - shadowCameraNear >= 0.0 ) {
			#ifdef USE_REVERSED_DEPTH_BUFFER
				float dp = ( shadowCameraNear * ( shadowCameraFar - viewSpaceZ ) ) / ( viewSpaceZ * ( shadowCameraFar - shadowCameraNear ) );
				dp -= shadowBias;
			#else
				float dp = ( shadowCameraFar * ( viewSpaceZ - shadowCameraNear ) ) / ( viewSpaceZ * ( shadowCameraFar - shadowCameraNear ) );
				dp += shadowBias;
			#endif
			float texelSize = shadowRadius / shadowMapSize.x;
			vec3 absDir = abs( bd3D );
			vec3 tangent = absDir.x > absDir.z ? vec3( 0.0, 1.0, 0.0 ) : vec3( 1.0, 0.0, 0.0 );
			tangent = normalize( cross( bd3D, tangent ) );
			vec3 bitangent = cross( bd3D, tangent );
			float phi = interleavedGradientNoise( gl_FragCoord.xy ) * PI2;
			vec2 sample0 = vogelDiskSample( 0, 5, phi );
			vec2 sample1 = vogelDiskSample( 1, 5, phi );
			vec2 sample2 = vogelDiskSample( 2, 5, phi );
			vec2 sample3 = vogelDiskSample( 3, 5, phi );
			vec2 sample4 = vogelDiskSample( 4, 5, phi );
			shadow = (
				texture( shadowMap, vec4( bd3D + ( tangent * sample0.x + bitangent * sample0.y ) * texelSize, dp ) ) +
				texture( shadowMap, vec4( bd3D + ( tangent * sample1.x + bitangent * sample1.y ) * texelSize, dp ) ) +
				texture( shadowMap, vec4( bd3D + ( tangent * sample2.x + bitangent * sample2.y ) * texelSize, dp ) ) +
				texture( shadowMap, vec4( bd3D + ( tangent * sample3.x + bitangent * sample3.y ) * texelSize, dp ) ) +
				texture( shadowMap, vec4( bd3D + ( tangent * sample4.x + bitangent * sample4.y ) * texelSize, dp ) )
			) * 0.2;
		}
		return mix( 1.0, shadow, shadowIntensity );
	}
	#elif defined( SHADOWMAP_TYPE_BASIC )
	float getPointShadow( samplerCube shadowMap, vec2 shadowMapSize, float shadowIntensity, float shadowBias, float shadowRadius, vec4 shadowCoord, float shadowCameraNear, float shadowCameraFar ) {
		float shadow = 1.0;
		vec3 lightToPosition = shadowCoord.xyz;
		vec3 absVec = abs( lightToPosition );
		float viewSpaceZ = max( max( absVec.x, absVec.y ), absVec.z );
		if ( viewSpaceZ - shadowCameraFar <= 0.0 && viewSpaceZ - shadowCameraNear >= 0.0 ) {
			float dp = ( shadowCameraFar * ( viewSpaceZ - shadowCameraNear ) ) / ( viewSpaceZ * ( shadowCameraFar - shadowCameraNear ) );
			dp += shadowBias;
			vec3 bd3D = normalize( lightToPosition );
			float depth = textureCube( shadowMap, bd3D ).r;
			#ifdef USE_REVERSED_DEPTH_BUFFER
				depth = 1.0 - depth;
			#endif
			shadow = step( dp, depth );
		}
		return mix( 1.0, shadow, shadowIntensity );
	}
	#endif
	#endif
#endif`,ua=`#if NUM_SPOT_LIGHT_COORDS > 0
	uniform mat4 spotLightMatrix[ NUM_SPOT_LIGHT_COORDS ];
	varying vec4 vSpotLightCoord[ NUM_SPOT_LIGHT_COORDS ];
#endif
#ifdef USE_SHADOWMAP
	#if NUM_DIR_LIGHT_SHADOWS > 0
		uniform mat4 directionalShadowMatrix[ NUM_DIR_LIGHT_SHADOWS ];
		varying vec4 vDirectionalShadowCoord[ NUM_DIR_LIGHT_SHADOWS ];
		struct DirectionalLightShadow {
			float shadowIntensity;
			float shadowBias;
			float shadowNormalBias;
			float shadowRadius;
			vec2 shadowMapSize;
		};
		uniform DirectionalLightShadow directionalLightShadows[ NUM_DIR_LIGHT_SHADOWS ];
	#endif
	#if NUM_SPOT_LIGHT_SHADOWS > 0
		struct SpotLightShadow {
			float shadowIntensity;
			float shadowBias;
			float shadowNormalBias;
			float shadowRadius;
			vec2 shadowMapSize;
		};
		uniform SpotLightShadow spotLightShadows[ NUM_SPOT_LIGHT_SHADOWS ];
	#endif
	#if NUM_POINT_LIGHT_SHADOWS > 0
		uniform mat4 pointShadowMatrix[ NUM_POINT_LIGHT_SHADOWS ];
		varying vec4 vPointShadowCoord[ NUM_POINT_LIGHT_SHADOWS ];
		struct PointLightShadow {
			float shadowIntensity;
			float shadowBias;
			float shadowNormalBias;
			float shadowRadius;
			vec2 shadowMapSize;
			float shadowCameraNear;
			float shadowCameraFar;
		};
		uniform PointLightShadow pointLightShadows[ NUM_POINT_LIGHT_SHADOWS ];
	#endif
#endif`,fa=`#if ( defined( USE_SHADOWMAP ) && ( NUM_DIR_LIGHT_SHADOWS > 0 || NUM_POINT_LIGHT_SHADOWS > 0 ) ) || ( NUM_SPOT_LIGHT_COORDS > 0 )
	vec3 shadowWorldNormal = inverseTransformDirection( transformedNormal, viewMatrix );
	vec4 shadowWorldPosition;
#endif
#if defined( USE_SHADOWMAP )
	#if NUM_DIR_LIGHT_SHADOWS > 0
		#pragma unroll_loop_start
		for ( int i = 0; i < NUM_DIR_LIGHT_SHADOWS; i ++ ) {
			shadowWorldPosition = worldPosition + vec4( shadowWorldNormal * directionalLightShadows[ i ].shadowNormalBias, 0 );
			vDirectionalShadowCoord[ i ] = directionalShadowMatrix[ i ] * shadowWorldPosition;
		}
		#pragma unroll_loop_end
	#endif
	#if NUM_POINT_LIGHT_SHADOWS > 0
		#pragma unroll_loop_start
		for ( int i = 0; i < NUM_POINT_LIGHT_SHADOWS; i ++ ) {
			shadowWorldPosition = worldPosition + vec4( shadowWorldNormal * pointLightShadows[ i ].shadowNormalBias, 0 );
			vPointShadowCoord[ i ] = pointShadowMatrix[ i ] * shadowWorldPosition;
		}
		#pragma unroll_loop_end
	#endif
#endif
#if NUM_SPOT_LIGHT_COORDS > 0
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_SPOT_LIGHT_COORDS; i ++ ) {
		shadowWorldPosition = worldPosition;
		#if ( defined( USE_SHADOWMAP ) && UNROLLED_LOOP_INDEX < NUM_SPOT_LIGHT_SHADOWS )
			shadowWorldPosition.xyz += shadowWorldNormal * spotLightShadows[ i ].shadowNormalBias;
		#endif
		vSpotLightCoord[ i ] = spotLightMatrix[ i ] * shadowWorldPosition;
	}
	#pragma unroll_loop_end
#endif`,Ci=`float getShadowMask() {
	float shadow = 1.0;
	#ifdef USE_SHADOWMAP
	#if NUM_DIR_LIGHT_SHADOWS > 0
	DirectionalLightShadow directionalLight;
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_DIR_LIGHT_SHADOWS; i ++ ) {
		directionalLight = directionalLightShadows[ i ];
		shadow *= receiveShadow ? getShadow( directionalShadowMap[ i ], directionalLight.shadowMapSize, directionalLight.shadowIntensity, directionalLight.shadowBias, directionalLight.shadowRadius, vDirectionalShadowCoord[ i ] ) : 1.0;
	}
	#pragma unroll_loop_end
	#endif
	#if NUM_SPOT_LIGHT_SHADOWS > 0
	SpotLightShadow spotLight;
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_SPOT_LIGHT_SHADOWS; i ++ ) {
		spotLight = spotLightShadows[ i ];
		shadow *= receiveShadow ? getShadow( spotShadowMap[ i ], spotLight.shadowMapSize, spotLight.shadowIntensity, spotLight.shadowBias, spotLight.shadowRadius, vSpotLightCoord[ i ] ) : 1.0;
	}
	#pragma unroll_loop_end
	#endif
	#if NUM_POINT_LIGHT_SHADOWS > 0 && ( defined( SHADOWMAP_TYPE_PCF ) || defined( SHADOWMAP_TYPE_BASIC ) )
	PointLightShadow pointLight;
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_POINT_LIGHT_SHADOWS; i ++ ) {
		pointLight = pointLightShadows[ i ];
		shadow *= receiveShadow ? getPointShadow( pointShadowMap[ i ], pointLight.shadowMapSize, pointLight.shadowIntensity, pointLight.shadowBias, pointLight.shadowRadius, vPointShadowCoord[ i ], pointLight.shadowCameraNear, pointLight.shadowCameraFar ) : 1.0;
	}
	#pragma unroll_loop_end
	#endif
	#endif
	return shadow;
}`,Ps=`#ifdef USE_SKINNING
	mat4 boneMatX = getBoneMatrix( skinIndex.x );
	mat4 boneMatY = getBoneMatrix( skinIndex.y );
	mat4 boneMatZ = getBoneMatrix( skinIndex.z );
	mat4 boneMatW = getBoneMatrix( skinIndex.w );
#endif`,da=`#ifdef USE_SKINNING
	uniform mat4 bindMatrix;
	uniform mat4 bindMatrixInverse;
	uniform highp sampler2D boneTexture;
	mat4 getBoneMatrix( const in float i ) {
		int size = textureSize( boneTexture, 0 ).x;
		int j = int( i ) * 4;
		int x = j % size;
		int y = j / size;
		vec4 v1 = texelFetch( boneTexture, ivec2( x, y ), 0 );
		vec4 v2 = texelFetch( boneTexture, ivec2( x + 1, y ), 0 );
		vec4 v3 = texelFetch( boneTexture, ivec2( x + 2, y ), 0 );
		vec4 v4 = texelFetch( boneTexture, ivec2( x + 3, y ), 0 );
		return mat4( v1, v2, v3, v4 );
	}
#endif`,cr=`#ifdef USE_SKINNING
	vec4 skinVertex = bindMatrix * vec4( transformed, 1.0 );
	vec4 skinned = vec4( 0.0 );
	skinned += boneMatX * skinVertex * skinWeight.x;
	skinned += boneMatY * skinVertex * skinWeight.y;
	skinned += boneMatZ * skinVertex * skinWeight.z;
	skinned += boneMatW * skinVertex * skinWeight.w;
	transformed = ( bindMatrixInverse * skinned ).xyz;
#endif`,hr=`#ifdef USE_SKINNING
	mat4 skinMatrix = mat4( 0.0 );
	skinMatrix += skinWeight.x * boneMatX;
	skinMatrix += skinWeight.y * boneMatY;
	skinMatrix += skinWeight.z * boneMatZ;
	skinMatrix += skinWeight.w * boneMatW;
	skinMatrix = bindMatrixInverse * skinMatrix * bindMatrix;
	objectNormal = vec4( skinMatrix * vec4( objectNormal, 0.0 ) ).xyz;
	#ifdef USE_TANGENT
		objectTangent = vec4( skinMatrix * vec4( objectTangent, 0.0 ) ).xyz;
	#endif
#endif`,pa=`float specularStrength;
#ifdef USE_SPECULARMAP
	vec4 texelSpecular = texture2D( specularMap, vSpecularMapUv );
	specularStrength = texelSpecular.r;
#else
	specularStrength = 1.0;
#endif`,ma=`#ifdef USE_SPECULARMAP
	uniform sampler2D specularMap;
#endif`,Lh=`#if defined( TONE_MAPPING )
	gl_FragColor.rgb = toneMapping( gl_FragColor.rgb );
#endif`,ga=`#ifndef saturate
#define saturate( a ) clamp( a, 0.0, 1.0 )
#endif
uniform float toneMappingExposure;
vec3 LinearToneMapping( vec3 color ) {
	return saturate( toneMappingExposure * color );
}
vec3 ReinhardToneMapping( vec3 color ) {
	color *= toneMappingExposure;
	return saturate( color / ( vec3( 1.0 ) + color ) );
}
vec3 CineonToneMapping( vec3 color ) {
	color *= toneMappingExposure;
	color = max( vec3( 0.0 ), color - 0.004 );
	return pow( ( color * ( 6.2 * color + 0.5 ) ) / ( color * ( 6.2 * color + 1.7 ) + 0.06 ), vec3( 2.2 ) );
}
vec3 RRTAndODTFit( vec3 v ) {
	vec3 a = v * ( v + 0.0245786 ) - 0.000090537;
	vec3 b = v * ( 0.983729 * v + 0.4329510 ) + 0.238081;
	return a / b;
}
vec3 ACESFilmicToneMapping( vec3 color ) {
	const mat3 ACESInputMat = mat3(
		vec3( 0.59719, 0.07600, 0.02840 ),		vec3( 0.35458, 0.90834, 0.13383 ),
		vec3( 0.04823, 0.01566, 0.83777 )
	);
	const mat3 ACESOutputMat = mat3(
		vec3(  1.60475, -0.10208, -0.00327 ),		vec3( -0.53108,  1.10813, -0.07276 ),
		vec3( -0.07367, -0.00605,  1.07602 )
	);
	color *= toneMappingExposure / 0.6;
	color = ACESInputMat * color;
	color = RRTAndODTFit( color );
	color = ACESOutputMat * color;
	return saturate( color );
}
const mat3 LINEAR_REC2020_TO_LINEAR_SRGB = mat3(
	vec3( 1.6605, - 0.1246, - 0.0182 ),
	vec3( - 0.5876, 1.1329, - 0.1006 ),
	vec3( - 0.0728, - 0.0083, 1.1187 )
);
const mat3 LINEAR_SRGB_TO_LINEAR_REC2020 = mat3(
	vec3( 0.6274, 0.0691, 0.0164 ),
	vec3( 0.3293, 0.9195, 0.0880 ),
	vec3( 0.0433, 0.0113, 0.8956 )
);
vec3 agxDefaultContrastApprox( vec3 x ) {
	vec3 x2 = x * x;
	vec3 x4 = x2 * x2;
	return + 15.5 * x4 * x2
		- 40.14 * x4 * x
		+ 31.96 * x4
		- 6.868 * x2 * x
		+ 0.4298 * x2
		+ 0.1191 * x
		- 0.00232;
}
vec3 AgXToneMapping( vec3 color ) {
	const mat3 AgXInsetMatrix = mat3(
		vec3( 0.856627153315983, 0.137318972929847, 0.11189821299995 ),
		vec3( 0.0951212405381588, 0.761241990602591, 0.0767994186031903 ),
		vec3( 0.0482516061458583, 0.101439036467562, 0.811302368396859 )
	);
	const mat3 AgXOutsetMatrix = mat3(
		vec3( 1.1271005818144368, - 0.1413297634984383, - 0.14132976349843826 ),
		vec3( - 0.11060664309660323, 1.157823702216272, - 0.11060664309660294 ),
		vec3( - 0.016493938717834573, - 0.016493938717834257, 1.2519364065950405 )
	);
	const float AgxMinEv = - 12.47393;	const float AgxMaxEv = 4.026069;
	color *= toneMappingExposure;
	color = LINEAR_SRGB_TO_LINEAR_REC2020 * color;
	color = AgXInsetMatrix * color;
	color = max( color, 1e-10 );	color = log2( color );
	color = ( color - AgxMinEv ) / ( AgxMaxEv - AgxMinEv );
	color = clamp( color, 0.0, 1.0 );
	color = agxDefaultContrastApprox( color );
	color = AgXOutsetMatrix * color;
	color = pow( max( vec3( 0.0 ), color ), vec3( 2.2 ) );
	color = LINEAR_REC2020_TO_LINEAR_SRGB * color;
	color = clamp( color, 0.0, 1.0 );
	return color;
}
vec3 NeutralToneMapping( vec3 color ) {
	const float StartCompression = 0.8 - 0.04;
	const float Desaturation = 0.15;
	color *= toneMappingExposure;
	float x = min( color.r, min( color.g, color.b ) );
	float offset = x < 0.08 ? x - 6.25 * x * x : 0.04;
	color -= offset;
	float peak = max( color.r, max( color.g, color.b ) );
	if ( peak < StartCompression ) return color;
	float d = 1. - StartCompression;
	float newPeak = 1. - d * d / ( peak + d - StartCompression );
	color *= newPeak / peak;
	float g = 1. - 1. / ( Desaturation * ( peak - newPeak ) + 1. );
	return mix( color, vec3( newPeak ), g );
}
vec3 CustomToneMapping( vec3 color ) { return color; }`,_a=`#ifdef USE_TRANSMISSION
	material.transmission = transmission;
	material.transmissionAlpha = 1.0;
	material.thickness = thickness;
	material.attenuationDistance = attenuationDistance;
	material.attenuationColor = attenuationColor;
	#ifdef USE_TRANSMISSIONMAP
		material.transmission *= texture2D( transmissionMap, vTransmissionMapUv ).r;
	#endif
	#ifdef USE_THICKNESSMAP
		material.thickness *= texture2D( thicknessMap, vThicknessMapUv ).g;
	#endif
	vec3 pos = vWorldPosition;
	vec3 v = normalize( cameraPosition - pos );
	vec3 n = inverseTransformDirection( normal, viewMatrix );
	vec4 transmitted = getIBLVolumeRefraction(
		n, v, material.roughness, material.diffuseContribution, material.specularColorBlended, material.specularF90,
		pos, modelMatrix, viewMatrix, projectionMatrix, material.dispersion, material.ior, material.thickness,
		material.attenuationColor, material.attenuationDistance );
	material.transmissionAlpha = mix( material.transmissionAlpha, transmitted.a, material.transmission );
	totalDiffuse = mix( totalDiffuse, transmitted.rgb, material.transmission );
#endif`,xa=`#ifdef USE_TRANSMISSION
	uniform float transmission;
	uniform float thickness;
	uniform float attenuationDistance;
	uniform vec3 attenuationColor;
	#ifdef USE_TRANSMISSIONMAP
		uniform sampler2D transmissionMap;
	#endif
	#ifdef USE_THICKNESSMAP
		uniform sampler2D thicknessMap;
	#endif
	uniform vec2 transmissionSamplerSize;
	uniform sampler2D transmissionSamplerMap;
	uniform mat4 modelMatrix;
	uniform mat4 projectionMatrix;
	varying vec3 vWorldPosition;
	float w0( float a ) {
		return ( 1.0 / 6.0 ) * ( a * ( a * ( - a + 3.0 ) - 3.0 ) + 1.0 );
	}
	float w1( float a ) {
		return ( 1.0 / 6.0 ) * ( a *  a * ( 3.0 * a - 6.0 ) + 4.0 );
	}
	float w2( float a ){
		return ( 1.0 / 6.0 ) * ( a * ( a * ( - 3.0 * a + 3.0 ) + 3.0 ) + 1.0 );
	}
	float w3( float a ) {
		return ( 1.0 / 6.0 ) * ( a * a * a );
	}
	float g0( float a ) {
		return w0( a ) + w1( a );
	}
	float g1( float a ) {
		return w2( a ) + w3( a );
	}
	float h0( float a ) {
		return - 1.0 + w1( a ) / ( w0( a ) + w1( a ) );
	}
	float h1( float a ) {
		return 1.0 + w3( a ) / ( w2( a ) + w3( a ) );
	}
	vec4 bicubic( sampler2D tex, vec2 uv, vec4 texelSize, float lod ) {
		uv = uv * texelSize.zw + 0.5;
		vec2 iuv = floor( uv );
		vec2 fuv = fract( uv );
		float g0x = g0( fuv.x );
		float g1x = g1( fuv.x );
		float h0x = h0( fuv.x );
		float h1x = h1( fuv.x );
		float h0y = h0( fuv.y );
		float h1y = h1( fuv.y );
		vec2 p0 = ( vec2( iuv.x + h0x, iuv.y + h0y ) - 0.5 ) * texelSize.xy;
		vec2 p1 = ( vec2( iuv.x + h1x, iuv.y + h0y ) - 0.5 ) * texelSize.xy;
		vec2 p2 = ( vec2( iuv.x + h0x, iuv.y + h1y ) - 0.5 ) * texelSize.xy;
		vec2 p3 = ( vec2( iuv.x + h1x, iuv.y + h1y ) - 0.5 ) * texelSize.xy;
		return g0( fuv.y ) * ( g0x * textureLod( tex, p0, lod ) + g1x * textureLod( tex, p1, lod ) ) +
			g1( fuv.y ) * ( g0x * textureLod( tex, p2, lod ) + g1x * textureLod( tex, p3, lod ) );
	}
	vec4 textureBicubic( sampler2D sampler, vec2 uv, float lod ) {
		vec2 fLodSize = vec2( textureSize( sampler, int( lod ) ) );
		vec2 cLodSize = vec2( textureSize( sampler, int( lod + 1.0 ) ) );
		vec2 fLodSizeInv = 1.0 / fLodSize;
		vec2 cLodSizeInv = 1.0 / cLodSize;
		vec4 fSample = bicubic( sampler, uv, vec4( fLodSizeInv, fLodSize ), floor( lod ) );
		vec4 cSample = bicubic( sampler, uv, vec4( cLodSizeInv, cLodSize ), ceil( lod ) );
		return mix( fSample, cSample, fract( lod ) );
	}
	vec3 getVolumeTransmissionRay( const in vec3 n, const in vec3 v, const in float thickness, const in float ior, const in mat4 modelMatrix ) {
		vec3 refractionVector = refract( - v, normalize( n ), 1.0 / ior );
		vec3 modelScale;
		modelScale.x = length( vec3( modelMatrix[ 0 ].xyz ) );
		modelScale.y = length( vec3( modelMatrix[ 1 ].xyz ) );
		modelScale.z = length( vec3( modelMatrix[ 2 ].xyz ) );
		return normalize( refractionVector ) * thickness * modelScale;
	}
	float applyIorToRoughness( const in float roughness, const in float ior ) {
		return roughness * clamp( ior * 2.0 - 2.0, 0.0, 1.0 );
	}
	vec4 getTransmissionSample( const in vec2 fragCoord, const in float roughness, const in float ior ) {
		float lod = log2( transmissionSamplerSize.x ) * applyIorToRoughness( roughness, ior );
		return textureBicubic( transmissionSamplerMap, fragCoord.xy, lod );
	}
	vec3 volumeAttenuation( const in float transmissionDistance, const in vec3 attenuationColor, const in float attenuationDistance ) {
		if ( isinf( attenuationDistance ) ) {
			return vec3( 1.0 );
		} else {
			vec3 attenuationCoefficient = -log( attenuationColor ) / attenuationDistance;
			vec3 transmittance = exp( - attenuationCoefficient * transmissionDistance );			return transmittance;
		}
	}
	vec4 getIBLVolumeRefraction( const in vec3 n, const in vec3 v, const in float roughness, const in vec3 diffuseColor,
		const in vec3 specularColor, const in float specularF90, const in vec3 position, const in mat4 modelMatrix,
		const in mat4 viewMatrix, const in mat4 projMatrix, const in float dispersion, const in float ior, const in float thickness,
		const in vec3 attenuationColor, const in float attenuationDistance ) {
		vec4 transmittedLight;
		vec3 transmittance;
		#ifdef USE_DISPERSION
			float halfSpread = ( ior - 1.0 ) * 0.025 * dispersion;
			vec3 iors = vec3( ior - halfSpread, ior, ior + halfSpread );
			for ( int i = 0; i < 3; i ++ ) {
				vec3 transmissionRay = getVolumeTransmissionRay( n, v, thickness, iors[ i ], modelMatrix );
				vec3 refractedRayExit = position + transmissionRay;
				vec4 ndcPos = projMatrix * viewMatrix * vec4( refractedRayExit, 1.0 );
				vec2 refractionCoords = ndcPos.xy / ndcPos.w;
				refractionCoords += 1.0;
				refractionCoords /= 2.0;
				vec4 transmissionSample = getTransmissionSample( refractionCoords, roughness, iors[ i ] );
				transmittedLight[ i ] = transmissionSample[ i ];
				transmittedLight.a += transmissionSample.a;
				transmittance[ i ] = diffuseColor[ i ] * volumeAttenuation( length( transmissionRay ), attenuationColor, attenuationDistance )[ i ];
			}
			transmittedLight.a /= 3.0;
		#else
			vec3 transmissionRay = getVolumeTransmissionRay( n, v, thickness, ior, modelMatrix );
			vec3 refractedRayExit = position + transmissionRay;
			vec4 ndcPos = projMatrix * viewMatrix * vec4( refractedRayExit, 1.0 );
			vec2 refractionCoords = ndcPos.xy / ndcPos.w;
			refractionCoords += 1.0;
			refractionCoords /= 2.0;
			transmittedLight = getTransmissionSample( refractionCoords, roughness, ior );
			transmittance = diffuseColor * volumeAttenuation( length( transmissionRay ), attenuationColor, attenuationDistance );
		#endif
		vec3 attenuatedColor = transmittance * transmittedLight.rgb;
		vec3 F = EnvironmentBRDF( n, v, specularColor, specularF90, roughness );
		float transmittanceFactor = ( transmittance.r + transmittance.g + transmittance.b ) / 3.0;
		return vec4( ( 1.0 - F ) * attenuatedColor, 1.0 - ( 1.0 - transmittedLight.a ) * transmittanceFactor );
	}
#endif`,va=`#if defined( USE_UV ) || defined( USE_ANISOTROPY )
	varying vec2 vUv;
#endif
#ifdef USE_MAP
	varying vec2 vMapUv;
#endif
#ifdef USE_ALPHAMAP
	varying vec2 vAlphaMapUv;
#endif
#ifdef USE_LIGHTMAP
	varying vec2 vLightMapUv;
#endif
#ifdef USE_AOMAP
	varying vec2 vAoMapUv;
#endif
#ifdef USE_BUMPMAP
	varying vec2 vBumpMapUv;
#endif
#ifdef USE_NORMALMAP
	varying vec2 vNormalMapUv;
#endif
#ifdef USE_EMISSIVEMAP
	varying vec2 vEmissiveMapUv;
#endif
#ifdef USE_METALNESSMAP
	varying vec2 vMetalnessMapUv;
#endif
#ifdef USE_ROUGHNESSMAP
	varying vec2 vRoughnessMapUv;
#endif
#ifdef USE_ANISOTROPYMAP
	varying vec2 vAnisotropyMapUv;
#endif
#ifdef USE_CLEARCOATMAP
	varying vec2 vClearcoatMapUv;
#endif
#ifdef USE_CLEARCOAT_NORMALMAP
	varying vec2 vClearcoatNormalMapUv;
#endif
#ifdef USE_CLEARCOAT_ROUGHNESSMAP
	varying vec2 vClearcoatRoughnessMapUv;
#endif
#ifdef USE_IRIDESCENCEMAP
	varying vec2 vIridescenceMapUv;
#endif
#ifdef USE_IRIDESCENCE_THICKNESSMAP
	varying vec2 vIridescenceThicknessMapUv;
#endif
#ifdef USE_SHEEN_COLORMAP
	varying vec2 vSheenColorMapUv;
#endif
#ifdef USE_SHEEN_ROUGHNESSMAP
	varying vec2 vSheenRoughnessMapUv;
#endif
#ifdef USE_SPECULARMAP
	varying vec2 vSpecularMapUv;
#endif
#ifdef USE_SPECULAR_COLORMAP
	varying vec2 vSpecularColorMapUv;
#endif
#ifdef USE_SPECULAR_INTENSITYMAP
	varying vec2 vSpecularIntensityMapUv;
#endif
#ifdef USE_TRANSMISSIONMAP
	uniform mat3 transmissionMapTransform;
	varying vec2 vTransmissionMapUv;
#endif
#ifdef USE_THICKNESSMAP
	uniform mat3 thicknessMapTransform;
	varying vec2 vThicknessMapUv;
#endif`,ya=`#if defined( USE_UV ) || defined( USE_ANISOTROPY )
	varying vec2 vUv;
#endif
#ifdef USE_MAP
	uniform mat3 mapTransform;
	varying vec2 vMapUv;
#endif
#ifdef USE_ALPHAMAP
	uniform mat3 alphaMapTransform;
	varying vec2 vAlphaMapUv;
#endif
#ifdef USE_LIGHTMAP
	uniform mat3 lightMapTransform;
	varying vec2 vLightMapUv;
#endif
#ifdef USE_AOMAP
	uniform mat3 aoMapTransform;
	varying vec2 vAoMapUv;
#endif
#ifdef USE_BUMPMAP
	uniform mat3 bumpMapTransform;
	varying vec2 vBumpMapUv;
#endif
#ifdef USE_NORMALMAP
	uniform mat3 normalMapTransform;
	varying vec2 vNormalMapUv;
#endif
#ifdef USE_DISPLACEMENTMAP
	uniform mat3 displacementMapTransform;
	varying vec2 vDisplacementMapUv;
#endif
#ifdef USE_EMISSIVEMAP
	uniform mat3 emissiveMapTransform;
	varying vec2 vEmissiveMapUv;
#endif
#ifdef USE_METALNESSMAP
	uniform mat3 metalnessMapTransform;
	varying vec2 vMetalnessMapUv;
#endif
#ifdef USE_ROUGHNESSMAP
	uniform mat3 roughnessMapTransform;
	varying vec2 vRoughnessMapUv;
#endif
#ifdef USE_ANISOTROPYMAP
	uniform mat3 anisotropyMapTransform;
	varying vec2 vAnisotropyMapUv;
#endif
#ifdef USE_CLEARCOATMAP
	uniform mat3 clearcoatMapTransform;
	varying vec2 vClearcoatMapUv;
#endif
#ifdef USE_CLEARCOAT_NORMALMAP
	uniform mat3 clearcoatNormalMapTransform;
	varying vec2 vClearcoatNormalMapUv;
#endif
#ifdef USE_CLEARCOAT_ROUGHNESSMAP
	uniform mat3 clearcoatRoughnessMapTransform;
	varying vec2 vClearcoatRoughnessMapUv;
#endif
#ifdef USE_SHEEN_COLORMAP
	uniform mat3 sheenColorMapTransform;
	varying vec2 vSheenColorMapUv;
#endif
#ifdef USE_SHEEN_ROUGHNESSMAP
	uniform mat3 sheenRoughnessMapTransform;
	varying vec2 vSheenRoughnessMapUv;
#endif
#ifdef USE_IRIDESCENCEMAP
	uniform mat3 iridescenceMapTransform;
	varying vec2 vIridescenceMapUv;
#endif
#ifdef USE_IRIDESCENCE_THICKNESSMAP
	uniform mat3 iridescenceThicknessMapTransform;
	varying vec2 vIridescenceThicknessMapUv;
#endif
#ifdef USE_SPECULARMAP
	uniform mat3 specularMapTransform;
	varying vec2 vSpecularMapUv;
#endif
#ifdef USE_SPECULAR_COLORMAP
	uniform mat3 specularColorMapTransform;
	varying vec2 vSpecularColorMapUv;
#endif
#ifdef USE_SPECULAR_INTENSITYMAP
	uniform mat3 specularIntensityMapTransform;
	varying vec2 vSpecularIntensityMapUv;
#endif
#ifdef USE_TRANSMISSIONMAP
	uniform mat3 transmissionMapTransform;
	varying vec2 vTransmissionMapUv;
#endif
#ifdef USE_THICKNESSMAP
	uniform mat3 thicknessMapTransform;
	varying vec2 vThicknessMapUv;
#endif`,Ma=`#if defined( USE_UV ) || defined( USE_ANISOTROPY )
	vUv = vec3( uv, 1 ).xy;
#endif
#ifdef USE_MAP
	vMapUv = ( mapTransform * vec3( MAP_UV, 1 ) ).xy;
#endif
#ifdef USE_ALPHAMAP
	vAlphaMapUv = ( alphaMapTransform * vec3( ALPHAMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_LIGHTMAP
	vLightMapUv = ( lightMapTransform * vec3( LIGHTMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_AOMAP
	vAoMapUv = ( aoMapTransform * vec3( AOMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_BUMPMAP
	vBumpMapUv = ( bumpMapTransform * vec3( BUMPMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_NORMALMAP
	vNormalMapUv = ( normalMapTransform * vec3( NORMALMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_DISPLACEMENTMAP
	vDisplacementMapUv = ( displacementMapTransform * vec3( DISPLACEMENTMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_EMISSIVEMAP
	vEmissiveMapUv = ( emissiveMapTransform * vec3( EMISSIVEMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_METALNESSMAP
	vMetalnessMapUv = ( metalnessMapTransform * vec3( METALNESSMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_ROUGHNESSMAP
	vRoughnessMapUv = ( roughnessMapTransform * vec3( ROUGHNESSMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_ANISOTROPYMAP
	vAnisotropyMapUv = ( anisotropyMapTransform * vec3( ANISOTROPYMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_CLEARCOATMAP
	vClearcoatMapUv = ( clearcoatMapTransform * vec3( CLEARCOATMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_CLEARCOAT_NORMALMAP
	vClearcoatNormalMapUv = ( clearcoatNormalMapTransform * vec3( CLEARCOAT_NORMALMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_CLEARCOAT_ROUGHNESSMAP
	vClearcoatRoughnessMapUv = ( clearcoatRoughnessMapTransform * vec3( CLEARCOAT_ROUGHNESSMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_IRIDESCENCEMAP
	vIridescenceMapUv = ( iridescenceMapTransform * vec3( IRIDESCENCEMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_IRIDESCENCE_THICKNESSMAP
	vIridescenceThicknessMapUv = ( iridescenceThicknessMapTransform * vec3( IRIDESCENCE_THICKNESSMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_SHEEN_COLORMAP
	vSheenColorMapUv = ( sheenColorMapTransform * vec3( SHEEN_COLORMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_SHEEN_ROUGHNESSMAP
	vSheenRoughnessMapUv = ( sheenRoughnessMapTransform * vec3( SHEEN_ROUGHNESSMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_SPECULARMAP
	vSpecularMapUv = ( specularMapTransform * vec3( SPECULARMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_SPECULAR_COLORMAP
	vSpecularColorMapUv = ( specularColorMapTransform * vec3( SPECULAR_COLORMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_SPECULAR_INTENSITYMAP
	vSpecularIntensityMapUv = ( specularIntensityMapTransform * vec3( SPECULAR_INTENSITYMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_TRANSMISSIONMAP
	vTransmissionMapUv = ( transmissionMapTransform * vec3( TRANSMISSIONMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_THICKNESSMAP
	vThicknessMapUv = ( thicknessMapTransform * vec3( THICKNESSMAP_UV, 1 ) ).xy;
#endif`,Sa=`#if defined( USE_ENVMAP ) || defined( DISTANCE ) || defined ( USE_SHADOWMAP ) || defined ( USE_TRANSMISSION ) || NUM_SPOT_LIGHT_COORDS > 0
	vec4 worldPosition = vec4( transformed, 1.0 );
	#ifdef USE_BATCHING
		worldPosition = batchingMatrix * worldPosition;
	#endif
	#ifdef USE_INSTANCING
		worldPosition = instanceMatrix * worldPosition;
	#endif
	worldPosition = modelMatrix * worldPosition;
#endif`;const ve={alphahash_fragment:zo,alphahash_pars_fragment:Vo,alphamap_fragment:Ah,alphamap_pars_fragment:Eh,alphatest_fragment:ko,alphatest_pars_fragment:Go,aomap_fragment:Ho,aomap_pars_fragment:As,batching_pars_vertex:Yr,batching_vertex:Wo,begin_vertex:Xo,beginnormal_vertex:Ks,bsdfs:qo,iridescence_fragment:Yo,bumpmap_pars_fragment:Zo,clipping_planes_fragment:$o,clipping_planes_pars_fragment:wh,clipping_planes_pars_vertex:Qs,clipping_planes_vertex:Jo,color_fragment:Ko,color_pars_fragment:Qo,color_pars_vertex:jo,color_vertex:tl,common:el,cube_uv_reflection_fragment:nl,defaultnormal_vertex:il,displacementmap_pars_vertex:js,displacementmap_vertex:tr,emissivemap_fragment:sl,emissivemap_pars_fragment:rl,colorspace_fragment:al,colorspace_pars_fragment:ol,envmap_fragment:ll,envmap_common_pars_fragment:cl,envmap_pars_fragment:hl,envmap_pars_vertex:ul,envmap_physical_pars_fragment:dl,envmap_vertex:fl,fog_vertex:er,fog_pars_vertex:nr,fog_fragment:ir,fog_pars_fragment:qi,gradientmap_pars_fragment:sr,lightmap_pars_fragment:rr,lights_lambert_fragment:ar,lights_lambert_pars_fragment:or,lights_pars_begin:Es,lights_toon_fragment:pl,lights_toon_pars_fragment:ml,lights_phong_fragment:gl,lights_phong_pars_fragment:_l,lights_physical_fragment:xl,lights_physical_pars_fragment:vl,lights_fragment_begin:yl,lights_fragment_maps:Ml,lights_fragment_end:Sl,logdepthbuf_fragment:Zr,logdepthbuf_pars_fragment:bl,logdepthbuf_pars_vertex:lr,logdepthbuf_vertex:Yi,map_fragment:$r,map_pars_fragment:Jr,map_particle_fragment:Kr,map_particle_pars_fragment:Qr,metalnessmap_fragment:ws,metalnessmap_pars_fragment:Ln,morphinstance_vertex:Cs,morphcolor_vertex:Mn,morphnormal_vertex:jr,morphtarget_pars_vertex:Ch,morphtarget_vertex:ta,normal_fragment_begin:Rh,normal_fragment_maps:Cn,normal_pars_fragment:ea,normal_pars_vertex:Ih,normal_vertex:Rs,normalmap_pars_fragment:Ph,clearcoat_normal_fragment_begin:Is,clearcoat_normal_fragment_maps:na,clearcoat_pars_fragment:ia,iridescence_pars_fragment:sa,opaque_fragment:ra,packing:Zi,premultiplied_alpha_fragment:gi,project_vertex:aa,dithering_fragment:oa,dithering_pars_fragment:la,roughnessmap_fragment:Tl,roughnessmap_pars_fragment:ca,shadowmap_pars_fragment:ha,shadowmap_pars_vertex:ua,shadowmap_vertex:fa,shadowmask_pars_fragment:Ci,skinbase_vertex:Ps,skinning_pars_vertex:da,skinning_vertex:cr,skinnormal_vertex:hr,specularmap_fragment:pa,specularmap_pars_fragment:ma,tonemapping_fragment:Lh,tonemapping_pars_fragment:ga,transmission_fragment:_a,transmission_pars_fragment:xa,uv_pars_fragment:va,uv_pars_vertex:ya,uv_vertex:Ma,worldpos_vertex:Sa,background_vert:`varying vec2 vUv;
uniform mat3 uvTransform;
void main() {
	vUv = ( uvTransform * vec3( uv, 1 ) ).xy;
	gl_Position = vec4( position.xy, 1.0, 1.0 );
}`,background_frag:`uniform sampler2D t2D;
uniform float backgroundIntensity;
varying vec2 vUv;
void main() {
	vec4 texColor = texture2D( t2D, vUv );
	#ifdef DECODE_VIDEO_TEXTURE
		texColor = vec4( mix( pow( texColor.rgb * 0.9478672986 + vec3( 0.0521327014 ), vec3( 2.4 ) ), texColor.rgb * 0.0773993808, vec3( lessThanEqual( texColor.rgb, vec3( 0.04045 ) ) ) ), texColor.w );
	#endif
	texColor.rgb *= backgroundIntensity;
	gl_FragColor = texColor;
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
}`,backgroundCube_vert:`varying vec3 vWorldDirection;
#include <common>
void main() {
	vWorldDirection = transformDirection( position, modelMatrix );
	#include <begin_vertex>
	#include <project_vertex>
	gl_Position.z = gl_Position.w;
}`,backgroundCube_frag:`#ifdef ENVMAP_TYPE_CUBE
	uniform samplerCube envMap;
#elif defined( ENVMAP_TYPE_CUBE_UV )
	uniform sampler2D envMap;
#endif
uniform float flipEnvMap;
uniform float backgroundBlurriness;
uniform float backgroundIntensity;
uniform mat3 backgroundRotation;
varying vec3 vWorldDirection;
#include <cube_uv_reflection_fragment>
void main() {
	#ifdef ENVMAP_TYPE_CUBE
		vec4 texColor = textureCube( envMap, backgroundRotation * vec3( flipEnvMap * vWorldDirection.x, vWorldDirection.yz ) );
	#elif defined( ENVMAP_TYPE_CUBE_UV )
		vec4 texColor = textureCubeUV( envMap, backgroundRotation * vWorldDirection, backgroundBlurriness );
	#else
		vec4 texColor = vec4( 0.0, 0.0, 0.0, 1.0 );
	#endif
	texColor.rgb *= backgroundIntensity;
	gl_FragColor = texColor;
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
}`,cube_vert:`varying vec3 vWorldDirection;
#include <common>
void main() {
	vWorldDirection = transformDirection( position, modelMatrix );
	#include <begin_vertex>
	#include <project_vertex>
	gl_Position.z = gl_Position.w;
}`,cube_frag:`uniform samplerCube tCube;
uniform float tFlip;
uniform float opacity;
varying vec3 vWorldDirection;
void main() {
	vec4 texColor = textureCube( tCube, vec3( tFlip * vWorldDirection.x, vWorldDirection.yz ) );
	gl_FragColor = texColor;
	gl_FragColor.a *= opacity;
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
}`,depth_vert:`#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <displacementmap_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
varying vec2 vHighPrecisionZW;
void main() {
	#include <uv_vertex>
	#include <batching_vertex>
	#include <skinbase_vertex>
	#include <morphinstance_vertex>
	#ifdef USE_DISPLACEMENTMAP
		#include <beginnormal_vertex>
		#include <morphnormal_vertex>
		#include <skinnormal_vertex>
	#endif
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	vHighPrecisionZW = gl_Position.zw;
}`,depth_frag:`#if DEPTH_PACKING == 3200
	uniform float opacity;
#endif
#include <common>
#include <packing>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
varying vec2 vHighPrecisionZW;
void main() {
	vec4 diffuseColor = vec4( 1.0 );
	#include <clipping_planes_fragment>
	#if DEPTH_PACKING == 3200
		diffuseColor.a = opacity;
	#endif
	#include <map_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	#include <logdepthbuf_fragment>
	#ifdef USE_REVERSED_DEPTH_BUFFER
		float fragCoordZ = vHighPrecisionZW[ 0 ] / vHighPrecisionZW[ 1 ];
	#else
		float fragCoordZ = 0.5 * vHighPrecisionZW[ 0 ] / vHighPrecisionZW[ 1 ] + 0.5;
	#endif
	#if DEPTH_PACKING == 3200
		gl_FragColor = vec4( vec3( 1.0 - fragCoordZ ), opacity );
	#elif DEPTH_PACKING == 3201
		gl_FragColor = packDepthToRGBA( fragCoordZ );
	#elif DEPTH_PACKING == 3202
		gl_FragColor = vec4( packDepthToRGB( fragCoordZ ), 1.0 );
	#elif DEPTH_PACKING == 3203
		gl_FragColor = vec4( packDepthToRG( fragCoordZ ), 0.0, 1.0 );
	#endif
}`,distance_vert:`#define DISTANCE
varying vec3 vWorldPosition;
#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <displacementmap_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <batching_vertex>
	#include <skinbase_vertex>
	#include <morphinstance_vertex>
	#ifdef USE_DISPLACEMENTMAP
		#include <beginnormal_vertex>
		#include <morphnormal_vertex>
		#include <skinnormal_vertex>
	#endif
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <worldpos_vertex>
	#include <clipping_planes_vertex>
	vWorldPosition = worldPosition.xyz;
}`,distance_frag:`#define DISTANCE
uniform vec3 referencePosition;
uniform float nearDistance;
uniform float farDistance;
varying vec3 vWorldPosition;
#include <common>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <clipping_planes_pars_fragment>
void main () {
	vec4 diffuseColor = vec4( 1.0 );
	#include <clipping_planes_fragment>
	#include <map_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	float dist = length( vWorldPosition - referencePosition );
	dist = ( dist - nearDistance ) / ( farDistance - nearDistance );
	dist = saturate( dist );
	gl_FragColor = vec4( dist, 0.0, 0.0, 1.0 );
}`,equirect_vert:`varying vec3 vWorldDirection;
#include <common>
void main() {
	vWorldDirection = transformDirection( position, modelMatrix );
	#include <begin_vertex>
	#include <project_vertex>
}`,equirect_frag:`uniform sampler2D tEquirect;
varying vec3 vWorldDirection;
#include <common>
void main() {
	vec3 direction = normalize( vWorldDirection );
	vec2 sampleUV = equirectUv( direction );
	gl_FragColor = texture2D( tEquirect, sampleUV );
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
}`,linedashed_vert:`uniform float scale;
attribute float lineDistance;
varying float vLineDistance;
#include <common>
#include <uv_pars_vertex>
#include <color_pars_vertex>
#include <fog_pars_vertex>
#include <morphtarget_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	vLineDistance = scale * lineDistance;
	#include <uv_vertex>
	#include <color_vertex>
	#include <morphinstance_vertex>
	#include <morphcolor_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	#include <fog_vertex>
}`,linedashed_frag:`uniform vec3 diffuse;
uniform float opacity;
uniform float dashSize;
uniform float totalSize;
varying float vLineDistance;
#include <common>
#include <color_pars_fragment>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <fog_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	if ( mod( vLineDistance, totalSize ) > dashSize ) {
		discard;
	}
	vec3 outgoingLight = vec3( 0.0 );
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <color_fragment>
	outgoingLight = diffuseColor.rgb;
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
}`,meshbasic_vert:`#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <envmap_pars_vertex>
#include <color_pars_vertex>
#include <fog_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <color_vertex>
	#include <morphinstance_vertex>
	#include <morphcolor_vertex>
	#include <batching_vertex>
	#if defined ( USE_ENVMAP ) || defined ( USE_SKINNING )
		#include <beginnormal_vertex>
		#include <morphnormal_vertex>
		#include <skinbase_vertex>
		#include <skinnormal_vertex>
		#include <defaultnormal_vertex>
	#endif
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	#include <worldpos_vertex>
	#include <envmap_vertex>
	#include <fog_vertex>
}`,meshbasic_frag:`uniform vec3 diffuse;
uniform float opacity;
#ifndef FLAT_SHADED
	varying vec3 vNormal;
#endif
#include <common>
#include <dithering_pars_fragment>
#include <color_pars_fragment>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <aomap_pars_fragment>
#include <lightmap_pars_fragment>
#include <envmap_common_pars_fragment>
#include <envmap_pars_fragment>
#include <fog_pars_fragment>
#include <specularmap_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <color_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	#include <specularmap_fragment>
	ReflectedLight reflectedLight = ReflectedLight( vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ) );
	#ifdef USE_LIGHTMAP
		vec4 lightMapTexel = texture2D( lightMap, vLightMapUv );
		reflectedLight.indirectDiffuse += lightMapTexel.rgb * lightMapIntensity * RECIPROCAL_PI;
	#else
		reflectedLight.indirectDiffuse += vec3( 1.0 );
	#endif
	#include <aomap_fragment>
	reflectedLight.indirectDiffuse *= diffuseColor.rgb;
	vec3 outgoingLight = reflectedLight.indirectDiffuse;
	#include <envmap_fragment>
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
	#include <dithering_fragment>
}`,meshlambert_vert:`#define LAMBERT
varying vec3 vViewPosition;
#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <displacementmap_pars_vertex>
#include <envmap_pars_vertex>
#include <color_pars_vertex>
#include <fog_pars_vertex>
#include <normal_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <shadowmap_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <color_vertex>
	#include <morphinstance_vertex>
	#include <morphcolor_vertex>
	#include <batching_vertex>
	#include <beginnormal_vertex>
	#include <morphnormal_vertex>
	#include <skinbase_vertex>
	#include <skinnormal_vertex>
	#include <defaultnormal_vertex>
	#include <normal_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	vViewPosition = - mvPosition.xyz;
	#include <worldpos_vertex>
	#include <envmap_vertex>
	#include <shadowmap_vertex>
	#include <fog_vertex>
}`,meshlambert_frag:`#define LAMBERT
uniform vec3 diffuse;
uniform vec3 emissive;
uniform float opacity;
#include <common>
#include <dithering_pars_fragment>
#include <color_pars_fragment>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <aomap_pars_fragment>
#include <lightmap_pars_fragment>
#include <emissivemap_pars_fragment>
#include <cube_uv_reflection_fragment>
#include <envmap_common_pars_fragment>
#include <envmap_pars_fragment>
#include <envmap_physical_pars_fragment>
#include <fog_pars_fragment>
#include <bsdfs>
#include <lights_pars_begin>
#include <normal_pars_fragment>
#include <lights_lambert_pars_fragment>
#include <shadowmap_pars_fragment>
#include <bumpmap_pars_fragment>
#include <normalmap_pars_fragment>
#include <specularmap_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	ReflectedLight reflectedLight = ReflectedLight( vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ) );
	vec3 totalEmissiveRadiance = emissive;
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <color_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	#include <specularmap_fragment>
	#include <normal_fragment_begin>
	#include <normal_fragment_maps>
	#include <emissivemap_fragment>
	#include <lights_lambert_fragment>
	#include <lights_fragment_begin>
	#include <lights_fragment_maps>
	#include <lights_fragment_end>
	#include <aomap_fragment>
	vec3 outgoingLight = reflectedLight.directDiffuse + reflectedLight.indirectDiffuse + totalEmissiveRadiance;
	#include <envmap_fragment>
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
	#include <dithering_fragment>
}`,meshmatcap_vert:`#define MATCAP
varying vec3 vViewPosition;
#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <color_pars_vertex>
#include <displacementmap_pars_vertex>
#include <fog_pars_vertex>
#include <normal_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <color_vertex>
	#include <morphinstance_vertex>
	#include <morphcolor_vertex>
	#include <batching_vertex>
	#include <beginnormal_vertex>
	#include <morphnormal_vertex>
	#include <skinbase_vertex>
	#include <skinnormal_vertex>
	#include <defaultnormal_vertex>
	#include <normal_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	#include <fog_vertex>
	vViewPosition = - mvPosition.xyz;
}`,meshmatcap_frag:`#define MATCAP
uniform vec3 diffuse;
uniform float opacity;
uniform sampler2D matcap;
varying vec3 vViewPosition;
#include <common>
#include <dithering_pars_fragment>
#include <color_pars_fragment>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <fog_pars_fragment>
#include <normal_pars_fragment>
#include <bumpmap_pars_fragment>
#include <normalmap_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <color_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	#include <normal_fragment_begin>
	#include <normal_fragment_maps>
	vec3 viewDir = normalize( vViewPosition );
	vec3 x = normalize( vec3( viewDir.z, 0.0, - viewDir.x ) );
	vec3 y = cross( viewDir, x );
	vec2 uv = vec2( dot( x, normal ), dot( y, normal ) ) * 0.495 + 0.5;
	#ifdef USE_MATCAP
		vec4 matcapColor = texture2D( matcap, uv );
	#else
		vec4 matcapColor = vec4( vec3( mix( 0.2, 0.8, uv.y ) ), 1.0 );
	#endif
	vec3 outgoingLight = diffuseColor.rgb * matcapColor.rgb;
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
	#include <dithering_fragment>
}`,meshnormal_vert:`#define NORMAL
#if defined( FLAT_SHADED ) || defined( USE_BUMPMAP ) || defined( USE_NORMALMAP_TANGENTSPACE )
	varying vec3 vViewPosition;
#endif
#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <displacementmap_pars_vertex>
#include <normal_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <batching_vertex>
	#include <beginnormal_vertex>
	#include <morphinstance_vertex>
	#include <morphnormal_vertex>
	#include <skinbase_vertex>
	#include <skinnormal_vertex>
	#include <defaultnormal_vertex>
	#include <normal_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
#if defined( FLAT_SHADED ) || defined( USE_BUMPMAP ) || defined( USE_NORMALMAP_TANGENTSPACE )
	vViewPosition = - mvPosition.xyz;
#endif
}`,meshnormal_frag:`#define NORMAL
uniform float opacity;
#if defined( FLAT_SHADED ) || defined( USE_BUMPMAP ) || defined( USE_NORMALMAP_TANGENTSPACE )
	varying vec3 vViewPosition;
#endif
#include <uv_pars_fragment>
#include <normal_pars_fragment>
#include <bumpmap_pars_fragment>
#include <normalmap_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( 0.0, 0.0, 0.0, opacity );
	#include <clipping_planes_fragment>
	#include <logdepthbuf_fragment>
	#include <normal_fragment_begin>
	#include <normal_fragment_maps>
	gl_FragColor = vec4( normalize( normal ) * 0.5 + 0.5, diffuseColor.a );
	#ifdef OPAQUE
		gl_FragColor.a = 1.0;
	#endif
}`,meshphong_vert:`#define PHONG
varying vec3 vViewPosition;
#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <displacementmap_pars_vertex>
#include <envmap_pars_vertex>
#include <color_pars_vertex>
#include <fog_pars_vertex>
#include <normal_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <shadowmap_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <color_vertex>
	#include <morphcolor_vertex>
	#include <batching_vertex>
	#include <beginnormal_vertex>
	#include <morphinstance_vertex>
	#include <morphnormal_vertex>
	#include <skinbase_vertex>
	#include <skinnormal_vertex>
	#include <defaultnormal_vertex>
	#include <normal_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	vViewPosition = - mvPosition.xyz;
	#include <worldpos_vertex>
	#include <envmap_vertex>
	#include <shadowmap_vertex>
	#include <fog_vertex>
}`,meshphong_frag:`#define PHONG
uniform vec3 diffuse;
uniform vec3 emissive;
uniform vec3 specular;
uniform float shininess;
uniform float opacity;
#include <common>
#include <dithering_pars_fragment>
#include <color_pars_fragment>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <aomap_pars_fragment>
#include <lightmap_pars_fragment>
#include <emissivemap_pars_fragment>
#include <cube_uv_reflection_fragment>
#include <envmap_common_pars_fragment>
#include <envmap_pars_fragment>
#include <envmap_physical_pars_fragment>
#include <fog_pars_fragment>
#include <bsdfs>
#include <lights_pars_begin>
#include <normal_pars_fragment>
#include <lights_phong_pars_fragment>
#include <shadowmap_pars_fragment>
#include <bumpmap_pars_fragment>
#include <normalmap_pars_fragment>
#include <specularmap_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	ReflectedLight reflectedLight = ReflectedLight( vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ) );
	vec3 totalEmissiveRadiance = emissive;
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <color_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	#include <specularmap_fragment>
	#include <normal_fragment_begin>
	#include <normal_fragment_maps>
	#include <emissivemap_fragment>
	#include <lights_phong_fragment>
	#include <lights_fragment_begin>
	#include <lights_fragment_maps>
	#include <lights_fragment_end>
	#include <aomap_fragment>
	vec3 outgoingLight = reflectedLight.directDiffuse + reflectedLight.indirectDiffuse + reflectedLight.directSpecular + reflectedLight.indirectSpecular + totalEmissiveRadiance;
	#include <envmap_fragment>
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
	#include <dithering_fragment>
}`,meshphysical_vert:`#define STANDARD
varying vec3 vViewPosition;
#ifdef USE_TRANSMISSION
	varying vec3 vWorldPosition;
#endif
#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <displacementmap_pars_vertex>
#include <color_pars_vertex>
#include <fog_pars_vertex>
#include <normal_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <shadowmap_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <color_vertex>
	#include <morphinstance_vertex>
	#include <morphcolor_vertex>
	#include <batching_vertex>
	#include <beginnormal_vertex>
	#include <morphnormal_vertex>
	#include <skinbase_vertex>
	#include <skinnormal_vertex>
	#include <defaultnormal_vertex>
	#include <normal_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	vViewPosition = - mvPosition.xyz;
	#include <worldpos_vertex>
	#include <shadowmap_vertex>
	#include <fog_vertex>
#ifdef USE_TRANSMISSION
	vWorldPosition = worldPosition.xyz;
#endif
}`,meshphysical_frag:`#define STANDARD
#ifdef PHYSICAL
	#define IOR
	#define USE_SPECULAR
#endif
uniform vec3 diffuse;
uniform vec3 emissive;
uniform float roughness;
uniform float metalness;
uniform float opacity;
#ifdef IOR
	uniform float ior;
#endif
#ifdef USE_SPECULAR
	uniform float specularIntensity;
	uniform vec3 specularColor;
	#ifdef USE_SPECULAR_COLORMAP
		uniform sampler2D specularColorMap;
	#endif
	#ifdef USE_SPECULAR_INTENSITYMAP
		uniform sampler2D specularIntensityMap;
	#endif
#endif
#ifdef USE_CLEARCOAT
	uniform float clearcoat;
	uniform float clearcoatRoughness;
#endif
#ifdef USE_DISPERSION
	uniform float dispersion;
#endif
#ifdef USE_IRIDESCENCE
	uniform float iridescence;
	uniform float iridescenceIOR;
	uniform float iridescenceThicknessMinimum;
	uniform float iridescenceThicknessMaximum;
#endif
#ifdef USE_SHEEN
	uniform vec3 sheenColor;
	uniform float sheenRoughness;
	#ifdef USE_SHEEN_COLORMAP
		uniform sampler2D sheenColorMap;
	#endif
	#ifdef USE_SHEEN_ROUGHNESSMAP
		uniform sampler2D sheenRoughnessMap;
	#endif
#endif
#ifdef USE_ANISOTROPY
	uniform vec2 anisotropyVector;
	#ifdef USE_ANISOTROPYMAP
		uniform sampler2D anisotropyMap;
	#endif
#endif
varying vec3 vViewPosition;
#include <common>
#include <dithering_pars_fragment>
#include <color_pars_fragment>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <aomap_pars_fragment>
#include <lightmap_pars_fragment>
#include <emissivemap_pars_fragment>
#include <iridescence_fragment>
#include <cube_uv_reflection_fragment>
#include <envmap_common_pars_fragment>
#include <envmap_physical_pars_fragment>
#include <fog_pars_fragment>
#include <lights_pars_begin>
#include <normal_pars_fragment>
#include <lights_physical_pars_fragment>
#include <transmission_pars_fragment>
#include <shadowmap_pars_fragment>
#include <bumpmap_pars_fragment>
#include <normalmap_pars_fragment>
#include <clearcoat_pars_fragment>
#include <iridescence_pars_fragment>
#include <roughnessmap_pars_fragment>
#include <metalnessmap_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	ReflectedLight reflectedLight = ReflectedLight( vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ) );
	vec3 totalEmissiveRadiance = emissive;
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <color_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	#include <roughnessmap_fragment>
	#include <metalnessmap_fragment>
	#include <normal_fragment_begin>
	#include <normal_fragment_maps>
	#include <clearcoat_normal_fragment_begin>
	#include <clearcoat_normal_fragment_maps>
	#include <emissivemap_fragment>
	#include <lights_physical_fragment>
	#include <lights_fragment_begin>
	#include <lights_fragment_maps>
	#include <lights_fragment_end>
	#include <aomap_fragment>
	vec3 totalDiffuse = reflectedLight.directDiffuse + reflectedLight.indirectDiffuse;
	vec3 totalSpecular = reflectedLight.directSpecular + reflectedLight.indirectSpecular;
	#include <transmission_fragment>
	vec3 outgoingLight = totalDiffuse + totalSpecular + totalEmissiveRadiance;
	#ifdef USE_SHEEN
 
		outgoingLight = outgoingLight + sheenSpecularDirect + sheenSpecularIndirect;
 
 	#endif
	#ifdef USE_CLEARCOAT
		float dotNVcc = saturate( dot( geometryClearcoatNormal, geometryViewDir ) );
		vec3 Fcc = F_Schlick( material.clearcoatF0, material.clearcoatF90, dotNVcc );
		outgoingLight = outgoingLight * ( 1.0 - material.clearcoat * Fcc ) + ( clearcoatSpecularDirect + clearcoatSpecularIndirect ) * material.clearcoat;
	#endif
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
	#include <dithering_fragment>
}`,meshtoon_vert:`#define TOON
varying vec3 vViewPosition;
#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <displacementmap_pars_vertex>
#include <color_pars_vertex>
#include <fog_pars_vertex>
#include <normal_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <shadowmap_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <color_vertex>
	#include <morphinstance_vertex>
	#include <morphcolor_vertex>
	#include <batching_vertex>
	#include <beginnormal_vertex>
	#include <morphnormal_vertex>
	#include <skinbase_vertex>
	#include <skinnormal_vertex>
	#include <defaultnormal_vertex>
	#include <normal_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	vViewPosition = - mvPosition.xyz;
	#include <worldpos_vertex>
	#include <shadowmap_vertex>
	#include <fog_vertex>
}`,meshtoon_frag:`#define TOON
uniform vec3 diffuse;
uniform vec3 emissive;
uniform float opacity;
#include <common>
#include <dithering_pars_fragment>
#include <color_pars_fragment>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <aomap_pars_fragment>
#include <lightmap_pars_fragment>
#include <emissivemap_pars_fragment>
#include <gradientmap_pars_fragment>
#include <fog_pars_fragment>
#include <bsdfs>
#include <lights_pars_begin>
#include <normal_pars_fragment>
#include <lights_toon_pars_fragment>
#include <shadowmap_pars_fragment>
#include <bumpmap_pars_fragment>
#include <normalmap_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	ReflectedLight reflectedLight = ReflectedLight( vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ) );
	vec3 totalEmissiveRadiance = emissive;
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <color_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	#include <normal_fragment_begin>
	#include <normal_fragment_maps>
	#include <emissivemap_fragment>
	#include <lights_toon_fragment>
	#include <lights_fragment_begin>
	#include <lights_fragment_maps>
	#include <lights_fragment_end>
	#include <aomap_fragment>
	vec3 outgoingLight = reflectedLight.directDiffuse + reflectedLight.indirectDiffuse + totalEmissiveRadiance;
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
	#include <dithering_fragment>
}`,points_vert:`uniform float size;
uniform float scale;
#include <common>
#include <color_pars_vertex>
#include <fog_pars_vertex>
#include <morphtarget_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
#ifdef USE_POINTS_UV
	varying vec2 vUv;
	uniform mat3 uvTransform;
#endif
void main() {
	#ifdef USE_POINTS_UV
		vUv = ( uvTransform * vec3( uv, 1 ) ).xy;
	#endif
	#include <color_vertex>
	#include <morphinstance_vertex>
	#include <morphcolor_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <project_vertex>
	gl_PointSize = size;
	#ifdef USE_SIZEATTENUATION
		bool isPerspective = isPerspectiveMatrix( projectionMatrix );
		if ( isPerspective ) gl_PointSize *= ( scale / - mvPosition.z );
	#endif
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	#include <worldpos_vertex>
	#include <fog_vertex>
}`,points_frag:`uniform vec3 diffuse;
uniform float opacity;
#include <common>
#include <color_pars_fragment>
#include <map_particle_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <fog_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	vec3 outgoingLight = vec3( 0.0 );
	#include <logdepthbuf_fragment>
	#include <map_particle_fragment>
	#include <color_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	outgoingLight = diffuseColor.rgb;
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
}`,shadow_vert:`#include <common>
#include <batching_pars_vertex>
#include <fog_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <shadowmap_pars_vertex>
void main() {
	#include <batching_vertex>
	#include <beginnormal_vertex>
	#include <morphinstance_vertex>
	#include <morphnormal_vertex>
	#include <skinbase_vertex>
	#include <skinnormal_vertex>
	#include <defaultnormal_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <worldpos_vertex>
	#include <shadowmap_vertex>
	#include <fog_vertex>
}`,shadow_frag:`uniform vec3 color;
uniform float opacity;
#include <common>
#include <fog_pars_fragment>
#include <bsdfs>
#include <lights_pars_begin>
#include <logdepthbuf_pars_fragment>
#include <shadowmap_pars_fragment>
#include <shadowmask_pars_fragment>
void main() {
	#include <logdepthbuf_fragment>
	gl_FragColor = vec4( color, opacity * ( 1.0 - getShadowMask() ) );
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
}`,sprite_vert:`uniform float rotation;
uniform vec2 center;
#include <common>
#include <uv_pars_vertex>
#include <fog_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	vec4 mvPosition = modelViewMatrix[ 3 ];
	vec2 scale = vec2( length( modelMatrix[ 0 ].xyz ), length( modelMatrix[ 1 ].xyz ) );
	#ifndef USE_SIZEATTENUATION
		bool isPerspective = isPerspectiveMatrix( projectionMatrix );
		if ( isPerspective ) scale *= - mvPosition.z;
	#endif
	vec2 alignedPosition = ( position.xy - ( center - vec2( 0.5 ) ) ) * scale;
	vec2 rotatedPosition;
	rotatedPosition.x = cos( rotation ) * alignedPosition.x - sin( rotation ) * alignedPosition.y;
	rotatedPosition.y = sin( rotation ) * alignedPosition.x + cos( rotation ) * alignedPosition.y;
	mvPosition.xy += rotatedPosition;
	gl_Position = projectionMatrix * mvPosition;
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	#include <fog_vertex>
}`,sprite_frag:`uniform vec3 diffuse;
uniform float opacity;
#include <common>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <fog_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	vec3 outgoingLight = vec3( 0.0 );
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	outgoingLight = diffuseColor.rgb;
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
}`},Nt={common:{diffuse:{value:new f.Q1f(16777215)},opacity:{value:1},map:{value:null},mapTransform:{value:new f.dwI},alphaMap:{value:null},alphaMapTransform:{value:new f.dwI},alphaTest:{value:0}},specularmap:{specularMap:{value:null},specularMapTransform:{value:new f.dwI}},envmap:{envMap:{value:null},envMapRotation:{value:new f.dwI},flipEnvMap:{value:-1},reflectivity:{value:1},ior:{value:1.5},refractionRatio:{value:.98},dfgLUT:{value:null}},aomap:{aoMap:{value:null},aoMapIntensity:{value:1},aoMapTransform:{value:new f.dwI}},lightmap:{lightMap:{value:null},lightMapIntensity:{value:1},lightMapTransform:{value:new f.dwI}},bumpmap:{bumpMap:{value:null},bumpMapTransform:{value:new f.dwI},bumpScale:{value:1}},normalmap:{normalMap:{value:null},normalMapTransform:{value:new f.dwI},normalScale:{value:new f.I9Y(1,1)}},displacementmap:{displacementMap:{value:null},displacementMapTransform:{value:new f.dwI},displacementScale:{value:1},displacementBias:{value:0}},emissivemap:{emissiveMap:{value:null},emissiveMapTransform:{value:new f.dwI}},metalnessmap:{metalnessMap:{value:null},metalnessMapTransform:{value:new f.dwI}},roughnessmap:{roughnessMap:{value:null},roughnessMapTransform:{value:new f.dwI}},gradientmap:{gradientMap:{value:null}},fog:{fogDensity:{value:25e-5},fogNear:{value:1},fogFar:{value:2e3},fogColor:{value:new f.Q1f(16777215)}},lights:{ambientLightColor:{value:[]},lightProbe:{value:[]},directionalLights:{value:[],properties:{direction:{},color:{}}},directionalLightShadows:{value:[],properties:{shadowIntensity:1,shadowBias:{},shadowNormalBias:{},shadowRadius:{},shadowMapSize:{}}},directionalShadowMatrix:{value:[]},spotLights:{value:[],properties:{color:{},position:{},direction:{},distance:{},coneCos:{},penumbraCos:{},decay:{}}},spotLightShadows:{value:[],properties:{shadowIntensity:1,shadowBias:{},shadowNormalBias:{},shadowRadius:{},shadowMapSize:{}}},spotLightMap:{value:[]},spotLightMatrix:{value:[]},pointLights:{value:[],properties:{color:{},position:{},decay:{},distance:{}}},pointLightShadows:{value:[],properties:{shadowIntensity:1,shadowBias:{},shadowNormalBias:{},shadowRadius:{},shadowMapSize:{},shadowCameraNear:{},shadowCameraFar:{}}},pointShadowMatrix:{value:[]},hemisphereLights:{value:[],properties:{direction:{},skyColor:{},groundColor:{}}},rectAreaLights:{value:[],properties:{color:{},position:{},width:{},height:{}}},ltc_1:{value:null},ltc_2:{value:null}},points:{diffuse:{value:new f.Q1f(16777215)},opacity:{value:1},size:{value:1},scale:{value:1},map:{value:null},alphaMap:{value:null},alphaMapTransform:{value:new f.dwI},alphaTest:{value:0},uvTransform:{value:new f.dwI}},sprite:{diffuse:{value:new f.Q1f(16777215)},opacity:{value:1},center:{value:new f.I9Y(.5,.5)},rotation:{value:0},map:{value:null},mapTransform:{value:new f.dwI},alphaMap:{value:null},alphaMapTransform:{value:new f.dwI},alphaTest:{value:0}}},Dn={basic:{uniforms:(0,f.Iit)([Nt.common,Nt.specularmap,Nt.envmap,Nt.aomap,Nt.lightmap,Nt.fog]),vertexShader:ve.meshbasic_vert,fragmentShader:ve.meshbasic_frag},lambert:{uniforms:(0,f.Iit)([Nt.common,Nt.specularmap,Nt.envmap,Nt.aomap,Nt.lightmap,Nt.emissivemap,Nt.bumpmap,Nt.normalmap,Nt.displacementmap,Nt.fog,Nt.lights,{emissive:{value:new f.Q1f(0)},envMapIntensity:{value:1}}]),vertexShader:ve.meshlambert_vert,fragmentShader:ve.meshlambert_frag},phong:{uniforms:(0,f.Iit)([Nt.common,Nt.specularmap,Nt.envmap,Nt.aomap,Nt.lightmap,Nt.emissivemap,Nt.bumpmap,Nt.normalmap,Nt.displacementmap,Nt.fog,Nt.lights,{emissive:{value:new f.Q1f(0)},specular:{value:new f.Q1f(1118481)},shininess:{value:30},envMapIntensity:{value:1}}]),vertexShader:ve.meshphong_vert,fragmentShader:ve.meshphong_frag},standard:{uniforms:(0,f.Iit)([Nt.common,Nt.envmap,Nt.aomap,Nt.lightmap,Nt.emissivemap,Nt.bumpmap,Nt.normalmap,Nt.displacementmap,Nt.roughnessmap,Nt.metalnessmap,Nt.fog,Nt.lights,{emissive:{value:new f.Q1f(0)},roughness:{value:1},metalness:{value:0},envMapIntensity:{value:1}}]),vertexShader:ve.meshphysical_vert,fragmentShader:ve.meshphysical_frag},toon:{uniforms:(0,f.Iit)([Nt.common,Nt.aomap,Nt.lightmap,Nt.emissivemap,Nt.bumpmap,Nt.normalmap,Nt.displacementmap,Nt.gradientmap,Nt.fog,Nt.lights,{emissive:{value:new f.Q1f(0)}}]),vertexShader:ve.meshtoon_vert,fragmentShader:ve.meshtoon_frag},matcap:{uniforms:(0,f.Iit)([Nt.common,Nt.bumpmap,Nt.normalmap,Nt.displacementmap,Nt.fog,{matcap:{value:null}}]),vertexShader:ve.meshmatcap_vert,fragmentShader:ve.meshmatcap_frag},points:{uniforms:(0,f.Iit)([Nt.points,Nt.fog]),vertexShader:ve.points_vert,fragmentShader:ve.points_frag},dashed:{uniforms:(0,f.Iit)([Nt.common,Nt.fog,{scale:{value:1},dashSize:{value:1},totalSize:{value:2}}]),vertexShader:ve.linedashed_vert,fragmentShader:ve.linedashed_frag},depth:{uniforms:(0,f.Iit)([Nt.common,Nt.displacementmap]),vertexShader:ve.depth_vert,fragmentShader:ve.depth_frag},normal:{uniforms:(0,f.Iit)([Nt.common,Nt.bumpmap,Nt.normalmap,Nt.displacementmap,{opacity:{value:1}}]),vertexShader:ve.meshnormal_vert,fragmentShader:ve.meshnormal_frag},sprite:{uniforms:(0,f.Iit)([Nt.sprite,Nt.fog]),vertexShader:ve.sprite_vert,fragmentShader:ve.sprite_frag},background:{uniforms:{uvTransform:{value:new f.dwI},t2D:{value:null},backgroundIntensity:{value:1}},vertexShader:ve.background_vert,fragmentShader:ve.background_frag},backgroundCube:{uniforms:{envMap:{value:null},flipEnvMap:{value:-1},backgroundBlurriness:{value:0},backgroundIntensity:{value:1},backgroundRotation:{value:new f.dwI}},vertexShader:ve.backgroundCube_vert,fragmentShader:ve.backgroundCube_frag},cube:{uniforms:{tCube:{value:null},tFlip:{value:-1},opacity:{value:1}},vertexShader:ve.cube_vert,fragmentShader:ve.cube_frag},equirect:{uniforms:{tEquirect:{value:null}},vertexShader:ve.equirect_vert,fragmentShader:ve.equirect_frag},distance:{uniforms:(0,f.Iit)([Nt.common,Nt.displacementmap,{referencePosition:{value:new f.Pq0},nearDistance:{value:1},farDistance:{value:1e3}}]),vertexShader:ve.distance_vert,fragmentShader:ve.distance_frag},shadow:{uniforms:(0,f.Iit)([Nt.lights,Nt.fog,{color:{value:new f.Q1f(0)},opacity:{value:1}}]),vertexShader:ve.shadow_vert,fragmentShader:ve.shadow_frag}};Dn.physical={uniforms:(0,f.Iit)([Dn.standard.uniforms,{clearcoat:{value:0},clearcoatMap:{value:null},clearcoatMapTransform:{value:new f.dwI},clearcoatNormalMap:{value:null},clearcoatNormalMapTransform:{value:new f.dwI},clearcoatNormalScale:{value:new f.I9Y(1,1)},clearcoatRoughness:{value:0},clearcoatRoughnessMap:{value:null},clearcoatRoughnessMapTransform:{value:new f.dwI},dispersion:{value:0},iridescence:{value:0},iridescenceMap:{value:null},iridescenceMapTransform:{value:new f.dwI},iridescenceIOR:{value:1.3},iridescenceThicknessMinimum:{value:100},iridescenceThicknessMaximum:{value:400},iridescenceThicknessMap:{value:null},iridescenceThicknessMapTransform:{value:new f.dwI},sheen:{value:0},sheenColor:{value:new f.Q1f(0)},sheenColorMap:{value:null},sheenColorMapTransform:{value:new f.dwI},sheenRoughness:{value:1},sheenRoughnessMap:{value:null},sheenRoughnessMapTransform:{value:new f.dwI},transmission:{value:0},transmissionMap:{value:null},transmissionMapTransform:{value:new f.dwI},transmissionSamplerSize:{value:new f.I9Y},transmissionSamplerMap:{value:null},thickness:{value:0},thicknessMap:{value:null},thicknessMapTransform:{value:new f.dwI},attenuationDistance:{value:0},attenuationColor:{value:new f.Q1f(0)},specularColor:{value:new f.Q1f(1,1,1)},specularColorMap:{value:null},specularColorMapTransform:{value:new f.dwI},specularIntensity:{value:1},specularIntensityMap:{value:null},specularIntensityMapTransform:{value:new f.dwI},anisotropyVector:{value:new f.I9Y},anisotropyMap:{value:null},anisotropyMapTransform:{value:new f.dwI}}]),vertexShader:ve.meshphysical_vert,fragmentShader:ve.meshphysical_frag};const Jn={r:0,b:0,g:0},Rn=new f.O9p,Ls=new f.kn4;function fr(o,p,h,x,E,S){const I=new f.Q1f(0);let V=E===!0?0:1,j,$,ht=null,K=0,F=null;function G(z){let q=z.isScene===!0?z.background:null;if(q&&q.isTexture){const H=z.backgroundBlurriness>0;q=p.get(q,H)}return q}function X(z){let q=!1;const H=G(z);H===null?A(I,V):H&&H.isColor&&(A(H,1),q=!0);const ot=o.xr.getEnvironmentBlendMode();ot==="additive"?h.buffers.color.setClear(0,0,0,1,S):ot==="alpha-blend"&&h.buffers.color.setClear(0,0,0,0,S),(o.autoClear||q)&&(h.buffers.depth.setTest(!0),h.buffers.depth.setMask(!0),h.buffers.color.setMask(!0),o.clear(o.autoClearColor,o.autoClearDepth,o.autoClearStencil))}function nt(z,q){const H=G(q);H&&(H.isCubeTexture||H.mapping===f.Om)?($===void 0&&($=new f.eaF(new f.iNn(1,1,1),new f.BKk({name:"BackgroundCubeMaterial",uniforms:(0,f.lxW)(Dn.backgroundCube.uniforms),vertexShader:Dn.backgroundCube.vertexShader,fragmentShader:Dn.backgroundCube.fragmentShader,side:f.hsX,depthTest:!1,depthWrite:!1,fog:!1,allowOverride:!1})),$.geometry.deleteAttribute("normal"),$.geometry.deleteAttribute("uv"),$.onBeforeRender=function(ot,J,st){this.matrixWorld.copyPosition(st.matrixWorld)},Object.defineProperty($.material,"envMap",{get:function(){return this.uniforms.envMap.value}}),x.update($)),Rn.copy(q.backgroundRotation),Rn.x*=-1,Rn.y*=-1,Rn.z*=-1,H.isCubeTexture&&H.isRenderTargetTexture===!1&&(Rn.y*=-1,Rn.z*=-1),$.material.uniforms.envMap.value=H,$.material.uniforms.flipEnvMap.value=H.isCubeTexture&&H.isRenderTargetTexture===!1?-1:1,$.material.uniforms.backgroundBlurriness.value=q.backgroundBlurriness,$.material.uniforms.backgroundIntensity.value=q.backgroundIntensity,$.material.uniforms.backgroundRotation.value.setFromMatrix4(Ls.makeRotationFromEuler(Rn)),$.material.toneMapped=f.ppV.getTransfer(H.colorSpace)!==f.KLL,(ht!==H||K!==H.version||F!==o.toneMapping)&&($.material.needsUpdate=!0,ht=H,K=H.version,F=o.toneMapping),$.layers.enableAll(),z.unshift($,$.geometry,$.material,0,0,null)):H&&H.isTexture&&(j===void 0&&(j=new f.eaF(new f.bdM(2,2),new f.BKk({name:"BackgroundMaterial",uniforms:(0,f.lxW)(Dn.background.uniforms),vertexShader:Dn.background.vertexShader,fragmentShader:Dn.background.fragmentShader,side:f.hB5,depthTest:!1,depthWrite:!1,fog:!1,allowOverride:!1})),j.geometry.deleteAttribute("normal"),Object.defineProperty(j.material,"map",{get:function(){return this.uniforms.t2D.value}}),x.update(j)),j.material.uniforms.t2D.value=H,j.material.uniforms.backgroundIntensity.value=q.backgroundIntensity,j.material.toneMapped=f.ppV.getTransfer(H.colorSpace)!==f.KLL,H.matrixAutoUpdate===!0&&H.updateMatrix(),j.material.uniforms.uvTransform.value.copy(H.matrix),(ht!==H||K!==H.version||F!==o.toneMapping)&&(j.material.needsUpdate=!0,ht=H,K=H.version,F=o.toneMapping),j.layers.enableAll(),z.unshift(j,j.geometry,j.material,0,0,null))}function A(z,q){z.getRGB(Jn,(0,f._Ut)(o)),h.buffers.color.setClear(Jn.r,Jn.g,Jn.b,q,S)}function b(){$!==void 0&&($.geometry.dispose(),$.material.dispose(),$=void 0),j!==void 0&&(j.geometry.dispose(),j.material.dispose(),j=void 0)}return{getClearColor:function(){return I},setClearColor:function(z,q=1){I.set(z),V=q,A(I,V)},getClearAlpha:function(){return V},setClearAlpha:function(z){V=z,A(I,V)},render:X,addToRenderList:nt,dispose:b}}function ba(o,p){const h=o.getParameter(o.MAX_VERTEX_ATTRIBS),x={},E=F(null);let S=E,I=!1;function V(Z,dt,mt,gt,_t){let ct=!1;const it=K(Z,gt,mt,dt);S!==it&&(S=it,$(S.object)),ct=G(Z,gt,mt,_t),ct&&X(Z,gt,mt,_t),_t!==null&&p.update(_t,o.ELEMENT_ARRAY_BUFFER),(ct||I)&&(I=!1,H(Z,dt,mt,gt),_t!==null&&o.bindBuffer(o.ELEMENT_ARRAY_BUFFER,p.get(_t).buffer))}function j(){return o.createVertexArray()}function $(Z){return o.bindVertexArray(Z)}function ht(Z){return o.deleteVertexArray(Z)}function K(Z,dt,mt,gt){const _t=gt.wireframe===!0;let ct=x[dt.id];ct===void 0&&(ct={},x[dt.id]=ct);const it=Z.isInstancedMesh===!0?Z.id:0;let It=ct[it];It===void 0&&(It={},ct[it]=It);let zt=It[mt.id];zt===void 0&&(zt={},It[mt.id]=zt);let kt=zt[_t];return kt===void 0&&(kt=F(j()),zt[_t]=kt),kt}function F(Z){const dt=[],mt=[],gt=[];for(let _t=0;_t<h;_t++)dt[_t]=0,mt[_t]=0,gt[_t]=0;return{geometry:null,program:null,wireframe:!1,newAttributes:dt,enabledAttributes:mt,attributeDivisors:gt,object:Z,attributes:{},index:null}}function G(Z,dt,mt,gt){const _t=S.attributes,ct=dt.attributes;let it=0;const It=mt.getAttributes();for(const zt in It)if(It[zt].location>=0){const ue=_t[zt];let te=ct[zt];if(te===void 0&&(zt==="instanceMatrix"&&Z.instanceMatrix&&(te=Z.instanceMatrix),zt==="instanceColor"&&Z.instanceColor&&(te=Z.instanceColor)),ue===void 0||ue.attribute!==te||te&&ue.data!==te.data)return!0;it++}return S.attributesNum!==it||S.index!==gt}function X(Z,dt,mt,gt){const _t={},ct=dt.attributes;let it=0;const It=mt.getAttributes();for(const zt in It)if(It[zt].location>=0){let ue=ct[zt];ue===void 0&&(zt==="instanceMatrix"&&Z.instanceMatrix&&(ue=Z.instanceMatrix),zt==="instanceColor"&&Z.instanceColor&&(ue=Z.instanceColor));const te={};te.attribute=ue,ue&&ue.data&&(te.data=ue.data),_t[zt]=te,it++}S.attributes=_t,S.attributesNum=it,S.index=gt}function nt(){const Z=S.newAttributes;for(let dt=0,mt=Z.length;dt<mt;dt++)Z[dt]=0}function A(Z){b(Z,0)}function b(Z,dt){const mt=S.newAttributes,gt=S.enabledAttributes,_t=S.attributeDivisors;mt[Z]=1,gt[Z]===0&&(o.enableVertexAttribArray(Z),gt[Z]=1),_t[Z]!==dt&&(o.vertexAttribDivisor(Z,dt),_t[Z]=dt)}function z(){const Z=S.newAttributes,dt=S.enabledAttributes;for(let mt=0,gt=dt.length;mt<gt;mt++)dt[mt]!==Z[mt]&&(o.disableVertexAttribArray(mt),dt[mt]=0)}function q(Z,dt,mt,gt,_t,ct,it){it===!0?o.vertexAttribIPointer(Z,dt,mt,_t,ct):o.vertexAttribPointer(Z,dt,mt,gt,_t,ct)}function H(Z,dt,mt,gt){nt();const _t=gt.attributes,ct=mt.getAttributes(),it=dt.defaultAttributeValues;for(const It in ct){const zt=ct[It];if(zt.location>=0){let kt=_t[It];if(kt===void 0&&(It==="instanceMatrix"&&Z.instanceMatrix&&(kt=Z.instanceMatrix),It==="instanceColor"&&Z.instanceColor&&(kt=Z.instanceColor)),kt!==void 0){const ue=kt.normalized,te=kt.itemSize,se=p.get(kt);if(se===void 0)continue;const Ze=se.buffer,$e=se.type,xt=se.bytesPerElement,Lt=$e===o.INT||$e===o.UNSIGNED_INT||kt.gpuType===f.Yuy;if(kt.isInterleavedBufferAttribute){const Dt=kt.data,Te=Dt.stride,de=kt.offset;if(Dt.isInstancedInterleavedBuffer){for(let _e=0;_e<zt.locationSize;_e++)b(zt.location+_e,Dt.meshPerAttribute);Z.isInstancedMesh!==!0&&gt._maxInstanceCount===void 0&&(gt._maxInstanceCount=Dt.meshPerAttribute*Dt.count)}else for(let _e=0;_e<zt.locationSize;_e++)A(zt.location+_e);o.bindBuffer(o.ARRAY_BUFFER,Ze);for(let _e=0;_e<zt.locationSize;_e++)q(zt.location+_e,te/zt.locationSize,$e,ue,Te*xt,(de+te/zt.locationSize*_e)*xt,Lt)}else{if(kt.isInstancedBufferAttribute){for(let Dt=0;Dt<zt.locationSize;Dt++)b(zt.location+Dt,kt.meshPerAttribute);Z.isInstancedMesh!==!0&&gt._maxInstanceCount===void 0&&(gt._maxInstanceCount=kt.meshPerAttribute*kt.count)}else for(let Dt=0;Dt<zt.locationSize;Dt++)A(zt.location+Dt);o.bindBuffer(o.ARRAY_BUFFER,Ze);for(let Dt=0;Dt<zt.locationSize;Dt++)q(zt.location+Dt,te/zt.locationSize,$e,ue,te*xt,te/zt.locationSize*Dt*xt,Lt)}}else if(it!==void 0){const ue=it[It];if(ue!==void 0)switch(ue.length){case 2:o.vertexAttrib2fv(zt.location,ue);break;case 3:o.vertexAttrib3fv(zt.location,ue);break;case 4:o.vertexAttrib4fv(zt.location,ue);break;default:o.vertexAttrib1fv(zt.location,ue)}}}}z()}function ot(){U();for(const Z in x){const dt=x[Z];for(const mt in dt){const gt=dt[mt];for(const _t in gt){const ct=gt[_t];for(const it in ct)ht(ct[it].object),delete ct[it];delete gt[_t]}}delete x[Z]}}function J(Z){if(x[Z.id]===void 0)return;const dt=x[Z.id];for(const mt in dt){const gt=dt[mt];for(const _t in gt){const ct=gt[_t];for(const it in ct)ht(ct[it].object),delete ct[it];delete gt[_t]}}delete x[Z.id]}function st(Z){for(const dt in x){const mt=x[dt];for(const gt in mt){const _t=mt[gt];if(_t[Z.id]===void 0)continue;const ct=_t[Z.id];for(const it in ct)ht(ct[it].object),delete ct[it];delete _t[Z.id]}}}function R(Z){for(const dt in x){const mt=x[dt],gt=Z.isInstancedMesh===!0?Z.id:0,_t=mt[gt];if(_t!==void 0){for(const ct in _t){const it=_t[ct];for(const It in it)ht(it[It].object),delete it[It];delete _t[ct]}delete mt[gt],Object.keys(mt).length===0&&delete x[dt]}}}function U(){Tt(),I=!0,S!==E&&(S=E,$(S.object))}function Tt(){E.geometry=null,E.program=null,E.wireframe=!1}return{setup:V,reset:U,resetDefaultState:Tt,dispose:ot,releaseStatesOfGeometry:J,releaseStatesOfObject:R,releaseStatesOfProgram:st,initAttributes:nt,enableAttribute:A,disableUnusedAttributes:z}}function Fh(o,p,h){let x;function E($){x=$}function S($,ht){o.drawArrays(x,$,ht),h.update(ht,x,1)}function I($,ht,K){K!==0&&(o.drawArraysInstanced(x,$,ht,K),h.update(ht,x,K))}function V($,ht,K){if(K===0)return;p.get("WEBGL_multi_draw").multiDrawArraysWEBGL(x,$,0,ht,0,K);let G=0;for(let X=0;X<K;X++)G+=ht[X];h.update(G,x,1)}function j($,ht,K,F){if(K===0)return;const G=p.get("WEBGL_multi_draw");if(G===null)for(let X=0;X<$.length;X++)I($[X],ht[X],F[X]);else{G.multiDrawArraysInstancedWEBGL(x,$,0,ht,0,F,0,K);let X=0;for(let nt=0;nt<K;nt++)X+=ht[nt]*F[nt];h.update(X,x,1)}}this.setMode=E,this.render=S,this.renderInstances=I,this.renderMultiDraw=V,this.renderMultiDrawInstances=j}function Oh(o,p,h,x){let E;function S(){if(E!==void 0)return E;if(p.has("EXT_texture_filter_anisotropic")===!0){const st=p.get("EXT_texture_filter_anisotropic");E=o.getParameter(st.MAX_TEXTURE_MAX_ANISOTROPY_EXT)}else E=0;return E}function I(st){return!(st!==f.GWd&&x.convert(st)!==o.getParameter(o.IMPLEMENTATION_COLOR_READ_FORMAT))}function V(st){const R=st===f.ix0&&(p.has("EXT_color_buffer_half_float")||p.has("EXT_color_buffer_float"));return!(st!==f.OUM&&x.convert(st)!==o.getParameter(o.IMPLEMENTATION_COLOR_READ_TYPE)&&st!==f.RQf&&!R)}function j(st){if(st==="highp"){if(o.getShaderPrecisionFormat(o.VERTEX_SHADER,o.HIGH_FLOAT).precision>0&&o.getShaderPrecisionFormat(o.FRAGMENT_SHADER,o.HIGH_FLOAT).precision>0)return"highp";st="mediump"}return st==="mediump"&&o.getShaderPrecisionFormat(o.VERTEX_SHADER,o.MEDIUM_FLOAT).precision>0&&o.getShaderPrecisionFormat(o.FRAGMENT_SHADER,o.MEDIUM_FLOAT).precision>0?"mediump":"lowp"}let $=h.precision!==void 0?h.precision:"highp";const ht=j($);ht!==$&&((0,f.R8M)("WebGLRenderer:",$,"not supported, using",ht,"instead."),$=ht);const K=h.logarithmicDepthBuffer===!0,F=h.reversedDepthBuffer===!0&&p.has("EXT_clip_control"),G=o.getParameter(o.MAX_TEXTURE_IMAGE_UNITS),X=o.getParameter(o.MAX_VERTEX_TEXTURE_IMAGE_UNITS),nt=o.getParameter(o.MAX_TEXTURE_SIZE),A=o.getParameter(o.MAX_CUBE_MAP_TEXTURE_SIZE),b=o.getParameter(o.MAX_VERTEX_ATTRIBS),z=o.getParameter(o.MAX_VERTEX_UNIFORM_VECTORS),q=o.getParameter(o.MAX_VARYING_VECTORS),H=o.getParameter(o.MAX_FRAGMENT_UNIFORM_VECTORS),ot=o.getParameter(o.MAX_SAMPLES),J=o.getParameter(o.SAMPLES);return{isWebGL2:!0,getMaxAnisotropy:S,getMaxPrecision:j,textureFormatReadable:I,textureTypeReadable:V,precision:$,logarithmicDepthBuffer:K,reversedDepthBuffer:F,maxTextures:G,maxVertexTextures:X,maxTextureSize:nt,maxCubemapSize:A,maxAttributes:b,maxVertexUniforms:z,maxVaryings:q,maxFragmentUniforms:H,maxSamples:ot,samples:J}}function Bh(o){const p=this;let h=null,x=0,E=!1,S=!1;const I=new f.Zcv,V=new f.dwI,j={value:null,needsUpdate:!1};this.uniform=j,this.numPlanes=0,this.numIntersection=0,this.init=function(K,F){const G=K.length!==0||F||x!==0||E;return E=F,x=K.length,G},this.beginShadows=function(){S=!0,ht(null)},this.endShadows=function(){S=!1},this.setGlobalState=function(K,F){h=ht(K,F,0)},this.setState=function(K,F,G){const X=K.clippingPlanes,nt=K.clipIntersection,A=K.clipShadows,b=o.get(K);if(!E||X===null||X.length===0||S&&!A)S?ht(null):$();else{const z=S?0:x,q=z*4;let H=b.clippingState||null;j.value=H,H=ht(X,F,q,G);for(let ot=0;ot!==q;++ot)H[ot]=h[ot];b.clippingState=H,this.numIntersection=nt?this.numPlanes:0,this.numPlanes+=z}};function $(){j.value!==h&&(j.value=h,j.needsUpdate=x>0),p.numPlanes=x,p.numIntersection=0}function ht(K,F,G,X){const nt=K!==null?K.length:0;let A=null;if(nt!==0){if(A=j.value,X!==!0||A===null){const b=G+nt*4,z=F.matrixWorldInverse;V.getNormalMatrix(z),(A===null||A.length<b)&&(A=new Float32Array(b));for(let q=0,H=G;q!==nt;++q,H+=4)I.copy(K[q]).applyMatrix4(z,V),I.normal.toArray(A,H),A[H+3]=I.constant}j.value=A,j.needsUpdate=!0}return p.numPlanes=nt,p.numIntersection=0,A}}const ri=4,ec=[.125,.215,.35,.446,.526,.582],Ri=20,zh=256,Un=new f.qUd,Ta=new f.Q1f;let $i=null,yn=0,Ji=0,Ki=!1;const Ds=new f.Pq0;class nc{constructor(p){this._renderer=p,this._pingPongRenderTarget=null,this._lodMax=0,this._cubeSize=0,this._sizeLods=[],this._sigmas=[],this._lodMeshes=[],this._backgroundBox=null,this._cubemapMaterial=null,this._equirectMaterial=null,this._blurMaterial=null,this._ggxMaterial=null}fromScene(p,h=0,x=.1,E=100,S={}){const{size:I=256,position:V=Ds}=S;$i=this._renderer.getRenderTarget(),yn=this._renderer.getActiveCubeFace(),Ji=this._renderer.getActiveMipmapLevel(),Ki=this._renderer.xr.enabled,this._renderer.xr.enabled=!1,this._setSize(I);const j=this._allocateTargets();return j.depthBuffer=!0,this._sceneToCubeUV(p,x,E,j,V),h>0&&this._blur(j,0,0,h),this._applyPMREM(j),this._cleanup(j),j}fromEquirectangular(p,h=null){return this._fromTexture(p,h)}fromCubemap(p,h=null){return this._fromTexture(p,h)}compileCubemapShader(){this._cubemapMaterial===null&&(this._cubemapMaterial=rc(),this._compileMaterial(this._cubemapMaterial))}compileEquirectangularShader(){this._equirectMaterial===null&&(this._equirectMaterial=sc(),this._compileMaterial(this._equirectMaterial))}dispose(){this._dispose(),this._cubemapMaterial!==null&&this._cubemapMaterial.dispose(),this._equirectMaterial!==null&&this._equirectMaterial.dispose(),this._backgroundBox!==null&&(this._backgroundBox.geometry.dispose(),this._backgroundBox.material.dispose())}_setSize(p){this._lodMax=Math.floor(Math.log2(p)),this._cubeSize=Math.pow(2,this._lodMax)}_dispose(){this._blurMaterial!==null&&this._blurMaterial.dispose(),this._ggxMaterial!==null&&this._ggxMaterial.dispose(),this._pingPongRenderTarget!==null&&this._pingPongRenderTarget.dispose();for(let p=0;p<this._lodMeshes.length;p++)this._lodMeshes[p].geometry.dispose()}_cleanup(p){this._renderer.setRenderTarget($i,yn,Ji),this._renderer.xr.enabled=Ki,p.scissorTest=!1,Qi(p,0,0,p.width,p.height)}_fromTexture(p,h){p.mapping===f.hy7||p.mapping===f.xFO?this._setSize(p.image.length===0?16:p.image[0].width||p.image[0].image.width):this._setSize(p.image.width/4),$i=this._renderer.getRenderTarget(),yn=this._renderer.getActiveCubeFace(),Ji=this._renderer.getActiveMipmapLevel(),Ki=this._renderer.xr.enabled,this._renderer.xr.enabled=!1;const x=h||this._allocateTargets();return this._textureToCubeUV(p,x),this._applyPMREM(x),this._cleanup(x),x}_allocateTargets(){const p=3*Math.max(this._cubeSize,112),h=4*this._cubeSize,x={magFilter:f.k6q,minFilter:f.k6q,generateMipmaps:!1,type:f.ix0,format:f.GWd,colorSpace:f.Zr2,depthBuffer:!1},E=ic(p,h,x);if(this._pingPongRenderTarget===null||this._pingPongRenderTarget.width!==p||this._pingPongRenderTarget.height!==h){this._pingPongRenderTarget!==null&&this._dispose(),this._pingPongRenderTarget=ic(p,h,x);const{_lodMax:S}=this;({lodMeshes:this._lodMeshes,sizeLods:this._sizeLods,sigmas:this._sigmas}=Vh(S)),this._blurMaterial=kh(S,p,h),this._ggxMaterial=Ii(S,p,h)}return E}_compileMaterial(p){const h=new f.eaF(new f.LoY,p);this._renderer.compile(h,Un)}_sceneToCubeUV(p,h,x,E,S){const j=new f.ubm(90,1,h,x),$=[1,-1,1,1,1,1],ht=[1,1,1,-1,-1,-1],K=this._renderer,F=K.autoClear,G=K.toneMapping;K.getClearColor(Ta),K.toneMapping=f.y_p,K.autoClear=!1,K.state.buffers.depth.getReversed()&&(K.setRenderTarget(E),K.clearDepth(),K.setRenderTarget(null)),this._backgroundBox===null&&(this._backgroundBox=new f.eaF(new f.iNn,new f.V9B({name:"PMREM.Background",side:f.hsX,depthWrite:!1,depthTest:!1})));const nt=this._backgroundBox,A=nt.material;let b=!1;const z=p.background;z?z.isColor&&(A.color.copy(z),p.background=null,b=!0):(A.color.copy(Ta),b=!0);for(let q=0;q<6;q++){const H=q%3;H===0?(j.up.set(0,$[q],0),j.position.set(S.x,S.y,S.z),j.lookAt(S.x+ht[q],S.y,S.z)):H===1?(j.up.set(0,0,$[q]),j.position.set(S.x,S.y,S.z),j.lookAt(S.x,S.y+ht[q],S.z)):(j.up.set(0,$[q],0),j.position.set(S.x,S.y,S.z),j.lookAt(S.x,S.y,S.z+ht[q]));const ot=this._cubeSize;Qi(E,H*ot,q>2?ot:0,ot,ot),K.setRenderTarget(E),b&&K.render(nt,j),K.render(p,j)}K.toneMapping=G,K.autoClear=F,p.background=z}_textureToCubeUV(p,h){const x=this._renderer,E=p.mapping===f.hy7||p.mapping===f.xFO;E?(this._cubemapMaterial===null&&(this._cubemapMaterial=rc()),this._cubemapMaterial.uniforms.flipEnvMap.value=p.isRenderTargetTexture===!1?-1:1):this._equirectMaterial===null&&(this._equirectMaterial=sc());const S=E?this._cubemapMaterial:this._equirectMaterial,I=this._lodMeshes[0];I.material=S;const V=S.uniforms;V.envMap.value=p;const j=this._cubeSize;Qi(h,0,0,3*j,2*j),x.setRenderTarget(h),x.render(I,Un)}_applyPMREM(p){const h=this._renderer,x=h.autoClear;h.autoClear=!1;const E=this._lodMeshes.length;for(let S=1;S<E;S++)this._applyGGXFilter(p,S-1,S);h.autoClear=x}_applyGGXFilter(p,h,x){const E=this._renderer,S=this._pingPongRenderTarget,I=this._ggxMaterial,V=this._lodMeshes[x];V.material=I;const j=I.uniforms,$=x/(this._lodMeshes.length-1),ht=h/(this._lodMeshes.length-1),K=Math.sqrt($*$-ht*ht),F=0+$*1.25,G=K*F,{_lodMax:X}=this,nt=this._sizeLods[x],A=3*nt*(x>X-ri?x-X+ri:0),b=4*(this._cubeSize-nt);j.envMap.value=p.texture,j.roughness.value=G,j.mipInt.value=X-h,Qi(S,A,b,3*nt,2*nt),E.setRenderTarget(S),E.render(V,Un),j.envMap.value=S.texture,j.roughness.value=0,j.mipInt.value=X-x,Qi(p,A,b,3*nt,2*nt),E.setRenderTarget(p),E.render(V,Un)}_blur(p,h,x,E,S){const I=this._pingPongRenderTarget;this._halfBlur(p,I,h,x,E,"latitudinal",S),this._halfBlur(I,p,x,x,E,"longitudinal",S)}_halfBlur(p,h,x,E,S,I,V){const j=this._renderer,$=this._blurMaterial;I!=="latitudinal"&&I!=="longitudinal"&&(0,f.z3S)("blur direction must be either latitudinal or longitudinal!");const ht=3,K=this._lodMeshes[E];K.material=$;const F=$.uniforms,G=this._sizeLods[x]-1,X=isFinite(S)?Math.PI/(2*G):2*Math.PI/(2*Ri-1),nt=S/X,A=isFinite(S)?1+Math.floor(ht*nt):Ri;A>Ri&&(0,f.R8M)(`sigmaRadians, ${S}, is too large and will clip, as it requested ${A} samples when the maximum is set to ${Ri}`);const b=[];let z=0;for(let st=0;st<Ri;++st){const R=st/nt,U=Math.exp(-R*R/2);b.push(U),st===0?z+=U:st<A&&(z+=2*U)}for(let st=0;st<b.length;st++)b[st]=b[st]/z;F.envMap.value=p.texture,F.samples.value=A,F.weights.value=b,F.latitudinal.value=I==="latitudinal",V&&(F.poleAxis.value=V);const{_lodMax:q}=this;F.dTheta.value=X,F.mipInt.value=q-x;const H=this._sizeLods[E],ot=3*H*(E>q-ri?E-q+ri:0),J=4*(this._cubeSize-H);Qi(h,ot,J,3*H,2*H),j.setRenderTarget(h),j.render(K,Un)}}function Vh(o){const p=[],h=[],x=[];let E=o;const S=o-ri+1+ec.length;for(let I=0;I<S;I++){const V=Math.pow(2,E);p.push(V);let j=1/V;I>o-ri?j=ec[I-o+ri-1]:I===0&&(j=0),h.push(j);const $=1/(V-2),ht=-$,K=1+$,F=[ht,ht,K,ht,K,K,ht,ht,K,K,ht,K],G=6,X=6,nt=3,A=2,b=1,z=new Float32Array(nt*X*G),q=new Float32Array(A*X*G),H=new Float32Array(b*X*G);for(let J=0;J<G;J++){const st=J%3*2/3-1,R=J>2?0:-1,U=[st,R,0,st+2/3,R,0,st+2/3,R+1,0,st,R,0,st+2/3,R+1,0,st,R+1,0];z.set(U,nt*X*J),q.set(F,A*X*J);const Tt=[J,J,J,J,J,J];H.set(Tt,b*X*J)}const ot=new f.LoY;ot.setAttribute("position",new f.THS(z,nt)),ot.setAttribute("uv",new f.THS(q,A)),ot.setAttribute("faceIndex",new f.THS(H,b)),x.push(new f.eaF(ot,null)),E>ri&&E--}return{lodMeshes:x,sizeLods:p,sigmas:h}}function ic(o,p,h){const x=new f.nWS(o,p,h);return x.texture.mapping=f.Om,x.texture.name="PMREM.cubeUv",x.scissorTest=!0,x}function Qi(o,p,h,x,E){o.viewport.set(p,h,x,E),o.scissor.set(p,h,x,E)}function Ii(o,p,h){return new f.BKk({name:"PMREMGGXConvolution",defines:{GGX_SAMPLES:zh,CUBEUV_TEXEL_WIDTH:1/p,CUBEUV_TEXEL_HEIGHT:1/h,CUBEUV_MAX_MIP:`${o}.0`},uniforms:{envMap:{value:null},roughness:{value:0},mipInt:{value:0}},vertexShader:dr(),fragmentShader:`

			precision highp float;
			precision highp int;

			varying vec3 vOutputDirection;

			uniform sampler2D envMap;
			uniform float roughness;
			uniform float mipInt;

			#define ENVMAP_TYPE_CUBE_UV
			#include <cube_uv_reflection_fragment>

			#define PI 3.14159265359

			// Van der Corput radical inverse
			float radicalInverse_VdC(uint bits) {
				bits = (bits << 16u) | (bits >> 16u);
				bits = ((bits & 0x55555555u) << 1u) | ((bits & 0xAAAAAAAAu) >> 1u);
				bits = ((bits & 0x33333333u) << 2u) | ((bits & 0xCCCCCCCCu) >> 2u);
				bits = ((bits & 0x0F0F0F0Fu) << 4u) | ((bits & 0xF0F0F0F0u) >> 4u);
				bits = ((bits & 0x00FF00FFu) << 8u) | ((bits & 0xFF00FF00u) >> 8u);
				return float(bits) * 2.3283064365386963e-10; // / 0x100000000
			}

			// Hammersley sequence
			vec2 hammersley(uint i, uint N) {
				return vec2(float(i) / float(N), radicalInverse_VdC(i));
			}

			// GGX VNDF importance sampling (Eric Heitz 2018)
			// "Sampling the GGX Distribution of Visible Normals"
			// https://jcgt.org/published/0007/04/01/
			vec3 importanceSampleGGX_VNDF(vec2 Xi, vec3 V, float roughness) {
				float alpha = roughness * roughness;

				// Section 4.1: Orthonormal basis
				vec3 T1 = vec3(1.0, 0.0, 0.0);
				vec3 T2 = cross(V, T1);

				// Section 4.2: Parameterization of projected area
				float r = sqrt(Xi.x);
				float phi = 2.0 * PI * Xi.y;
				float t1 = r * cos(phi);
				float t2 = r * sin(phi);
				float s = 0.5 * (1.0 + V.z);
				t2 = (1.0 - s) * sqrt(1.0 - t1 * t1) + s * t2;

				// Section 4.3: Reprojection onto hemisphere
				vec3 Nh = t1 * T1 + t2 * T2 + sqrt(max(0.0, 1.0 - t1 * t1 - t2 * t2)) * V;

				// Section 3.4: Transform back to ellipsoid configuration
				return normalize(vec3(alpha * Nh.x, alpha * Nh.y, max(0.0, Nh.z)));
			}

			void main() {
				vec3 N = normalize(vOutputDirection);
				vec3 V = N; // Assume view direction equals normal for pre-filtering

				vec3 prefilteredColor = vec3(0.0);
				float totalWeight = 0.0;

				// For very low roughness, just sample the environment directly
				if (roughness < 0.001) {
					gl_FragColor = vec4(bilinearCubeUV(envMap, N, mipInt), 1.0);
					return;
				}

				// Tangent space basis for VNDF sampling
				vec3 up = abs(N.z) < 0.999 ? vec3(0.0, 0.0, 1.0) : vec3(1.0, 0.0, 0.0);
				vec3 tangent = normalize(cross(up, N));
				vec3 bitangent = cross(N, tangent);

				for(uint i = 0u; i < uint(GGX_SAMPLES); i++) {
					vec2 Xi = hammersley(i, uint(GGX_SAMPLES));

					// For PMREM, V = N, so in tangent space V is always (0, 0, 1)
					vec3 H_tangent = importanceSampleGGX_VNDF(Xi, vec3(0.0, 0.0, 1.0), roughness);

					// Transform H back to world space
					vec3 H = normalize(tangent * H_tangent.x + bitangent * H_tangent.y + N * H_tangent.z);
					vec3 L = normalize(2.0 * dot(V, H) * H - V);

					float NdotL = max(dot(N, L), 0.0);

					if(NdotL > 0.0) {
						// Sample environment at fixed mip level
						// VNDF importance sampling handles the distribution filtering
						vec3 sampleColor = bilinearCubeUV(envMap, L, mipInt);

						// Weight by NdotL for the split-sum approximation
						// VNDF PDF naturally accounts for the visible microfacet distribution
						prefilteredColor += sampleColor * NdotL;
						totalWeight += NdotL;
					}
				}

				if (totalWeight > 0.0) {
					prefilteredColor = prefilteredColor / totalWeight;
				}

				gl_FragColor = vec4(prefilteredColor, 1.0);
			}
		`,blending:f.XIg,depthTest:!1,depthWrite:!1})}function kh(o,p,h){const x=new Float32Array(Ri),E=new f.Pq0(0,1,0);return new f.BKk({name:"SphericalGaussianBlur",defines:{n:Ri,CUBEUV_TEXEL_WIDTH:1/p,CUBEUV_TEXEL_HEIGHT:1/h,CUBEUV_MAX_MIP:`${o}.0`},uniforms:{envMap:{value:null},samples:{value:1},weights:{value:x},latitudinal:{value:!1},dTheta:{value:0},mipInt:{value:0},poleAxis:{value:E}},vertexShader:dr(),fragmentShader:`

			precision mediump float;
			precision mediump int;

			varying vec3 vOutputDirection;

			uniform sampler2D envMap;
			uniform int samples;
			uniform float weights[ n ];
			uniform bool latitudinal;
			uniform float dTheta;
			uniform float mipInt;
			uniform vec3 poleAxis;

			#define ENVMAP_TYPE_CUBE_UV
			#include <cube_uv_reflection_fragment>

			vec3 getSample( float theta, vec3 axis ) {

				float cosTheta = cos( theta );
				// Rodrigues' axis-angle rotation
				vec3 sampleDirection = vOutputDirection * cosTheta
					+ cross( axis, vOutputDirection ) * sin( theta )
					+ axis * dot( axis, vOutputDirection ) * ( 1.0 - cosTheta );

				return bilinearCubeUV( envMap, sampleDirection, mipInt );

			}

			void main() {

				vec3 axis = latitudinal ? poleAxis : cross( poleAxis, vOutputDirection );

				if ( all( equal( axis, vec3( 0.0 ) ) ) ) {

					axis = vec3( vOutputDirection.z, 0.0, - vOutputDirection.x );

				}

				axis = normalize( axis );

				gl_FragColor = vec4( 0.0, 0.0, 0.0, 1.0 );
				gl_FragColor.rgb += weights[ 0 ] * getSample( 0.0, axis );

				for ( int i = 1; i < n; i++ ) {

					if ( i >= samples ) {

						break;

					}

					float theta = dTheta * float( i );
					gl_FragColor.rgb += weights[ i ] * getSample( -1.0 * theta, axis );
					gl_FragColor.rgb += weights[ i ] * getSample( theta, axis );

				}

			}
		`,blending:f.XIg,depthTest:!1,depthWrite:!1})}function sc(){return new f.BKk({name:"EquirectangularToCubeUV",uniforms:{envMap:{value:null}},vertexShader:dr(),fragmentShader:`

			precision mediump float;
			precision mediump int;

			varying vec3 vOutputDirection;

			uniform sampler2D envMap;

			#include <common>

			void main() {

				vec3 outputDirection = normalize( vOutputDirection );
				vec2 uv = equirectUv( outputDirection );

				gl_FragColor = vec4( texture2D ( envMap, uv ).rgb, 1.0 );

			}
		`,blending:f.XIg,depthTest:!1,depthWrite:!1})}function rc(){return new f.BKk({name:"CubemapToCubeUV",uniforms:{envMap:{value:null},flipEnvMap:{value:-1}},vertexShader:dr(),fragmentShader:`

			precision mediump float;
			precision mediump int;

			uniform float flipEnvMap;

			varying vec3 vOutputDirection;

			uniform samplerCube envMap;

			void main() {

				gl_FragColor = textureCube( envMap, vec3( flipEnvMap * vOutputDirection.x, vOutputDirection.yz ) );

			}
		`,blending:f.XIg,depthTest:!1,depthWrite:!1})}function dr(){return`

		precision mediump float;
		precision mediump int;

		attribute float faceIndex;

		varying vec3 vOutputDirection;

		// RH coordinate system; PMREM face-indexing convention
		vec3 getDirection( vec2 uv, float face ) {

			uv = 2.0 * uv - 1.0;

			vec3 direction = vec3( uv, 1.0 );

			if ( face == 0.0 ) {

				direction = direction.zyx; // ( 1, v, u ) pos x

			} else if ( face == 1.0 ) {

				direction = direction.xzy;
				direction.xz *= -1.0; // ( -u, 1, -v ) pos y

			} else if ( face == 2.0 ) {

				direction.x *= -1.0; // ( -u, v, 1 ) pos z

			} else if ( face == 3.0 ) {

				direction = direction.zyx;
				direction.xz *= -1.0; // ( -1, v, -u ) neg x

			} else if ( face == 4.0 ) {

				direction = direction.xzy;
				direction.xy *= -1.0; // ( -u, -1, v ) neg y

			} else if ( face == 5.0 ) {

				direction.z *= -1.0; // ( u, v, -1 ) neg z

			}

			return direction;

		}

		void main() {

			vOutputDirection = getDirection( uv, faceIndex );
			gl_Position = vec4( position, 1.0 );

		}
	`}class ac extends f.nWS{constructor(p=1,h={}){super(p,p,h),this.isWebGLCubeRenderTarget=!0;const x={width:p,height:p,depth:1},E=[x,x,x,x,x,x];this.texture=new f.b4q(E),this._setTextureOptions(h),this.texture.isRenderTargetTexture=!0}fromEquirectangularTexture(p,h){this.texture.type=h.type,this.texture.colorSpace=h.colorSpace,this.texture.generateMipmaps=h.generateMipmaps,this.texture.minFilter=h.minFilter,this.texture.magFilter=h.magFilter;const x={uniforms:{tEquirect:{value:null}},vertexShader:`

				varying vec3 vWorldDirection;

				vec3 transformDirection( in vec3 dir, in mat4 matrix ) {

					return normalize( ( matrix * vec4( dir, 0.0 ) ).xyz );

				}

				void main() {

					vWorldDirection = transformDirection( position, modelMatrix );

					#include <begin_vertex>
					#include <project_vertex>

				}
			`,fragmentShader:`

				uniform sampler2D tEquirect;

				varying vec3 vWorldDirection;

				#include <common>

				void main() {

					vec3 direction = normalize( vWorldDirection );

					vec2 sampleUV = equirectUv( direction );

					gl_FragColor = texture2D( tEquirect, sampleUV );

				}
			`},E=new f.iNn(5,5,5),S=new f.BKk({name:"CubemapFromEquirect",uniforms:(0,f.lxW)(x.uniforms),vertexShader:x.vertexShader,fragmentShader:x.fragmentShader,side:f.hsX,blending:f.XIg});S.uniforms.tEquirect.value=h;const I=new f.eaF(E,S),V=h.minFilter;return h.minFilter===f.$_I&&(h.minFilter=f.k6q),new f.F1T(1,10,this).update(p,I),h.minFilter=V,I.geometry.dispose(),I.material.dispose(),this}clear(p,h=!0,x=!0,E=!0){const S=p.getRenderTarget();for(let I=0;I<6;I++)p.setRenderTarget(this,I),p.clear(h,x,E);p.setRenderTarget(S)}}function Gh(o){let p=new WeakMap,h=new WeakMap,x=null;function E(F,G=!1){return F==null?null:G?I(F):S(F)}function S(F){if(F&&F.isTexture){const G=F.mapping;if(G===f.wfO||G===f.uV5)if(p.has(F)){const X=p.get(F).texture;return V(X,F.mapping)}else{const X=F.image;if(X&&X.height>0){const nt=new ac(X.height);return nt.fromEquirectangularTexture(o,F),p.set(F,nt),F.addEventListener("dispose",$),V(nt.texture,F.mapping)}else return null}}return F}function I(F){if(F&&F.isTexture){const G=F.mapping,X=G===f.wfO||G===f.uV5,nt=G===f.hy7||G===f.xFO;if(X||nt){let A=h.get(F);const b=A!==void 0?A.texture.pmremVersion:0;if(F.isRenderTargetTexture&&F.pmremVersion!==b)return x===null&&(x=new nc(o)),A=X?x.fromEquirectangular(F,A):x.fromCubemap(F,A),A.texture.pmremVersion=F.pmremVersion,h.set(F,A),A.texture;if(A!==void 0)return A.texture;{const z=F.image;return X&&z&&z.height>0||nt&&z&&j(z)?(x===null&&(x=new nc(o)),A=X?x.fromEquirectangular(F):x.fromCubemap(F),A.texture.pmremVersion=F.pmremVersion,h.set(F,A),F.addEventListener("dispose",ht),A.texture):null}}}return F}function V(F,G){return G===f.wfO?F.mapping=f.hy7:G===f.uV5&&(F.mapping=f.xFO),F}function j(F){let G=0;const X=6;for(let nt=0;nt<X;nt++)F[nt]!==void 0&&G++;return G===X}function $(F){const G=F.target;G.removeEventListener("dispose",$);const X=p.get(G);X!==void 0&&(p.delete(G),X.dispose())}function ht(F){const G=F.target;G.removeEventListener("dispose",ht);const X=h.get(G);X!==void 0&&(h.delete(G),X.dispose())}function K(){p=new WeakMap,h=new WeakMap,x!==null&&(x.dispose(),x=null)}return{get:E,dispose:K}}function Hh(o){const p={};function h(x){if(p[x]!==void 0)return p[x];const E=o.getExtension(x);return p[x]=E,E}return{has:function(x){return h(x)!==null},init:function(){h("EXT_color_buffer_float"),h("WEBGL_clip_cull_distance"),h("OES_texture_float_linear"),h("EXT_color_buffer_half_float"),h("WEBGL_multisampled_render_to_texture"),h("WEBGL_render_shared_exponent")},get:function(x){const E=h(x);return E===null&&(0,f.mcG)("WebGLRenderer: "+x+" extension not supported."),E}}}function Wh(o,p,h,x){const E={},S=new WeakMap;function I(K){const F=K.target;F.index!==null&&p.remove(F.index);for(const X in F.attributes)p.remove(F.attributes[X]);F.removeEventListener("dispose",I),delete E[F.id];const G=S.get(F);G&&(p.remove(G),S.delete(F)),x.releaseStatesOfGeometry(F),F.isInstancedBufferGeometry===!0&&delete F._maxInstanceCount,h.memory.geometries--}function V(K,F){return E[F.id]===!0||(F.addEventListener("dispose",I),E[F.id]=!0,h.memory.geometries++),F}function j(K){const F=K.attributes;for(const G in F)p.update(F[G],o.ARRAY_BUFFER)}function $(K){const F=[],G=K.index,X=K.attributes.position;let nt=0;if(X===void 0)return;if(G!==null){const z=G.array;nt=G.version;for(let q=0,H=z.length;q<H;q+=3){const ot=z[q+0],J=z[q+1],st=z[q+2];F.push(ot,J,J,st,st,ot)}}else{const z=X.array;nt=X.version;for(let q=0,H=z.length/3-1;q<H;q+=3){const ot=q+0,J=q+1,st=q+2;F.push(ot,J,J,st,st,ot)}}const A=new(X.count>=65535?f.MW4:f.A$4)(F,1);A.version=nt;const b=S.get(K);b&&p.remove(b),S.set(K,A)}function ht(K){const F=S.get(K);if(F){const G=K.index;G!==null&&F.version<G.version&&$(K)}else $(K);return S.get(K)}return{get:V,update:j,getWireframeAttribute:ht}}function Xh(o,p,h){let x;function E(F){x=F}let S,I;function V(F){S=F.type,I=F.bytesPerElement}function j(F,G){o.drawElements(x,G,S,F*I),h.update(G,x,1)}function $(F,G,X){X!==0&&(o.drawElementsInstanced(x,G,S,F*I,X),h.update(G,x,X))}function ht(F,G,X){if(X===0)return;p.get("WEBGL_multi_draw").multiDrawElementsWEBGL(x,G,0,S,F,0,X);let A=0;for(let b=0;b<X;b++)A+=G[b];h.update(A,x,1)}function K(F,G,X,nt){if(X===0)return;const A=p.get("WEBGL_multi_draw");if(A===null)for(let b=0;b<F.length;b++)$(F[b]/I,G[b],nt[b]);else{A.multiDrawElementsInstancedWEBGL(x,G,0,S,F,0,nt,0,X);let b=0;for(let z=0;z<X;z++)b+=G[z]*nt[z];h.update(b,x,1)}}this.setMode=E,this.setIndex=V,this.render=j,this.renderInstances=$,this.renderMultiDraw=ht,this.renderMultiDrawInstances=K}function qh(o){const p={geometries:0,textures:0},h={frame:0,calls:0,triangles:0,points:0,lines:0};function x(S,I,V){switch(h.calls++,I){case o.TRIANGLES:h.triangles+=V*(S/3);break;case o.LINES:h.lines+=V*(S/2);break;case o.LINE_STRIP:h.lines+=V*(S-1);break;case o.LINE_LOOP:h.lines+=V*S;break;case o.POINTS:h.points+=V*S;break;default:(0,f.z3S)("WebGLInfo: Unknown draw mode:",I);break}}function E(){h.calls=0,h.triangles=0,h.points=0,h.lines=0}return{memory:p,render:h,programs:null,autoReset:!0,reset:E,update:x}}function Yh(o,p,h){const x=new WeakMap,E=new f.IUQ;function S(I,V,j){const $=I.morphTargetInfluences,ht=V.morphAttributes.position||V.morphAttributes.normal||V.morphAttributes.color,K=ht!==void 0?ht.length:0;let F=x.get(V);if(F===void 0||F.count!==K){let U=function(){st.dispose(),x.delete(V),V.removeEventListener("dispose",U)};F!==void 0&&F.texture.dispose();const G=V.morphAttributes.position!==void 0,X=V.morphAttributes.normal!==void 0,nt=V.morphAttributes.color!==void 0,A=V.morphAttributes.position||[],b=V.morphAttributes.normal||[],z=V.morphAttributes.color||[];let q=0;G===!0&&(q=1),X===!0&&(q=2),nt===!0&&(q=3);let H=V.attributes.position.count*q,ot=1;H>p.maxTextureSize&&(ot=Math.ceil(H/p.maxTextureSize),H=p.maxTextureSize);const J=new Float32Array(H*ot*4*K),st=new f.rFo(J,H,ot,K);st.type=f.RQf,st.needsUpdate=!0;const R=q*4;for(let Tt=0;Tt<K;Tt++){const Z=A[Tt],dt=b[Tt],mt=z[Tt],gt=H*ot*4*Tt;for(let _t=0;_t<Z.count;_t++){const ct=_t*R;G===!0&&(E.fromBufferAttribute(Z,_t),J[gt+ct+0]=E.x,J[gt+ct+1]=E.y,J[gt+ct+2]=E.z,J[gt+ct+3]=0),X===!0&&(E.fromBufferAttribute(dt,_t),J[gt+ct+4]=E.x,J[gt+ct+5]=E.y,J[gt+ct+6]=E.z,J[gt+ct+7]=0),nt===!0&&(E.fromBufferAttribute(mt,_t),J[gt+ct+8]=E.x,J[gt+ct+9]=E.y,J[gt+ct+10]=E.z,J[gt+ct+11]=mt.itemSize===4?E.w:1)}}F={count:K,texture:st,size:new f.I9Y(H,ot)},x.set(V,F),V.addEventListener("dispose",U)}if(I.isInstancedMesh===!0&&I.morphTexture!==null)j.getUniforms().setValue(o,"morphTexture",I.morphTexture,h);else{let G=0;for(let nt=0;nt<$.length;nt++)G+=$[nt];const X=V.morphTargetsRelative?1:1-G;j.getUniforms().setValue(o,"morphTargetBaseInfluence",X),j.getUniforms().setValue(o,"morphTargetInfluences",$)}j.getUniforms().setValue(o,"morphTargetsTexture",F.texture,h),j.getUniforms().setValue(o,"morphTargetsTextureSize",F.size)}return{update:S}}function Zh(o,p,h,x,E){let S=new WeakMap;function I($){const ht=E.render.frame,K=$.geometry,F=p.get($,K);if(S.get(F)!==ht&&(p.update(F),S.set(F,ht)),$.isInstancedMesh&&($.hasEventListener("dispose",j)===!1&&$.addEventListener("dispose",j),S.get($)!==ht&&(h.update($.instanceMatrix,o.ARRAY_BUFFER),$.instanceColor!==null&&h.update($.instanceColor,o.ARRAY_BUFFER),S.set($,ht))),$.isSkinnedMesh){const G=$.skeleton;S.get(G)!==ht&&(G.update(),S.set(G,ht))}return F}function V(){S=new WeakMap}function j($){const ht=$.target;ht.removeEventListener("dispose",j),x.releaseStatesOfObject(ht),h.remove(ht.instanceMatrix),ht.instanceColor!==null&&h.remove(ht.instanceColor)}return{update:I,dispose:V}}const $h={[f.kyO]:"LINEAR_TONE_MAPPING",[f.Mjd]:"REINHARD_TONE_MAPPING",[f.nNL]:"CINEON_TONE_MAPPING",[f.FV]:"ACES_FILMIC_TONE_MAPPING",[f.LAk]:"AGX_TONE_MAPPING",[f.aJ8]:"NEUTRAL_TONE_MAPPING",[f.g7M]:"CUSTOM_TONE_MAPPING"};function Aa(o,p,h,x,E){const S=new f.nWS(p,h,{type:o,depthBuffer:x,stencilBuffer:E}),I=new f.nWS(p,h,{type:f.ix0,depthBuffer:!1,stencilBuffer:!1}),V=new f.LoY;V.setAttribute("position",new f.qtW([-1,3,0,-1,-1,0,3,-1,0],3)),V.setAttribute("uv",new f.qtW([0,2,0,0,2,0],2));const j=new f.D$Q({uniforms:{tDiffuse:{value:null}},vertexShader:`
			precision highp float;

			uniform mat4 modelViewMatrix;
			uniform mat4 projectionMatrix;

			attribute vec3 position;
			attribute vec2 uv;

			varying vec2 vUv;

			void main() {
				vUv = uv;
				gl_Position = projectionMatrix * modelViewMatrix * vec4( position, 1.0 );
			}`,fragmentShader:`
			precision highp float;

			uniform sampler2D tDiffuse;

			varying vec2 vUv;

			#include <tonemapping_pars_fragment>
			#include <colorspace_pars_fragment>

			void main() {
				gl_FragColor = texture2D( tDiffuse, vUv );

				#ifdef LINEAR_TONE_MAPPING
					gl_FragColor.rgb = LinearToneMapping( gl_FragColor.rgb );
				#elif defined( REINHARD_TONE_MAPPING )
					gl_FragColor.rgb = ReinhardToneMapping( gl_FragColor.rgb );
				#elif defined( CINEON_TONE_MAPPING )
					gl_FragColor.rgb = CineonToneMapping( gl_FragColor.rgb );
				#elif defined( ACES_FILMIC_TONE_MAPPING )
					gl_FragColor.rgb = ACESFilmicToneMapping( gl_FragColor.rgb );
				#elif defined( AGX_TONE_MAPPING )
					gl_FragColor.rgb = AgXToneMapping( gl_FragColor.rgb );
				#elif defined( NEUTRAL_TONE_MAPPING )
					gl_FragColor.rgb = NeutralToneMapping( gl_FragColor.rgb );
				#elif defined( CUSTOM_TONE_MAPPING )
					gl_FragColor.rgb = CustomToneMapping( gl_FragColor.rgb );
				#endif

				#ifdef SRGB_TRANSFER
					gl_FragColor = sRGBTransferOETF( gl_FragColor );
				#endif
			}`,depthTest:!1,depthWrite:!1}),$=new f.eaF(V,j),ht=new f.qUd(-1,1,1,-1,0,1);let K=null,F=null,G=!1,X,nt=null,A=[],b=!1;this.setSize=function(z,q){S.setSize(z,q),I.setSize(z,q);for(let H=0;H<A.length;H++){const ot=A[H];ot.setSize&&ot.setSize(z,q)}},this.setEffects=function(z){A=z,b=A.length>0&&A[0].isRenderPass===!0;const q=S.width,H=S.height;for(let ot=0;ot<A.length;ot++){const J=A[ot];J.setSize&&J.setSize(q,H)}},this.begin=function(z,q){if(G||z.toneMapping===f.y_p&&A.length===0)return!1;if(nt=q,q!==null){const H=q.width,ot=q.height;(S.width!==H||S.height!==ot)&&this.setSize(H,ot)}return b===!1&&z.setRenderTarget(S),X=z.toneMapping,z.toneMapping=f.y_p,!0},this.hasRenderPass=function(){return b},this.end=function(z,q){z.toneMapping=X,G=!0;let H=S,ot=I;for(let J=0;J<A.length;J++){const st=A[J];if(st.enabled!==!1&&(st.render(z,ot,H,q),st.needsSwap!==!1)){const R=H;H=ot,ot=R}}if(K!==z.outputColorSpace||F!==z.toneMapping){K=z.outputColorSpace,F=z.toneMapping,j.defines={},f.ppV.getTransfer(K)===f.KLL&&(j.defines.SRGB_TRANSFER="");const J=$h[F];J&&(j.defines[J]=""),j.needsUpdate=!0}j.uniforms.tDiffuse.value=H.texture,z.setRenderTarget(nt),z.render($,ht),nt=null,G=!1},this.isCompositing=function(){return G},this.dispose=function(){S.dispose(),I.dispose(),V.dispose(),j.dispose()}}const Ea=new f.gPd,pr=new f.VCu(1,1),wa=new f.rFo,Ca=new f.dYF,Ra=new f.b4q,Ia=[],Pa=[],La=new Float32Array(16),ji=new Float32Array(9),oc=new Float32Array(4);function ts(o,p,h){const x=o[0];if(x<=0||x>0)return o;const E=p*h;let S=Ia[E];if(S===void 0&&(S=new Float32Array(E),Ia[E]=S),p!==0){x.toArray(S,0);for(let I=1,V=0;I!==p;++I)V+=h,o[I].toArray(S,V)}return S}function on(o,p){if(o.length!==p.length)return!1;for(let h=0,x=o.length;h<x;h++)if(o[h]!==p[h])return!1;return!0}function ln(o,p){for(let h=0,x=p.length;h<x;h++)o[h]=p[h]}function mr(o,p){let h=Pa[p];h===void 0&&(h=new Int32Array(p),Pa[p]=h);for(let x=0;x!==p;++x)h[x]=o.allocateTextureUnit();return h}function Jh(o,p){const h=this.cache;h[0]!==p&&(o.uniform1f(this.addr,p),h[0]=p)}function Kh(o,p){const h=this.cache;if(p.x!==void 0)(h[0]!==p.x||h[1]!==p.y)&&(o.uniform2f(this.addr,p.x,p.y),h[0]=p.x,h[1]=p.y);else{if(on(h,p))return;o.uniform2fv(this.addr,p),ln(h,p)}}function Qh(o,p){const h=this.cache;if(p.x!==void 0)(h[0]!==p.x||h[1]!==p.y||h[2]!==p.z)&&(o.uniform3f(this.addr,p.x,p.y,p.z),h[0]=p.x,h[1]=p.y,h[2]=p.z);else if(p.r!==void 0)(h[0]!==p.r||h[1]!==p.g||h[2]!==p.b)&&(o.uniform3f(this.addr,p.r,p.g,p.b),h[0]=p.r,h[1]=p.g,h[2]=p.b);else{if(on(h,p))return;o.uniform3fv(this.addr,p),ln(h,p)}}function jh(o,p){const h=this.cache;if(p.x!==void 0)(h[0]!==p.x||h[1]!==p.y||h[2]!==p.z||h[3]!==p.w)&&(o.uniform4f(this.addr,p.x,p.y,p.z,p.w),h[0]=p.x,h[1]=p.y,h[2]=p.z,h[3]=p.w);else{if(on(h,p))return;o.uniform4fv(this.addr,p),ln(h,p)}}function lc(o,p){const h=this.cache,x=p.elements;if(x===void 0){if(on(h,p))return;o.uniformMatrix2fv(this.addr,!1,p),ln(h,p)}else{if(on(h,x))return;oc.set(x),o.uniformMatrix2fv(this.addr,!1,oc),ln(h,x)}}function Nn(o,p){const h=this.cache,x=p.elements;if(x===void 0){if(on(h,p))return;o.uniformMatrix3fv(this.addr,!1,p),ln(h,p)}else{if(on(h,x))return;ji.set(x),o.uniformMatrix3fv(this.addr,!1,ji),ln(h,x)}}function Pi(o,p){const h=this.cache,x=p.elements;if(x===void 0){if(on(h,p))return;o.uniformMatrix4fv(this.addr,!1,p),ln(h,p)}else{if(on(h,x))return;La.set(x),o.uniformMatrix4fv(this.addr,!1,La),ln(h,x)}}function tu(o,p){const h=this.cache;h[0]!==p&&(o.uniform1i(this.addr,p),h[0]=p)}function eu(o,p){const h=this.cache;if(p.x!==void 0)(h[0]!==p.x||h[1]!==p.y)&&(o.uniform2i(this.addr,p.x,p.y),h[0]=p.x,h[1]=p.y);else{if(on(h,p))return;o.uniform2iv(this.addr,p),ln(h,p)}}function nu(o,p){const h=this.cache;if(p.x!==void 0)(h[0]!==p.x||h[1]!==p.y||h[2]!==p.z)&&(o.uniform3i(this.addr,p.x,p.y,p.z),h[0]=p.x,h[1]=p.y,h[2]=p.z);else{if(on(h,p))return;o.uniform3iv(this.addr,p),ln(h,p)}}function iu(o,p){const h=this.cache;if(p.x!==void 0)(h[0]!==p.x||h[1]!==p.y||h[2]!==p.z||h[3]!==p.w)&&(o.uniform4i(this.addr,p.x,p.y,p.z,p.w),h[0]=p.x,h[1]=p.y,h[2]=p.z,h[3]=p.w);else{if(on(h,p))return;o.uniform4iv(this.addr,p),ln(h,p)}}function cc(o,p){const h=this.cache;h[0]!==p&&(o.uniform1ui(this.addr,p),h[0]=p)}function hc(o,p){const h=this.cache;if(p.x!==void 0)(h[0]!==p.x||h[1]!==p.y)&&(o.uniform2ui(this.addr,p.x,p.y),h[0]=p.x,h[1]=p.y);else{if(on(h,p))return;o.uniform2uiv(this.addr,p),ln(h,p)}}function Li(o,p){const h=this.cache;if(p.x!==void 0)(h[0]!==p.x||h[1]!==p.y||h[2]!==p.z)&&(o.uniform3ui(this.addr,p.x,p.y,p.z),h[0]=p.x,h[1]=p.y,h[2]=p.z);else{if(on(h,p))return;o.uniform3uiv(this.addr,p),ln(h,p)}}function Da(o,p){const h=this.cache;if(p.x!==void 0)(h[0]!==p.x||h[1]!==p.y||h[2]!==p.z||h[3]!==p.w)&&(o.uniform4ui(this.addr,p.x,p.y,p.z,p.w),h[0]=p.x,h[1]=p.y,h[2]=p.z,h[3]=p.w);else{if(on(h,p))return;o.uniform4uiv(this.addr,p),ln(h,p)}}function es(o,p,h){const x=this.cache,E=h.allocateTextureUnit();x[0]!==E&&(o.uniform1i(this.addr,E),x[0]=E);let S;this.type===o.SAMPLER_2D_SHADOW?(pr.compareFunction=h.isReversedDepthBuffer()?f.gWB:f.TiK,S=pr):S=Ea,h.setTexture2D(p||S,E)}function uc(o,p,h){const x=this.cache,E=h.allocateTextureUnit();x[0]!==E&&(o.uniform1i(this.addr,E),x[0]=E),h.setTexture3D(p||Ca,E)}function Ua(o,p,h){const x=this.cache,E=h.allocateTextureUnit();x[0]!==E&&(o.uniform1i(this.addr,E),x[0]=E),h.setTextureCube(p||Ra,E)}function ai(o,p,h){const x=this.cache,E=h.allocateTextureUnit();x[0]!==E&&(o.uniform1i(this.addr,E),x[0]=E),h.setTexture2DArray(p||wa,E)}function su(o){switch(o){case 5126:return Jh;case 35664:return Kh;case 35665:return Qh;case 35666:return jh;case 35674:return lc;case 35675:return Nn;case 35676:return Pi;case 5124:case 35670:return tu;case 35667:case 35671:return eu;case 35668:case 35672:return nu;case 35669:case 35673:return iu;case 5125:return cc;case 36294:return hc;case 36295:return Li;case 36296:return Da;case 35678:case 36198:case 36298:case 36306:case 35682:return es;case 35679:case 36299:case 36307:return uc;case 35680:case 36300:case 36308:case 36293:return Ua;case 36289:case 36303:case 36311:case 36292:return ai}}function ru(o,p){o.uniform1fv(this.addr,p)}function gr(o,p){const h=ts(p,this.size,2);o.uniform2fv(this.addr,h)}function Na(o,p){const h=ts(p,this.size,3);o.uniform3fv(this.addr,h)}function le(o,p){const h=ts(p,this.size,4);o.uniform4fv(this.addr,h)}function Re(o,p){const h=ts(p,this.size,4);o.uniformMatrix2fv(this.addr,!1,h)}function _r(o,p){const h=ts(p,this.size,9);o.uniformMatrix3fv(this.addr,!1,h)}function fc(o,p){const h=ts(p,this.size,16);o.uniformMatrix4fv(this.addr,!1,h)}function dc(o,p){o.uniform1iv(this.addr,p)}function Kn(o,p){o.uniform2iv(this.addr,p)}function mn(o,p){o.uniform3iv(this.addr,p)}function Fa(o,p){o.uniform4iv(this.addr,p)}function _i(o,p){o.uniform1uiv(this.addr,p)}function Di(o,p){o.uniform2uiv(this.addr,p)}function An(o,p){o.uniform3uiv(this.addr,p)}function pe(o,p){o.uniform4uiv(this.addr,p)}function xr(o,p,h){const x=this.cache,E=p.length,S=mr(h,E);on(x,S)||(o.uniform1iv(this.addr,S),ln(x,S));let I;this.type===o.SAMPLER_2D_SHADOW?I=pr:I=Ea;for(let V=0;V!==E;++V)h.setTexture2D(p[V]||I,S[V])}function pc(o,p,h){const x=this.cache,E=p.length,S=mr(h,E);on(x,S)||(o.uniform1iv(this.addr,S),ln(x,S));for(let I=0;I!==E;++I)h.setTexture3D(p[I]||Ca,S[I])}function mc(o,p,h){const x=this.cache,E=p.length,S=mr(h,E);on(x,S)||(o.uniform1iv(this.addr,S),ln(x,S));for(let I=0;I!==E;++I)h.setTextureCube(p[I]||Ra,S[I])}function ns(o,p,h){const x=this.cache,E=p.length,S=mr(h,E);on(x,S)||(o.uniform1iv(this.addr,S),ln(x,S));for(let I=0;I!==E;++I)h.setTexture2DArray(p[I]||wa,S[I])}function gc(o){switch(o){case 5126:return ru;case 35664:return gr;case 35665:return Na;case 35666:return le;case 35674:return Re;case 35675:return _r;case 35676:return fc;case 5124:case 35670:return dc;case 35667:case 35671:return Kn;case 35668:case 35672:return mn;case 35669:case 35673:return Fa;case 5125:return _i;case 36294:return Di;case 36295:return An;case 36296:return pe;case 35678:case 36198:case 36298:case 36306:case 35682:return xr;case 35679:case 36299:case 36307:return pc;case 35680:case 36300:case 36308:case 36293:return mc;case 36289:case 36303:case 36311:case 36292:return ns}}class _c{constructor(p,h,x){this.id=p,this.addr=x,this.cache=[],this.type=h.type,this.setValue=su(h.type)}}class xc{constructor(p,h,x){this.id=p,this.addr=x,this.cache=[],this.type=h.type,this.size=h.size,this.setValue=gc(h.type)}}class vc{constructor(p){this.id=p,this.seq=[],this.map={}}setValue(p,h,x){const E=this.seq;for(let S=0,I=E.length;S!==I;++S){const V=E[S];V.setValue(p,h[V.id],x)}}}const vr=/(\w+)(\])?(\[|\.)?/g;function Oa(o,p){o.seq.push(p),o.map[p.id]=p}function yc(o,p,h){const x=o.name,E=x.length;for(vr.lastIndex=0;;){const S=vr.exec(x),I=vr.lastIndex;let V=S[1];const j=S[2]==="]",$=S[3];if(j&&(V=V|0),$===void 0||$==="["&&I+2===E){Oa(h,$===void 0?new _c(V,o,p):new xc(V,o,p));break}else{let K=h.map[V];K===void 0&&(K=new vc(V),Oa(h,K)),h=K}}}class Us{constructor(p,h){this.seq=[],this.map={};const x=p.getProgramParameter(h,p.ACTIVE_UNIFORMS);for(let I=0;I<x;++I){const V=p.getActiveUniform(h,I),j=p.getUniformLocation(h,V.name);yc(V,j,this)}const E=[],S=[];for(const I of this.seq)I.type===p.SAMPLER_2D_SHADOW||I.type===p.SAMPLER_CUBE_SHADOW||I.type===p.SAMPLER_2D_ARRAY_SHADOW?E.push(I):S.push(I);E.length>0&&(this.seq=E.concat(S))}setValue(p,h,x,E){const S=this.map[h];S!==void 0&&S.setValue(p,x,E)}setOptional(p,h,x){const E=h[x];E!==void 0&&this.setValue(p,x,E)}static upload(p,h,x,E){for(let S=0,I=h.length;S!==I;++S){const V=h[S],j=x[V.id];j.needsUpdate!==!1&&V.setValue(p,j.value,E)}}static seqWithValue(p,h){const x=[];for(let E=0,S=p.length;E!==S;++E){const I=p[E];I.id in h&&x.push(I)}return x}}function Ba(o,p,h){const x=o.createShader(p);return o.shaderSource(x,h),o.compileShader(x),x}const Mc=37297;let Sc=0;function bc(o,p){const h=o.split(`
`),x=[],E=Math.max(p-6,0),S=Math.min(p+6,h.length);for(let I=E;I<S;I++){const V=I+1;x.push(`${V===p?">":" "} ${V}: ${h[I]}`)}return x.join(`
`)}const za=new f.dwI;function Tc(o){f.ppV._getMatrix(za,f.ppV.workingColorSpace,o);const p=`mat3( ${za.elements.map(h=>h.toFixed(4))} )`;switch(f.ppV.getTransfer(o)){case f.VxR:return[p,"LinearTransferOETF"];case f.KLL:return[p,"sRGBTransferOETF"];default:return(0,f.R8M)("WebGLProgram: Unsupported color space: ",o),[p,"LinearTransferOETF"]}}function gn(o,p,h){const x=o.getShaderParameter(p,o.COMPILE_STATUS),S=(o.getShaderInfoLog(p)||"").trim();if(x&&S==="")return"";const I=/ERROR: 0:(\d+)/.exec(S);if(I){const V=parseInt(I[1]);return h.toUpperCase()+`

`+S+`

`+bc(o.getShaderSource(p),V)}else return S}function ye(o,p){const h=Tc(p);return[`vec4 ${o}( vec4 value ) {`,`	return ${h[1]}( vec4( value.rgb * ${h[0]}, value.a ) );`,"}"].join(`
`)}const au={[f.kyO]:"Linear",[f.Mjd]:"Reinhard",[f.nNL]:"Cineon",[f.FV]:"ACESFilmic",[f.LAk]:"AgX",[f.aJ8]:"Neutral",[f.g7M]:"Custom"};function At(o,p){const h=au[p];return h===void 0?((0,f.R8M)("WebGLProgram: Unsupported toneMapping:",p),"vec3 "+o+"( vec3 color ) { return LinearToneMapping( color ); }"):"vec3 "+o+"( vec3 color ) { return "+h+"ToneMapping( color ); }"}const un=new f.Pq0;function O(){f.ppV.getLuminanceCoefficients(un);const o=un.x.toFixed(4),p=un.y.toFixed(4),h=un.z.toFixed(4);return["float luminance( const in vec3 rgb ) {",`	const vec3 weights = vec3( ${o}, ${p}, ${h} );`,"	return dot( weights, rgb );","}"].join(`
`)}function yr(o){return[o.extensionClipCullDistance?"#extension GL_ANGLE_clip_cull_distance : require":"",o.extensionMultiDraw?"#extension GL_ANGLE_multi_draw : require":""].filter(xi).join(`
`)}function Va(o){const p=[];for(const h in o){const x=o[h];x!==!1&&p.push("#define "+h+" "+x)}return p.join(`
`)}function Fn(o,p){const h={},x=o.getProgramParameter(p,o.ACTIVE_ATTRIBUTES);for(let E=0;E<x;E++){const S=o.getActiveAttrib(p,E),I=S.name;let V=1;S.type===o.FLOAT_MAT2&&(V=2),S.type===o.FLOAT_MAT3&&(V=3),S.type===o.FLOAT_MAT4&&(V=4),h[I]={type:S.type,location:o.getAttribLocation(p,I),locationSize:V}}return h}function xi(o){return o!==""}function Mr(o,p){const h=p.numSpotLightShadows+p.numSpotLightMaps-p.numSpotLightShadowsWithMaps;return o.replace(/NUM_DIR_LIGHTS/g,p.numDirLights).replace(/NUM_SPOT_LIGHTS/g,p.numSpotLights).replace(/NUM_SPOT_LIGHT_MAPS/g,p.numSpotLightMaps).replace(/NUM_SPOT_LIGHT_COORDS/g,h).replace(/NUM_RECT_AREA_LIGHTS/g,p.numRectAreaLights).replace(/NUM_POINT_LIGHTS/g,p.numPointLights).replace(/NUM_HEMI_LIGHTS/g,p.numHemiLights).replace(/NUM_DIR_LIGHT_SHADOWS/g,p.numDirLightShadows).replace(/NUM_SPOT_LIGHT_SHADOWS_WITH_MAPS/g,p.numSpotLightShadowsWithMaps).replace(/NUM_SPOT_LIGHT_SHADOWS/g,p.numSpotLightShadows).replace(/NUM_POINT_LIGHT_SHADOWS/g,p.numPointLightShadows)}function Sr(o,p){return o.replace(/NUM_CLIPPING_PLANES/g,p.numClippingPlanes).replace(/UNION_CLIPPING_PLANES/g,p.numClippingPlanes-p.numClipIntersection)}const Ac=/^[ \t]*#include +<([\w\d./]+)>/gm;function _n(o){return o.replace(Ac,Ui)}const Qn=new Map;function Ui(o,p){let h=ve[p];if(h===void 0){const x=Qn.get(p);if(x!==void 0)h=ve[x],(0,f.R8M)('WebGLRenderer: Shader chunk "%s" has been deprecated. Use "%s" instead.',p,x);else throw new Error("Can not resolve #include <"+p+">")}return _n(h)}const Ni=/#pragma unroll_loop_start\s+for\s*\(\s*int\s+i\s*=\s*(\d+)\s*;\s*i\s*<\s*(\d+)\s*;\s*i\s*\+\+\s*\)\s*{([\s\S]+?)}\s+#pragma unroll_loop_end/g;function ka(o){return o.replace(Ni,Ec)}function Ec(o,p,h,x){let E="";for(let S=parseInt(p);S<parseInt(h);S++)E+=x.replace(/\[\s*i\s*\]/g,"[ "+S+" ]").replace(/UNROLLED_LOOP_INDEX/g,S);return E}function oi(o){let p=`precision ${o.precision} float;
	precision ${o.precision} int;
	precision ${o.precision} sampler2D;
	precision ${o.precision} samplerCube;
	precision ${o.precision} sampler3D;
	precision ${o.precision} sampler2DArray;
	precision ${o.precision} sampler2DShadow;
	precision ${o.precision} samplerCubeShadow;
	precision ${o.precision} sampler2DArrayShadow;
	precision ${o.precision} isampler2D;
	precision ${o.precision} isampler3D;
	precision ${o.precision} isamplerCube;
	precision ${o.precision} isampler2DArray;
	precision ${o.precision} usampler2D;
	precision ${o.precision} usampler3D;
	precision ${o.precision} usamplerCube;
	precision ${o.precision} usampler2DArray;
	`;return o.precision==="highp"?p+=`
#define HIGH_PRECISION`:o.precision==="mediump"?p+=`
#define MEDIUM_PRECISION`:o.precision==="lowp"&&(p+=`
#define LOW_PRECISION`),p}const br={[f.QP0]:"SHADOWMAP_TYPE_PCF",[f.RyA]:"SHADOWMAP_TYPE_VSM"};function wc(o){return br[o.shadowMapType]||"SHADOWMAP_TYPE_BASIC"}const Tr={[f.hy7]:"ENVMAP_TYPE_CUBE",[f.xFO]:"ENVMAP_TYPE_CUBE",[f.Om]:"ENVMAP_TYPE_CUBE_UV"};function Ye(o){return o.envMap===!1?"ENVMAP_TYPE_CUBE":Tr[o.envMapMode]||"ENVMAP_TYPE_CUBE"}const En={[f.xFO]:"ENVMAP_MODE_REFRACTION"};function Ga(o){return o.envMap===!1?"ENVMAP_MODE_REFLECTION":En[o.envMapMode]||"ENVMAP_MODE_REFLECTION"}const Ar={[f.caT]:"ENVMAP_BLENDING_MULTIPLY",[f.KRh]:"ENVMAP_BLENDING_MIX",[f.XrR]:"ENVMAP_BLENDING_ADD"};function Ha(o){return o.envMap===!1?"ENVMAP_BLENDING_NONE":Ar[o.combine]||"ENVMAP_BLENDING_NONE"}function ou(o){const p=o.envMapCubeUVHeight;if(p===null)return null;const h=Math.log2(p)-2,x=1/p;return{texelWidth:1/(3*Math.max(Math.pow(2,h),112)),texelHeight:x,maxMip:h}}function Er(o,p,h,x){const E=o.getContext(),S=h.defines;let I=h.vertexShader,V=h.fragmentShader;const j=wc(h),$=Ye(h),ht=Ga(h),K=Ha(h),F=ou(h),G=yr(h),X=Va(S),nt=E.createProgram();let A,b,z=h.glslVersion?"#version "+h.glslVersion+`
`:"";h.isRawShaderMaterial?(A=["#define SHADER_TYPE "+h.shaderType,"#define SHADER_NAME "+h.shaderName,X].filter(xi).join(`
`),A.length>0&&(A+=`
`),b=["#define SHADER_TYPE "+h.shaderType,"#define SHADER_NAME "+h.shaderName,X].filter(xi).join(`
`),b.length>0&&(b+=`
`)):(A=[oi(h),"#define SHADER_TYPE "+h.shaderType,"#define SHADER_NAME "+h.shaderName,X,h.extensionClipCullDistance?"#define USE_CLIP_DISTANCE":"",h.batching?"#define USE_BATCHING":"",h.batchingColor?"#define USE_BATCHING_COLOR":"",h.instancing?"#define USE_INSTANCING":"",h.instancingColor?"#define USE_INSTANCING_COLOR":"",h.instancingMorph?"#define USE_INSTANCING_MORPH":"",h.useFog&&h.fog?"#define USE_FOG":"",h.useFog&&h.fogExp2?"#define FOG_EXP2":"",h.map?"#define USE_MAP":"",h.envMap?"#define USE_ENVMAP":"",h.envMap?"#define "+ht:"",h.lightMap?"#define USE_LIGHTMAP":"",h.aoMap?"#define USE_AOMAP":"",h.bumpMap?"#define USE_BUMPMAP":"",h.normalMap?"#define USE_NORMALMAP":"",h.normalMapObjectSpace?"#define USE_NORMALMAP_OBJECTSPACE":"",h.normalMapTangentSpace?"#define USE_NORMALMAP_TANGENTSPACE":"",h.displacementMap?"#define USE_DISPLACEMENTMAP":"",h.emissiveMap?"#define USE_EMISSIVEMAP":"",h.anisotropy?"#define USE_ANISOTROPY":"",h.anisotropyMap?"#define USE_ANISOTROPYMAP":"",h.clearcoatMap?"#define USE_CLEARCOATMAP":"",h.clearcoatRoughnessMap?"#define USE_CLEARCOAT_ROUGHNESSMAP":"",h.clearcoatNormalMap?"#define USE_CLEARCOAT_NORMALMAP":"",h.iridescenceMap?"#define USE_IRIDESCENCEMAP":"",h.iridescenceThicknessMap?"#define USE_IRIDESCENCE_THICKNESSMAP":"",h.specularMap?"#define USE_SPECULARMAP":"",h.specularColorMap?"#define USE_SPECULAR_COLORMAP":"",h.specularIntensityMap?"#define USE_SPECULAR_INTENSITYMAP":"",h.roughnessMap?"#define USE_ROUGHNESSMAP":"",h.metalnessMap?"#define USE_METALNESSMAP":"",h.alphaMap?"#define USE_ALPHAMAP":"",h.alphaHash?"#define USE_ALPHAHASH":"",h.transmission?"#define USE_TRANSMISSION":"",h.transmissionMap?"#define USE_TRANSMISSIONMAP":"",h.thicknessMap?"#define USE_THICKNESSMAP":"",h.sheenColorMap?"#define USE_SHEEN_COLORMAP":"",h.sheenRoughnessMap?"#define USE_SHEEN_ROUGHNESSMAP":"",h.mapUv?"#define MAP_UV "+h.mapUv:"",h.alphaMapUv?"#define ALPHAMAP_UV "+h.alphaMapUv:"",h.lightMapUv?"#define LIGHTMAP_UV "+h.lightMapUv:"",h.aoMapUv?"#define AOMAP_UV "+h.aoMapUv:"",h.emissiveMapUv?"#define EMISSIVEMAP_UV "+h.emissiveMapUv:"",h.bumpMapUv?"#define BUMPMAP_UV "+h.bumpMapUv:"",h.normalMapUv?"#define NORMALMAP_UV "+h.normalMapUv:"",h.displacementMapUv?"#define DISPLACEMENTMAP_UV "+h.displacementMapUv:"",h.metalnessMapUv?"#define METALNESSMAP_UV "+h.metalnessMapUv:"",h.roughnessMapUv?"#define ROUGHNESSMAP_UV "+h.roughnessMapUv:"",h.anisotropyMapUv?"#define ANISOTROPYMAP_UV "+h.anisotropyMapUv:"",h.clearcoatMapUv?"#define CLEARCOATMAP_UV "+h.clearcoatMapUv:"",h.clearcoatNormalMapUv?"#define CLEARCOAT_NORMALMAP_UV "+h.clearcoatNormalMapUv:"",h.clearcoatRoughnessMapUv?"#define CLEARCOAT_ROUGHNESSMAP_UV "+h.clearcoatRoughnessMapUv:"",h.iridescenceMapUv?"#define IRIDESCENCEMAP_UV "+h.iridescenceMapUv:"",h.iridescenceThicknessMapUv?"#define IRIDESCENCE_THICKNESSMAP_UV "+h.iridescenceThicknessMapUv:"",h.sheenColorMapUv?"#define SHEEN_COLORMAP_UV "+h.sheenColorMapUv:"",h.sheenRoughnessMapUv?"#define SHEEN_ROUGHNESSMAP_UV "+h.sheenRoughnessMapUv:"",h.specularMapUv?"#define SPECULARMAP_UV "+h.specularMapUv:"",h.specularColorMapUv?"#define SPECULAR_COLORMAP_UV "+h.specularColorMapUv:"",h.specularIntensityMapUv?"#define SPECULAR_INTENSITYMAP_UV "+h.specularIntensityMapUv:"",h.transmissionMapUv?"#define TRANSMISSIONMAP_UV "+h.transmissionMapUv:"",h.thicknessMapUv?"#define THICKNESSMAP_UV "+h.thicknessMapUv:"",h.vertexTangents&&h.flatShading===!1?"#define USE_TANGENT":"",h.vertexColors?"#define USE_COLOR":"",h.vertexAlphas?"#define USE_COLOR_ALPHA":"",h.vertexUv1s?"#define USE_UV1":"",h.vertexUv2s?"#define USE_UV2":"",h.vertexUv3s?"#define USE_UV3":"",h.pointsUvs?"#define USE_POINTS_UV":"",h.flatShading?"#define FLAT_SHADED":"",h.skinning?"#define USE_SKINNING":"",h.morphTargets?"#define USE_MORPHTARGETS":"",h.morphNormals&&h.flatShading===!1?"#define USE_MORPHNORMALS":"",h.morphColors?"#define USE_MORPHCOLORS":"",h.morphTargetsCount>0?"#define MORPHTARGETS_TEXTURE_STRIDE "+h.morphTextureStride:"",h.morphTargetsCount>0?"#define MORPHTARGETS_COUNT "+h.morphTargetsCount:"",h.doubleSided?"#define DOUBLE_SIDED":"",h.flipSided?"#define FLIP_SIDED":"",h.shadowMapEnabled?"#define USE_SHADOWMAP":"",h.shadowMapEnabled?"#define "+j:"",h.sizeAttenuation?"#define USE_SIZEATTENUATION":"",h.numLightProbes>0?"#define USE_LIGHT_PROBES":"",h.logarithmicDepthBuffer?"#define USE_LOGARITHMIC_DEPTH_BUFFER":"",h.reversedDepthBuffer?"#define USE_REVERSED_DEPTH_BUFFER":"","uniform mat4 modelMatrix;","uniform mat4 modelViewMatrix;","uniform mat4 projectionMatrix;","uniform mat4 viewMatrix;","uniform mat3 normalMatrix;","uniform vec3 cameraPosition;","uniform bool isOrthographic;","#ifdef USE_INSTANCING","	attribute mat4 instanceMatrix;","#endif","#ifdef USE_INSTANCING_COLOR","	attribute vec3 instanceColor;","#endif","#ifdef USE_INSTANCING_MORPH","	uniform sampler2D morphTexture;","#endif","attribute vec3 position;","attribute vec3 normal;","attribute vec2 uv;","#ifdef USE_UV1","	attribute vec2 uv1;","#endif","#ifdef USE_UV2","	attribute vec2 uv2;","#endif","#ifdef USE_UV3","	attribute vec2 uv3;","#endif","#ifdef USE_TANGENT","	attribute vec4 tangent;","#endif","#if defined( USE_COLOR_ALPHA )","	attribute vec4 color;","#elif defined( USE_COLOR )","	attribute vec3 color;","#endif","#ifdef USE_SKINNING","	attribute vec4 skinIndex;","	attribute vec4 skinWeight;","#endif",`
`].filter(xi).join(`
`),b=[oi(h),"#define SHADER_TYPE "+h.shaderType,"#define SHADER_NAME "+h.shaderName,X,h.useFog&&h.fog?"#define USE_FOG":"",h.useFog&&h.fogExp2?"#define FOG_EXP2":"",h.alphaToCoverage?"#define ALPHA_TO_COVERAGE":"",h.map?"#define USE_MAP":"",h.matcap?"#define USE_MATCAP":"",h.envMap?"#define USE_ENVMAP":"",h.envMap?"#define "+$:"",h.envMap?"#define "+ht:"",h.envMap?"#define "+K:"",F?"#define CUBEUV_TEXEL_WIDTH "+F.texelWidth:"",F?"#define CUBEUV_TEXEL_HEIGHT "+F.texelHeight:"",F?"#define CUBEUV_MAX_MIP "+F.maxMip+".0":"",h.lightMap?"#define USE_LIGHTMAP":"",h.aoMap?"#define USE_AOMAP":"",h.bumpMap?"#define USE_BUMPMAP":"",h.normalMap?"#define USE_NORMALMAP":"",h.normalMapObjectSpace?"#define USE_NORMALMAP_OBJECTSPACE":"",h.normalMapTangentSpace?"#define USE_NORMALMAP_TANGENTSPACE":"",h.emissiveMap?"#define USE_EMISSIVEMAP":"",h.anisotropy?"#define USE_ANISOTROPY":"",h.anisotropyMap?"#define USE_ANISOTROPYMAP":"",h.clearcoat?"#define USE_CLEARCOAT":"",h.clearcoatMap?"#define USE_CLEARCOATMAP":"",h.clearcoatRoughnessMap?"#define USE_CLEARCOAT_ROUGHNESSMAP":"",h.clearcoatNormalMap?"#define USE_CLEARCOAT_NORMALMAP":"",h.dispersion?"#define USE_DISPERSION":"",h.iridescence?"#define USE_IRIDESCENCE":"",h.iridescenceMap?"#define USE_IRIDESCENCEMAP":"",h.iridescenceThicknessMap?"#define USE_IRIDESCENCE_THICKNESSMAP":"",h.specularMap?"#define USE_SPECULARMAP":"",h.specularColorMap?"#define USE_SPECULAR_COLORMAP":"",h.specularIntensityMap?"#define USE_SPECULAR_INTENSITYMAP":"",h.roughnessMap?"#define USE_ROUGHNESSMAP":"",h.metalnessMap?"#define USE_METALNESSMAP":"",h.alphaMap?"#define USE_ALPHAMAP":"",h.alphaTest?"#define USE_ALPHATEST":"",h.alphaHash?"#define USE_ALPHAHASH":"",h.sheen?"#define USE_SHEEN":"",h.sheenColorMap?"#define USE_SHEEN_COLORMAP":"",h.sheenRoughnessMap?"#define USE_SHEEN_ROUGHNESSMAP":"",h.transmission?"#define USE_TRANSMISSION":"",h.transmissionMap?"#define USE_TRANSMISSIONMAP":"",h.thicknessMap?"#define USE_THICKNESSMAP":"",h.vertexTangents&&h.flatShading===!1?"#define USE_TANGENT":"",h.vertexColors||h.instancingColor?"#define USE_COLOR":"",h.vertexAlphas||h.batchingColor?"#define USE_COLOR_ALPHA":"",h.vertexUv1s?"#define USE_UV1":"",h.vertexUv2s?"#define USE_UV2":"",h.vertexUv3s?"#define USE_UV3":"",h.pointsUvs?"#define USE_POINTS_UV":"",h.gradientMap?"#define USE_GRADIENTMAP":"",h.flatShading?"#define FLAT_SHADED":"",h.doubleSided?"#define DOUBLE_SIDED":"",h.flipSided?"#define FLIP_SIDED":"",h.shadowMapEnabled?"#define USE_SHADOWMAP":"",h.shadowMapEnabled?"#define "+j:"",h.premultipliedAlpha?"#define PREMULTIPLIED_ALPHA":"",h.numLightProbes>0?"#define USE_LIGHT_PROBES":"",h.decodeVideoTexture?"#define DECODE_VIDEO_TEXTURE":"",h.decodeVideoTextureEmissive?"#define DECODE_VIDEO_TEXTURE_EMISSIVE":"",h.logarithmicDepthBuffer?"#define USE_LOGARITHMIC_DEPTH_BUFFER":"",h.reversedDepthBuffer?"#define USE_REVERSED_DEPTH_BUFFER":"","uniform mat4 viewMatrix;","uniform vec3 cameraPosition;","uniform bool isOrthographic;",h.toneMapping!==f.y_p?"#define TONE_MAPPING":"",h.toneMapping!==f.y_p?ve.tonemapping_pars_fragment:"",h.toneMapping!==f.y_p?At("toneMapping",h.toneMapping):"",h.dithering?"#define DITHERING":"",h.opaque?"#define OPAQUE":"",ve.colorspace_pars_fragment,ye("linearToOutputTexel",h.outputColorSpace),O(),h.useDepthPacking?"#define DEPTH_PACKING "+h.depthPacking:"",`
`].filter(xi).join(`
`)),I=_n(I),I=Mr(I,h),I=Sr(I,h),V=_n(V),V=Mr(V,h),V=Sr(V,h),I=ka(I),V=ka(V),h.isRawShaderMaterial!==!0&&(z=`#version 300 es
`,A=[G,"#define attribute in","#define varying out","#define texture2D texture"].join(`
`)+`
`+A,b=["#define varying in",h.glslVersion===f.Wdf?"":"layout(location = 0) out highp vec4 pc_fragColor;",h.glslVersion===f.Wdf?"":"#define gl_FragColor pc_fragColor","#define gl_FragDepthEXT gl_FragDepth","#define texture2D texture","#define textureCube texture","#define texture2DProj textureProj","#define texture2DLodEXT textureLod","#define texture2DProjLodEXT textureProjLod","#define textureCubeLodEXT textureLod","#define texture2DGradEXT textureGrad","#define texture2DProjGradEXT textureProjGrad","#define textureCubeGradEXT textureGrad"].join(`
`)+`
`+b);const q=z+A+I,H=z+b+V,ot=Ba(E,E.VERTEX_SHADER,q),J=Ba(E,E.FRAGMENT_SHADER,H);E.attachShader(nt,ot),E.attachShader(nt,J),h.index0AttributeName!==void 0?E.bindAttribLocation(nt,0,h.index0AttributeName):h.morphTargets===!0&&E.bindAttribLocation(nt,0,"position"),E.linkProgram(nt);function st(Z){if(o.debug.checkShaderErrors){const dt=E.getProgramInfoLog(nt)||"",mt=E.getShaderInfoLog(ot)||"",gt=E.getShaderInfoLog(J)||"",_t=dt.trim(),ct=mt.trim(),it=gt.trim();let It=!0,zt=!0;if(E.getProgramParameter(nt,E.LINK_STATUS)===!1)if(It=!1,typeof o.debug.onShaderError=="function")o.debug.onShaderError(E,nt,ot,J);else{const kt=gn(E,ot,"vertex"),ue=gn(E,J,"fragment");(0,f.z3S)("THREE.WebGLProgram: Shader Error "+E.getError()+" - VALIDATE_STATUS "+E.getProgramParameter(nt,E.VALIDATE_STATUS)+`

Material Name: `+Z.name+`
Material Type: `+Z.type+`

Program Info Log: `+_t+`
`+kt+`
`+ue)}else _t!==""?(0,f.R8M)("WebGLProgram: Program Info Log:",_t):(ct===""||it==="")&&(zt=!1);zt&&(Z.diagnostics={runnable:It,programLog:_t,vertexShader:{log:ct,prefix:A},fragmentShader:{log:it,prefix:b}})}E.deleteShader(ot),E.deleteShader(J),R=new Us(E,nt),U=Fn(E,nt)}let R;this.getUniforms=function(){return R===void 0&&st(this),R};let U;this.getAttributes=function(){return U===void 0&&st(this),U};let Tt=h.rendererExtensionParallelShaderCompile===!1;return this.isReady=function(){return Tt===!1&&(Tt=E.getProgramParameter(nt,Mc)),Tt},this.destroy=function(){x.releaseStatesOfProgram(this),E.deleteProgram(nt),this.program=void 0},this.type=h.shaderType,this.name=h.shaderName,this.id=Sc++,this.cacheKey=p,this.usedTimes=1,this.program=nt,this.vertexShader=ot,this.fragmentShader=J,this}let lu=0;class Me{constructor(){this.shaderCache=new Map,this.materialCache=new Map}update(p){const h=p.vertexShader,x=p.fragmentShader,E=this._getShaderStage(h),S=this._getShaderStage(x),I=this._getShaderCacheForMaterial(p);return I.has(E)===!1&&(I.add(E),E.usedTimes++),I.has(S)===!1&&(I.add(S),S.usedTimes++),this}remove(p){const h=this.materialCache.get(p);for(const x of h)x.usedTimes--,x.usedTimes===0&&this.shaderCache.delete(x.code);return this.materialCache.delete(p),this}getVertexShaderID(p){return this._getShaderStage(p.vertexShader).id}getFragmentShaderID(p){return this._getShaderStage(p.fragmentShader).id}dispose(){this.shaderCache.clear(),this.materialCache.clear()}_getShaderCacheForMaterial(p){const h=this.materialCache;let x=h.get(p);return x===void 0&&(x=new Set,h.set(p,x)),x}_getShaderStage(p){const h=this.shaderCache;let x=h.get(p);return x===void 0&&(x=new Fi(p),h.set(p,x)),x}}class Fi{constructor(p){this.id=lu++,this.code=p,this.usedTimes=0}}function On(o,p,h,x,E,S){const I=new f.zgK,V=new Me,j=new Set,$=[],ht=new Map,K=x.logarithmicDepthBuffer;let F=x.precision;const G={MeshDepthMaterial:"depth",MeshDistanceMaterial:"distance",MeshNormalMaterial:"normal",MeshBasicMaterial:"basic",MeshLambertMaterial:"lambert",MeshPhongMaterial:"phong",MeshToonMaterial:"toon",MeshStandardMaterial:"physical",MeshPhysicalMaterial:"physical",MeshMatcapMaterial:"matcap",LineBasicMaterial:"basic",LineDashedMaterial:"dashed",PointsMaterial:"points",ShadowMaterial:"shadow",SpriteMaterial:"sprite"};function X(R){return j.add(R),R===0?"uv":`uv${R}`}function nt(R,U,Tt,Z,dt){const mt=Z.fog,gt=dt.geometry,_t=R.isMeshStandardMaterial||R.isMeshLambertMaterial||R.isMeshPhongMaterial?Z.environment:null,ct=R.isMeshStandardMaterial||R.isMeshLambertMaterial&&!R.envMap||R.isMeshPhongMaterial&&!R.envMap,it=p.get(R.envMap||_t,ct),It=it&&it.mapping===f.Om?it.image.height:null,zt=G[R.type];R.precision!==null&&(F=x.getMaxPrecision(R.precision),F!==R.precision&&(0,f.R8M)("WebGLProgram.getParameters:",R.precision,"not supported, using",F,"instead."));const kt=gt.morphAttributes.position||gt.morphAttributes.normal||gt.morphAttributes.color,ue=kt!==void 0?kt.length:0;let te=0;gt.morphAttributes.position!==void 0&&(te=1),gt.morphAttributes.normal!==void 0&&(te=2),gt.morphAttributes.color!==void 0&&(te=3);let se,Ze,$e,xt;if(zt){const ze=Dn[zt];se=ze.vertexShader,Ze=ze.fragmentShader}else se=R.vertexShader,Ze=R.fragmentShader,V.update(R),$e=V.getVertexShaderID(R),xt=V.getFragmentShaderID(R);const Lt=o.getRenderTarget(),Dt=o.state.buffers.depth.getReversed(),Te=dt.isInstancedMesh===!0,de=dt.isBatchedMesh===!0,_e=!!R.map,Ut=!!R.matcap,Ie=!!it,Pe=!!R.aoMap,Be=!!R.lightMap,ae=!!R.bumpMap,Je=!!R.normalMap,B=!!R.displacementMap,Xe=!!R.emissiveMap,Ce=!!R.metalnessMap,Ee=!!R.roughnessMap,Jt=R.anisotropy>0,P=R.clearcoat>0,y=R.dispersion>0,W=R.iridescence>0,ut=R.sheen>0,yt=R.transmission>0,pt=Jt&&!!R.anisotropyMap,Ft=P&&!!R.clearcoatMap,Pt=P&&!!R.clearcoatNormalMap,ne=P&&!!R.clearcoatRoughnessMap,ce=W&&!!R.iridescenceMap,bt=W&&!!R.iridescenceThicknessMap,wt=ut&&!!R.sheenColorMap,qt=ut&&!!R.sheenRoughnessMap,Zt=!!R.specularMap,Ht=!!R.specularColorMap,Se=!!R.specularIntensityMap,k=yt&&!!R.transmissionMap,Rt=yt&&!!R.thicknessMap,Ct=!!R.gradientMap,Xt=!!R.alphaMap,Et=R.alphaTest>0,ft=!!R.alphaHash,$t=!!R.extensions;let me=f.y_p;R.toneMapped&&(Lt===null||Lt.isXRRenderTarget===!0)&&(me=o.toneMapping);const Ge={shaderID:zt,shaderType:R.type,shaderName:R.name,vertexShader:se,fragmentShader:Ze,defines:R.defines,customVertexShaderID:$e,customFragmentShaderID:xt,isRawShaderMaterial:R.isRawShaderMaterial===!0,glslVersion:R.glslVersion,precision:F,batching:de,batchingColor:de&&dt._colorsTexture!==null,instancing:Te,instancingColor:Te&&dt.instanceColor!==null,instancingMorph:Te&&dt.morphTexture!==null,outputColorSpace:Lt===null?o.outputColorSpace:Lt.isXRRenderTarget===!0?Lt.texture.colorSpace:f.Zr2,alphaToCoverage:!!R.alphaToCoverage,map:_e,matcap:Ut,envMap:Ie,envMapMode:Ie&&it.mapping,envMapCubeUVHeight:It,aoMap:Pe,lightMap:Be,bumpMap:ae,normalMap:Je,displacementMap:B,emissiveMap:Xe,normalMapObjectSpace:Je&&R.normalMapType===f.vyJ,normalMapTangentSpace:Je&&R.normalMapType===f.bI3,metalnessMap:Ce,roughnessMap:Ee,anisotropy:Jt,anisotropyMap:pt,clearcoat:P,clearcoatMap:Ft,clearcoatNormalMap:Pt,clearcoatRoughnessMap:ne,dispersion:y,iridescence:W,iridescenceMap:ce,iridescenceThicknessMap:bt,sheen:ut,sheenColorMap:wt,sheenRoughnessMap:qt,specularMap:Zt,specularColorMap:Ht,specularIntensityMap:Se,transmission:yt,transmissionMap:k,thicknessMap:Rt,gradientMap:Ct,opaque:R.transparent===!1&&R.blending===f.NTi&&R.alphaToCoverage===!1,alphaMap:Xt,alphaTest:Et,alphaHash:ft,combine:R.combine,mapUv:_e&&X(R.map.channel),aoMapUv:Pe&&X(R.aoMap.channel),lightMapUv:Be&&X(R.lightMap.channel),bumpMapUv:ae&&X(R.bumpMap.channel),normalMapUv:Je&&X(R.normalMap.channel),displacementMapUv:B&&X(R.displacementMap.channel),emissiveMapUv:Xe&&X(R.emissiveMap.channel),metalnessMapUv:Ce&&X(R.metalnessMap.channel),roughnessMapUv:Ee&&X(R.roughnessMap.channel),anisotropyMapUv:pt&&X(R.anisotropyMap.channel),clearcoatMapUv:Ft&&X(R.clearcoatMap.channel),clearcoatNormalMapUv:Pt&&X(R.clearcoatNormalMap.channel),clearcoatRoughnessMapUv:ne&&X(R.clearcoatRoughnessMap.channel),iridescenceMapUv:ce&&X(R.iridescenceMap.channel),iridescenceThicknessMapUv:bt&&X(R.iridescenceThicknessMap.channel),sheenColorMapUv:wt&&X(R.sheenColorMap.channel),sheenRoughnessMapUv:qt&&X(R.sheenRoughnessMap.channel),specularMapUv:Zt&&X(R.specularMap.channel),specularColorMapUv:Ht&&X(R.specularColorMap.channel),specularIntensityMapUv:Se&&X(R.specularIntensityMap.channel),transmissionMapUv:k&&X(R.transmissionMap.channel),thicknessMapUv:Rt&&X(R.thicknessMap.channel),alphaMapUv:Xt&&X(R.alphaMap.channel),vertexTangents:!!gt.attributes.tangent&&(Je||Jt),vertexColors:R.vertexColors,vertexAlphas:R.vertexColors===!0&&!!gt.attributes.color&&gt.attributes.color.itemSize===4,pointsUvs:dt.isPoints===!0&&!!gt.attributes.uv&&(_e||Xt),fog:!!mt,useFog:R.fog===!0,fogExp2:!!mt&&mt.isFogExp2,flatShading:R.wireframe===!1&&(R.flatShading===!0||gt.attributes.normal===void 0&&Je===!1&&(R.isMeshLambertMaterial||R.isMeshPhongMaterial||R.isMeshStandardMaterial||R.isMeshPhysicalMaterial)),sizeAttenuation:R.sizeAttenuation===!0,logarithmicDepthBuffer:K,reversedDepthBuffer:Dt,skinning:dt.isSkinnedMesh===!0,morphTargets:gt.morphAttributes.position!==void 0,morphNormals:gt.morphAttributes.normal!==void 0,morphColors:gt.morphAttributes.color!==void 0,morphTargetsCount:ue,morphTextureStride:te,numDirLights:U.directional.length,numPointLights:U.point.length,numSpotLights:U.spot.length,numSpotLightMaps:U.spotLightMap.length,numRectAreaLights:U.rectArea.length,numHemiLights:U.hemi.length,numDirLightShadows:U.directionalShadowMap.length,numPointLightShadows:U.pointShadowMap.length,numSpotLightShadows:U.spotShadowMap.length,numSpotLightShadowsWithMaps:U.numSpotLightShadowsWithMaps,numLightProbes:U.numLightProbes,numClippingPlanes:S.numPlanes,numClipIntersection:S.numIntersection,dithering:R.dithering,shadowMapEnabled:o.shadowMap.enabled&&Tt.length>0,shadowMapType:o.shadowMap.type,toneMapping:me,decodeVideoTexture:_e&&R.map.isVideoTexture===!0&&f.ppV.getTransfer(R.map.colorSpace)===f.KLL,decodeVideoTextureEmissive:Xe&&R.emissiveMap.isVideoTexture===!0&&f.ppV.getTransfer(R.emissiveMap.colorSpace)===f.KLL,premultipliedAlpha:R.premultipliedAlpha,doubleSided:R.side===f.$EB,flipSided:R.side===f.hsX,useDepthPacking:R.depthPacking>=0,depthPacking:R.depthPacking||0,index0AttributeName:R.index0AttributeName,extensionClipCullDistance:$t&&R.extensions.clipCullDistance===!0&&h.has("WEBGL_clip_cull_distance"),extensionMultiDraw:($t&&R.extensions.multiDraw===!0||de)&&h.has("WEBGL_multi_draw"),rendererExtensionParallelShaderCompile:h.has("KHR_parallel_shader_compile"),customProgramCacheKey:R.customProgramCacheKey()};return Ge.vertexUv1s=j.has(1),Ge.vertexUv2s=j.has(2),Ge.vertexUv3s=j.has(3),j.clear(),Ge}function A(R){const U=[];if(R.shaderID?U.push(R.shaderID):(U.push(R.customVertexShaderID),U.push(R.customFragmentShaderID)),R.defines!==void 0)for(const Tt in R.defines)U.push(Tt),U.push(R.defines[Tt]);return R.isRawShaderMaterial===!1&&(b(U,R),z(U,R),U.push(o.outputColorSpace)),U.push(R.customProgramCacheKey),U.join()}function b(R,U){R.push(U.precision),R.push(U.outputColorSpace),R.push(U.envMapMode),R.push(U.envMapCubeUVHeight),R.push(U.mapUv),R.push(U.alphaMapUv),R.push(U.lightMapUv),R.push(U.aoMapUv),R.push(U.bumpMapUv),R.push(U.normalMapUv),R.push(U.displacementMapUv),R.push(U.emissiveMapUv),R.push(U.metalnessMapUv),R.push(U.roughnessMapUv),R.push(U.anisotropyMapUv),R.push(U.clearcoatMapUv),R.push(U.clearcoatNormalMapUv),R.push(U.clearcoatRoughnessMapUv),R.push(U.iridescenceMapUv),R.push(U.iridescenceThicknessMapUv),R.push(U.sheenColorMapUv),R.push(U.sheenRoughnessMapUv),R.push(U.specularMapUv),R.push(U.specularColorMapUv),R.push(U.specularIntensityMapUv),R.push(U.transmissionMapUv),R.push(U.thicknessMapUv),R.push(U.combine),R.push(U.fogExp2),R.push(U.sizeAttenuation),R.push(U.morphTargetsCount),R.push(U.morphAttributeCount),R.push(U.numDirLights),R.push(U.numPointLights),R.push(U.numSpotLights),R.push(U.numSpotLightMaps),R.push(U.numHemiLights),R.push(U.numRectAreaLights),R.push(U.numDirLightShadows),R.push(U.numPointLightShadows),R.push(U.numSpotLightShadows),R.push(U.numSpotLightShadowsWithMaps),R.push(U.numLightProbes),R.push(U.shadowMapType),R.push(U.toneMapping),R.push(U.numClippingPlanes),R.push(U.numClipIntersection),R.push(U.depthPacking)}function z(R,U){I.disableAll(),U.instancing&&I.enable(0),U.instancingColor&&I.enable(1),U.instancingMorph&&I.enable(2),U.matcap&&I.enable(3),U.envMap&&I.enable(4),U.normalMapObjectSpace&&I.enable(5),U.normalMapTangentSpace&&I.enable(6),U.clearcoat&&I.enable(7),U.iridescence&&I.enable(8),U.alphaTest&&I.enable(9),U.vertexColors&&I.enable(10),U.vertexAlphas&&I.enable(11),U.vertexUv1s&&I.enable(12),U.vertexUv2s&&I.enable(13),U.vertexUv3s&&I.enable(14),U.vertexTangents&&I.enable(15),U.anisotropy&&I.enable(16),U.alphaHash&&I.enable(17),U.batching&&I.enable(18),U.dispersion&&I.enable(19),U.batchingColor&&I.enable(20),U.gradientMap&&I.enable(21),R.push(I.mask),I.disableAll(),U.fog&&I.enable(0),U.useFog&&I.enable(1),U.flatShading&&I.enable(2),U.logarithmicDepthBuffer&&I.enable(3),U.reversedDepthBuffer&&I.enable(4),U.skinning&&I.enable(5),U.morphTargets&&I.enable(6),U.morphNormals&&I.enable(7),U.morphColors&&I.enable(8),U.premultipliedAlpha&&I.enable(9),U.shadowMapEnabled&&I.enable(10),U.doubleSided&&I.enable(11),U.flipSided&&I.enable(12),U.useDepthPacking&&I.enable(13),U.dithering&&I.enable(14),U.transmission&&I.enable(15),U.sheen&&I.enable(16),U.opaque&&I.enable(17),U.pointsUvs&&I.enable(18),U.decodeVideoTexture&&I.enable(19),U.decodeVideoTextureEmissive&&I.enable(20),U.alphaToCoverage&&I.enable(21),R.push(I.mask)}function q(R){const U=G[R.type];let Tt;if(U){const Z=Dn[U];Tt=f.LlO.clone(Z.uniforms)}else Tt=R.uniforms;return Tt}function H(R,U){let Tt=ht.get(U);return Tt!==void 0?++Tt.usedTimes:(Tt=new Er(o,U,R,E),$.push(Tt),ht.set(U,Tt)),Tt}function ot(R){if(--R.usedTimes===0){const U=$.indexOf(R);$[U]=$[$.length-1],$.pop(),ht.delete(R.cacheKey),R.destroy()}}function J(R){V.remove(R)}function st(){V.dispose()}return{getParameters:nt,getProgramCacheKey:A,getUniforms:q,acquireProgram:H,releaseProgram:ot,releaseShaderCache:J,programs:$,dispose:st}}function Cc(){let o=new WeakMap;function p(I){return o.has(I)}function h(I){let V=o.get(I);return V===void 0&&(V={},o.set(I,V)),V}function x(I){o.delete(I)}function E(I,V,j){o.get(I)[V]=j}function S(){o=new WeakMap}return{has:p,get:h,remove:x,update:E,dispose:S}}function Rc(o,p){return o.groupOrder!==p.groupOrder?o.groupOrder-p.groupOrder:o.renderOrder!==p.renderOrder?o.renderOrder-p.renderOrder:o.material.id!==p.material.id?o.material.id-p.material.id:o.materialVariant!==p.materialVariant?o.materialVariant-p.materialVariant:o.z!==p.z?o.z-p.z:o.id-p.id}function jn(o,p){return o.groupOrder!==p.groupOrder?o.groupOrder-p.groupOrder:o.renderOrder!==p.renderOrder?o.renderOrder-p.renderOrder:o.z!==p.z?p.z-o.z:o.id-p.id}function is(){const o=[];let p=0;const h=[],x=[],E=[];function S(){p=0,h.length=0,x.length=0,E.length=0}function I(F){let G=0;return F.isInstancedMesh&&(G+=2),F.isSkinnedMesh&&(G+=1),G}function V(F,G,X,nt,A,b){let z=o[p];return z===void 0?(z={id:F.id,object:F,geometry:G,material:X,materialVariant:I(F),groupOrder:nt,renderOrder:F.renderOrder,z:A,group:b},o[p]=z):(z.id=F.id,z.object=F,z.geometry=G,z.material=X,z.materialVariant=I(F),z.groupOrder=nt,z.renderOrder=F.renderOrder,z.z=A,z.group=b),p++,z}function j(F,G,X,nt,A,b){const z=V(F,G,X,nt,A,b);X.transmission>0?x.push(z):X.transparent===!0?E.push(z):h.push(z)}function $(F,G,X,nt,A,b){const z=V(F,G,X,nt,A,b);X.transmission>0?x.unshift(z):X.transparent===!0?E.unshift(z):h.unshift(z)}function ht(F,G){h.length>1&&h.sort(F||Rc),x.length>1&&x.sort(G||jn),E.length>1&&E.sort(G||jn)}function K(){for(let F=p,G=o.length;F<G;F++){const X=o[F];if(X.id===null)break;X.id=null,X.object=null,X.geometry=null,X.material=null,X.group=null}}return{opaque:h,transmissive:x,transparent:E,init:S,push:j,unshift:$,finish:K,sort:ht}}function wn(){let o=new WeakMap;function p(x,E){const S=o.get(x);let I;return S===void 0?(I=new is,o.set(x,[I])):E>=S.length?(I=new is,S.push(I)):I=S[E],I}function h(){o=new WeakMap}return{get:p,dispose:h}}function Wa(){const o={};return{get:function(p){if(o[p.id]!==void 0)return o[p.id];let h;switch(p.type){case"DirectionalLight":h={direction:new f.Pq0,color:new f.Q1f};break;case"SpotLight":h={position:new f.Pq0,direction:new f.Pq0,color:new f.Q1f,distance:0,coneCos:0,penumbraCos:0,decay:0};break;case"PointLight":h={position:new f.Pq0,color:new f.Q1f,distance:0,decay:0};break;case"HemisphereLight":h={direction:new f.Pq0,skyColor:new f.Q1f,groundColor:new f.Q1f};break;case"RectAreaLight":h={color:new f.Q1f,position:new f.Pq0,halfWidth:new f.Pq0,halfHeight:new f.Pq0};break}return o[p.id]=h,h}}}function Xa(){const o={};return{get:function(p){if(o[p.id]!==void 0)return o[p.id];let h;switch(p.type){case"DirectionalLight":h={shadowIntensity:1,shadowBias:0,shadowNormalBias:0,shadowRadius:1,shadowMapSize:new f.I9Y};break;case"SpotLight":h={shadowIntensity:1,shadowBias:0,shadowNormalBias:0,shadowRadius:1,shadowMapSize:new f.I9Y};break;case"PointLight":h={shadowIntensity:1,shadowBias:0,shadowNormalBias:0,shadowRadius:1,shadowMapSize:new f.I9Y,shadowCameraNear:1,shadowCameraFar:1e3};break}return o[p.id]=h,h}}}let Bn=0;function wr(o,p){return(p.castShadow?2:0)-(o.castShadow?2:0)+(p.map?1:0)-(o.map?1:0)}function Ic(o){const p=new Wa,h=Xa(),x={version:0,hash:{directionalLength:-1,pointLength:-1,spotLength:-1,rectAreaLength:-1,hemiLength:-1,numDirectionalShadows:-1,numPointShadows:-1,numSpotShadows:-1,numSpotMaps:-1,numLightProbes:-1},ambient:[0,0,0],probe:[],directional:[],directionalShadow:[],directionalShadowMap:[],directionalShadowMatrix:[],spot:[],spotLightMap:[],spotShadow:[],spotShadowMap:[],spotLightMatrix:[],rectArea:[],rectAreaLTC1:null,rectAreaLTC2:null,point:[],pointShadow:[],pointShadowMap:[],pointShadowMatrix:[],hemi:[],numSpotLightShadowsWithMaps:0,numLightProbes:0};for(let $=0;$<9;$++)x.probe.push(new f.Pq0);const E=new f.Pq0,S=new f.kn4,I=new f.kn4;function V($){let ht=0,K=0,F=0;for(let U=0;U<9;U++)x.probe[U].set(0,0,0);let G=0,X=0,nt=0,A=0,b=0,z=0,q=0,H=0,ot=0,J=0,st=0;$.sort(wr);for(let U=0,Tt=$.length;U<Tt;U++){const Z=$[U],dt=Z.color,mt=Z.intensity,gt=Z.distance;let _t=null;if(Z.shadow&&Z.shadow.map&&(Z.shadow.map.texture.format===f.paN?_t=Z.shadow.map.texture:_t=Z.shadow.map.depthTexture||Z.shadow.map.texture),Z.isAmbientLight)ht+=dt.r*mt,K+=dt.g*mt,F+=dt.b*mt;else if(Z.isLightProbe){for(let ct=0;ct<9;ct++)x.probe[ct].addScaledVector(Z.sh.coefficients[ct],mt);st++}else if(Z.isDirectionalLight){const ct=p.get(Z);if(ct.color.copy(Z.color).multiplyScalar(Z.intensity),Z.castShadow){const it=Z.shadow,It=h.get(Z);It.shadowIntensity=it.intensity,It.shadowBias=it.bias,It.shadowNormalBias=it.normalBias,It.shadowRadius=it.radius,It.shadowMapSize=it.mapSize,x.directionalShadow[G]=It,x.directionalShadowMap[G]=_t,x.directionalShadowMatrix[G]=Z.shadow.matrix,z++}x.directional[G]=ct,G++}else if(Z.isSpotLight){const ct=p.get(Z);ct.position.setFromMatrixPosition(Z.matrixWorld),ct.color.copy(dt).multiplyScalar(mt),ct.distance=gt,ct.coneCos=Math.cos(Z.angle),ct.penumbraCos=Math.cos(Z.angle*(1-Z.penumbra)),ct.decay=Z.decay,x.spot[nt]=ct;const it=Z.shadow;if(Z.map&&(x.spotLightMap[ot]=Z.map,ot++,it.updateMatrices(Z),Z.castShadow&&J++),x.spotLightMatrix[nt]=it.matrix,Z.castShadow){const It=h.get(Z);It.shadowIntensity=it.intensity,It.shadowBias=it.bias,It.shadowNormalBias=it.normalBias,It.shadowRadius=it.radius,It.shadowMapSize=it.mapSize,x.spotShadow[nt]=It,x.spotShadowMap[nt]=_t,H++}nt++}else if(Z.isRectAreaLight){const ct=p.get(Z);ct.color.copy(dt).multiplyScalar(mt),ct.halfWidth.set(Z.width*.5,0,0),ct.halfHeight.set(0,Z.height*.5,0),x.rectArea[A]=ct,A++}else if(Z.isPointLight){const ct=p.get(Z);if(ct.color.copy(Z.color).multiplyScalar(Z.intensity),ct.distance=Z.distance,ct.decay=Z.decay,Z.castShadow){const it=Z.shadow,It=h.get(Z);It.shadowIntensity=it.intensity,It.shadowBias=it.bias,It.shadowNormalBias=it.normalBias,It.shadowRadius=it.radius,It.shadowMapSize=it.mapSize,It.shadowCameraNear=it.camera.near,It.shadowCameraFar=it.camera.far,x.pointShadow[X]=It,x.pointShadowMap[X]=_t,x.pointShadowMatrix[X]=Z.shadow.matrix,q++}x.point[X]=ct,X++}else if(Z.isHemisphereLight){const ct=p.get(Z);ct.skyColor.copy(Z.color).multiplyScalar(mt),ct.groundColor.copy(Z.groundColor).multiplyScalar(mt),x.hemi[b]=ct,b++}}A>0&&(o.has("OES_texture_float_linear")===!0?(x.rectAreaLTC1=Nt.LTC_FLOAT_1,x.rectAreaLTC2=Nt.LTC_FLOAT_2):(x.rectAreaLTC1=Nt.LTC_HALF_1,x.rectAreaLTC2=Nt.LTC_HALF_2)),x.ambient[0]=ht,x.ambient[1]=K,x.ambient[2]=F;const R=x.hash;(R.directionalLength!==G||R.pointLength!==X||R.spotLength!==nt||R.rectAreaLength!==A||R.hemiLength!==b||R.numDirectionalShadows!==z||R.numPointShadows!==q||R.numSpotShadows!==H||R.numSpotMaps!==ot||R.numLightProbes!==st)&&(x.directional.length=G,x.spot.length=nt,x.rectArea.length=A,x.point.length=X,x.hemi.length=b,x.directionalShadow.length=z,x.directionalShadowMap.length=z,x.pointShadow.length=q,x.pointShadowMap.length=q,x.spotShadow.length=H,x.spotShadowMap.length=H,x.directionalShadowMatrix.length=z,x.pointShadowMatrix.length=q,x.spotLightMatrix.length=H+ot-J,x.spotLightMap.length=ot,x.numSpotLightShadowsWithMaps=J,x.numLightProbes=st,R.directionalLength=G,R.pointLength=X,R.spotLength=nt,R.rectAreaLength=A,R.hemiLength=b,R.numDirectionalShadows=z,R.numPointShadows=q,R.numSpotShadows=H,R.numSpotMaps=ot,R.numLightProbes=st,x.version=Bn++)}function j($,ht){let K=0,F=0,G=0,X=0,nt=0;const A=ht.matrixWorldInverse;for(let b=0,z=$.length;b<z;b++){const q=$[b];if(q.isDirectionalLight){const H=x.directional[K];H.direction.setFromMatrixPosition(q.matrixWorld),E.setFromMatrixPosition(q.target.matrixWorld),H.direction.sub(E),H.direction.transformDirection(A),K++}else if(q.isSpotLight){const H=x.spot[G];H.position.setFromMatrixPosition(q.matrixWorld),H.position.applyMatrix4(A),H.direction.setFromMatrixPosition(q.matrixWorld),E.setFromMatrixPosition(q.target.matrixWorld),H.direction.sub(E),H.direction.transformDirection(A),G++}else if(q.isRectAreaLight){const H=x.rectArea[X];H.position.setFromMatrixPosition(q.matrixWorld),H.position.applyMatrix4(A),I.identity(),S.copy(q.matrixWorld),S.premultiply(A),I.extractRotation(S),H.halfWidth.set(q.width*.5,0,0),H.halfHeight.set(0,q.height*.5,0),H.halfWidth.applyMatrix4(I),H.halfHeight.applyMatrix4(I),X++}else if(q.isPointLight){const H=x.point[F];H.position.setFromMatrixPosition(q.matrixWorld),H.position.applyMatrix4(A),F++}else if(q.isHemisphereLight){const H=x.hemi[nt];H.direction.setFromMatrixPosition(q.matrixWorld),H.direction.transformDirection(A),nt++}}}return{setup:V,setupView:j,state:x}}function Cr(o){const p=new Ic(o),h=[],x=[];function E(ht){$.camera=ht,h.length=0,x.length=0}function S(ht){h.push(ht)}function I(ht){x.push(ht)}function V(){p.setup(h)}function j(ht){p.setupView(h,ht)}const $={lightsArray:h,shadowsArray:x,camera:null,lights:p,transmissionRenderTarget:{}};return{init:E,state:$,setupLights:V,setupLightsView:j,pushLight:S,pushShadow:I}}function Oi(o){let p=new WeakMap;function h(E,S=0){const I=p.get(E);let V;return I===void 0?(V=new Cr(o),p.set(E,[V])):S>=I.length?(V=new Cr(o),I.push(V)):V=I[S],V}function x(){p=new WeakMap}return{get:h,dispose:x}}const ti=`void main() {
	gl_Position = vec4( position, 1.0 );
}`,Ns=`uniform sampler2D shadow_pass;
uniform vec2 resolution;
uniform float radius;
void main() {
	const float samples = float( VSM_SAMPLES );
	float mean = 0.0;
	float squared_mean = 0.0;
	float uvStride = samples <= 1.0 ? 0.0 : 2.0 / ( samples - 1.0 );
	float uvStart = samples <= 1.0 ? 0.0 : - 1.0;
	for ( float i = 0.0; i < samples; i ++ ) {
		float uvOffset = uvStart + i * uvStride;
		#ifdef HORIZONTAL_PASS
			vec2 distribution = texture2D( shadow_pass, ( gl_FragCoord.xy + vec2( uvOffset, 0.0 ) * radius ) / resolution ).rg;
			mean += distribution.x;
			squared_mean += distribution.y * distribution.y + distribution.x * distribution.x;
		#else
			float depth = texture2D( shadow_pass, ( gl_FragCoord.xy + vec2( 0.0, uvOffset ) * radius ) / resolution ).r;
			mean += depth;
			squared_mean += depth * depth;
		#endif
	}
	mean = mean / samples;
	squared_mean = squared_mean / samples;
	float std_dev = sqrt( max( 0.0, squared_mean - mean * mean ) );
	gl_FragColor = vec4( mean, std_dev, 0.0, 1.0 );
}`,ss=[new f.Pq0(1,0,0),new f.Pq0(-1,0,0),new f.Pq0(0,1,0),new f.Pq0(0,-1,0),new f.Pq0(0,0,1),new f.Pq0(0,0,-1)],Pc=[new f.Pq0(0,-1,0),new f.Pq0(0,-1,0),new f.Pq0(0,0,1),new f.Pq0(0,0,-1),new f.Pq0(0,-1,0),new f.Pq0(0,-1,0)],qa=new f.kn4,Bi=new f.Pq0,Fs=new f.Pq0;function Ya(o,p,h){let x=new f.PPD;const E=new f.I9Y,S=new f.I9Y,I=new f.IUQ,V=new f.CSG,j=new f.aVO,$={},ht=h.maxTextureSize,K={[f.hB5]:f.hsX,[f.hsX]:f.hB5,[f.$EB]:f.$EB},F=new f.BKk({defines:{VSM_SAMPLES:8},uniforms:{shadow_pass:{value:null},resolution:{value:new f.I9Y},radius:{value:4}},vertexShader:ti,fragmentShader:Ns}),G=F.clone();G.defines.HORIZONTAL_PASS=1;const X=new f.LoY;X.setAttribute("position",new f.THS(new Float32Array([-1,-1,.5,3,-1,.5,-1,3,.5]),3));const nt=new f.eaF(X,F),A=this;this.enabled=!1,this.autoUpdate=!0,this.needsUpdate=!1,this.type=f.QP0;let b=this.type;this.render=function(J,st,R){if(A.enabled===!1||A.autoUpdate===!1&&A.needsUpdate===!1||J.length===0)return;this.type===f.Wk7&&((0,f.R8M)("WebGLShadowMap: PCFSoftShadowMap has been deprecated. Using PCFShadowMap instead."),this.type=f.QP0);const U=o.getRenderTarget(),Tt=o.getActiveCubeFace(),Z=o.getActiveMipmapLevel(),dt=o.state;dt.setBlending(f.XIg),dt.buffers.depth.getReversed()===!0?dt.buffers.color.setClear(0,0,0,0):dt.buffers.color.setClear(1,1,1,1),dt.buffers.depth.setTest(!0),dt.setScissorTest(!1);const mt=b!==this.type;mt&&st.traverse(function(gt){gt.material&&(Array.isArray(gt.material)?gt.material.forEach(_t=>_t.needsUpdate=!0):gt.material.needsUpdate=!0)});for(let gt=0,_t=J.length;gt<_t;gt++){const ct=J[gt],it=ct.shadow;if(it===void 0){(0,f.R8M)("WebGLShadowMap:",ct,"has no shadow.");continue}if(it.autoUpdate===!1&&it.needsUpdate===!1)continue;E.copy(it.mapSize);const It=it.getFrameExtents();E.multiply(It),S.copy(it.mapSize),(E.x>ht||E.y>ht)&&(E.x>ht&&(S.x=Math.floor(ht/It.x),E.x=S.x*It.x,it.mapSize.x=S.x),E.y>ht&&(S.y=Math.floor(ht/It.y),E.y=S.y*It.y,it.mapSize.y=S.y));const zt=o.state.buffers.depth.getReversed();if(it.camera._reversedDepth=zt,it.map===null||mt===!0){if(it.map!==null&&(it.map.depthTexture!==null&&(it.map.depthTexture.dispose(),it.map.depthTexture=null),it.map.dispose()),this.type===f.RyA){if(ct.isPointLight){(0,f.R8M)("WebGLShadowMap: VSM shadow maps are not supported for PointLights. Use PCF or BasicShadowMap instead.");continue}it.map=new f.nWS(E.x,E.y,{format:f.paN,type:f.ix0,minFilter:f.k6q,magFilter:f.k6q,generateMipmaps:!1}),it.map.texture.name=ct.name+".shadowMap",it.map.depthTexture=new f.VCu(E.x,E.y,f.RQf),it.map.depthTexture.name=ct.name+".shadowMapDepth",it.map.depthTexture.format=f.zdS,it.map.depthTexture.compareFunction=null,it.map.depthTexture.minFilter=f.hxR,it.map.depthTexture.magFilter=f.hxR}else ct.isPointLight?(it.map=new ac(E.x),it.map.depthTexture=new f.Gc6(E.x,f.bkx)):(it.map=new f.nWS(E.x,E.y),it.map.depthTexture=new f.VCu(E.x,E.y,f.bkx)),it.map.depthTexture.name=ct.name+".shadowMap",it.map.depthTexture.format=f.zdS,this.type===f.QP0?(it.map.depthTexture.compareFunction=zt?f.gWB:f.TiK,it.map.depthTexture.minFilter=f.k6q,it.map.depthTexture.magFilter=f.k6q):(it.map.depthTexture.compareFunction=null,it.map.depthTexture.minFilter=f.hxR,it.map.depthTexture.magFilter=f.hxR);it.camera.updateProjectionMatrix()}const kt=it.map.isWebGLCubeRenderTarget?6:1;for(let ue=0;ue<kt;ue++){if(it.map.isWebGLCubeRenderTarget)o.setRenderTarget(it.map,ue),o.clear();else{ue===0&&(o.setRenderTarget(it.map),o.clear());const te=it.getViewport(ue);I.set(S.x*te.x,S.y*te.y,S.x*te.z,S.y*te.w),dt.viewport(I)}if(ct.isPointLight){const te=it.camera,se=it.matrix,Ze=ct.distance||te.far;Ze!==te.far&&(te.far=Ze,te.updateProjectionMatrix()),Bi.setFromMatrixPosition(ct.matrixWorld),te.position.copy(Bi),Fs.copy(te.position),Fs.add(ss[ue]),te.up.copy(Pc[ue]),te.lookAt(Fs),te.updateMatrixWorld(),se.makeTranslation(-Bi.x,-Bi.y,-Bi.z),qa.multiplyMatrices(te.projectionMatrix,te.matrixWorldInverse),it._frustum.setFromProjectionMatrix(qa,te.coordinateSystem,te.reversedDepth)}else it.updateMatrices(ct);x=it.getFrustum(),H(st,R,it.camera,ct,this.type)}it.isPointLightShadow!==!0&&this.type===f.RyA&&z(it,R),it.needsUpdate=!1}b=this.type,A.needsUpdate=!1,o.setRenderTarget(U,Tt,Z)};function z(J,st){const R=p.update(nt);F.defines.VSM_SAMPLES!==J.blurSamples&&(F.defines.VSM_SAMPLES=J.blurSamples,G.defines.VSM_SAMPLES=J.blurSamples,F.needsUpdate=!0,G.needsUpdate=!0),J.mapPass===null&&(J.mapPass=new f.nWS(E.x,E.y,{format:f.paN,type:f.ix0})),F.uniforms.shadow_pass.value=J.map.depthTexture,F.uniforms.resolution.value=J.mapSize,F.uniforms.radius.value=J.radius,o.setRenderTarget(J.mapPass),o.clear(),o.renderBufferDirect(st,null,R,F,nt,null),G.uniforms.shadow_pass.value=J.mapPass.texture,G.uniforms.resolution.value=J.mapSize,G.uniforms.radius.value=J.radius,o.setRenderTarget(J.map),o.clear(),o.renderBufferDirect(st,null,R,G,nt,null)}function q(J,st,R,U){let Tt=null;const Z=R.isPointLight===!0?J.customDistanceMaterial:J.customDepthMaterial;if(Z!==void 0)Tt=Z;else if(Tt=R.isPointLight===!0?j:V,o.localClippingEnabled&&st.clipShadows===!0&&Array.isArray(st.clippingPlanes)&&st.clippingPlanes.length!==0||st.displacementMap&&st.displacementScale!==0||st.alphaMap&&st.alphaTest>0||st.map&&st.alphaTest>0||st.alphaToCoverage===!0){const dt=Tt.uuid,mt=st.uuid;let gt=$[dt];gt===void 0&&(gt={},$[dt]=gt);let _t=gt[mt];_t===void 0&&(_t=Tt.clone(),gt[mt]=_t,st.addEventListener("dispose",ot)),Tt=_t}if(Tt.visible=st.visible,Tt.wireframe=st.wireframe,U===f.RyA?Tt.side=st.shadowSide!==null?st.shadowSide:st.side:Tt.side=st.shadowSide!==null?st.shadowSide:K[st.side],Tt.alphaMap=st.alphaMap,Tt.alphaTest=st.alphaToCoverage===!0?.5:st.alphaTest,Tt.map=st.map,Tt.clipShadows=st.clipShadows,Tt.clippingPlanes=st.clippingPlanes,Tt.clipIntersection=st.clipIntersection,Tt.displacementMap=st.displacementMap,Tt.displacementScale=st.displacementScale,Tt.displacementBias=st.displacementBias,Tt.wireframeLinewidth=st.wireframeLinewidth,Tt.linewidth=st.linewidth,R.isPointLight===!0&&Tt.isMeshDistanceMaterial===!0){const dt=o.properties.get(Tt);dt.light=R}return Tt}function H(J,st,R,U,Tt){if(J.visible===!1)return;if(J.layers.test(st.layers)&&(J.isMesh||J.isLine||J.isPoints)&&(J.castShadow||J.receiveShadow&&Tt===f.RyA)&&(!J.frustumCulled||x.intersectsObject(J))){J.modelViewMatrix.multiplyMatrices(R.matrixWorldInverse,J.matrixWorld);const mt=p.update(J),gt=J.material;if(Array.isArray(gt)){const _t=mt.groups;for(let ct=0,it=_t.length;ct<it;ct++){const It=_t[ct],zt=gt[It.materialIndex];if(zt&&zt.visible){const kt=q(J,zt,U,Tt);J.onBeforeShadow(o,J,st,R,mt,kt,It),o.renderBufferDirect(R,null,mt,kt,J,It),J.onAfterShadow(o,J,st,R,mt,kt,It)}}}else if(gt.visible){const _t=q(J,gt,U,Tt);J.onBeforeShadow(o,J,st,R,mt,_t,null),o.renderBufferDirect(R,null,mt,_t,J,null),J.onAfterShadow(o,J,st,R,mt,_t,null)}}const dt=J.children;for(let mt=0,gt=dt.length;mt<gt;mt++)H(dt[mt],st,R,U,Tt)}function ot(J){J.target.removeEventListener("dispose",ot);for(const R in $){const U=$[R],Tt=J.target.uuid;Tt in U&&(U[Tt].dispose(),delete U[Tt])}}}function Za(o,p){function h(){let k=!1;const Rt=new f.IUQ;let Ct=null;const Xt=new f.IUQ(0,0,0,0);return{setMask:function(Et){Ct!==Et&&!k&&(o.colorMask(Et,Et,Et,Et),Ct=Et)},setLocked:function(Et){k=Et},setClear:function(Et,ft,$t,me,Ge){Ge===!0&&(Et*=me,ft*=me,$t*=me),Rt.set(Et,ft,$t,me),Xt.equals(Rt)===!1&&(o.clearColor(Et,ft,$t,me),Xt.copy(Rt))},reset:function(){k=!1,Ct=null,Xt.set(-1,0,0,0)}}}function x(){let k=!1,Rt=!1,Ct=null,Xt=null,Et=null;return{setReversed:function(ft){if(Rt!==ft){const $t=p.get("EXT_clip_control");ft?$t.clipControlEXT($t.LOWER_LEFT_EXT,$t.ZERO_TO_ONE_EXT):$t.clipControlEXT($t.LOWER_LEFT_EXT,$t.NEGATIVE_ONE_TO_ONE_EXT),Rt=ft;const me=Et;Et=null,this.setClear(me)}},getReversed:function(){return Rt},setTest:function(ft){ft?Lt(o.DEPTH_TEST):Dt(o.DEPTH_TEST)},setMask:function(ft){Ct!==ft&&!k&&(o.depthMask(ft),Ct=ft)},setFunc:function(ft){if(Rt&&(ft=f.ri6[ft]),Xt!==ft){switch(ft){case f.eHc:o.depthFunc(o.NEVER);break;case f.lGu:o.depthFunc(o.ALWAYS);break;case f.brA:o.depthFunc(o.LESS);break;case f.xSv:o.depthFunc(o.LEQUAL);break;case f.U3G:o.depthFunc(o.EQUAL);break;case f.Gwm:o.depthFunc(o.GEQUAL);break;case f.K52:o.depthFunc(o.GREATER);break;case f.bw0:o.depthFunc(o.NOTEQUAL);break;default:o.depthFunc(o.LEQUAL)}Xt=ft}},setLocked:function(ft){k=ft},setClear:function(ft){Et!==ft&&(Et=ft,Rt&&(ft=1-ft),o.clearDepth(ft))},reset:function(){k=!1,Ct=null,Xt=null,Et=null,Rt=!1}}}function E(){let k=!1,Rt=null,Ct=null,Xt=null,Et=null,ft=null,$t=null,me=null,Ge=null;return{setTest:function(ze){k||(ze?Lt(o.STENCIL_TEST):Dt(o.STENCIL_TEST))},setMask:function(ze){Rt!==ze&&!k&&(o.stencilMask(ze),Rt=ze)},setFunc:function(ze,fn,zn){(Ct!==ze||Xt!==fn||Et!==zn)&&(o.stencilFunc(ze,fn,zn),Ct=ze,Xt=fn,Et=zn)},setOp:function(ze,fn,zn){(ft!==ze||$t!==fn||me!==zn)&&(o.stencilOp(ze,fn,zn),ft=ze,$t=fn,me=zn)},setLocked:function(ze){k=ze},setClear:function(ze){Ge!==ze&&(o.clearStencil(ze),Ge=ze)},reset:function(){k=!1,Rt=null,Ct=null,Xt=null,Et=null,ft=null,$t=null,me=null,Ge=null}}}const S=new h,I=new x,V=new E,j=new WeakMap,$=new WeakMap;let ht={},K={},F=new WeakMap,G=[],X=null,nt=!1,A=null,b=null,z=null,q=null,H=null,ot=null,J=null,st=new f.Q1f(0,0,0),R=0,U=!1,Tt=null,Z=null,dt=null,mt=null,gt=null;const _t=o.getParameter(o.MAX_COMBINED_TEXTURE_IMAGE_UNITS);let ct=!1,it=0;const It=o.getParameter(o.VERSION);It.indexOf("WebGL")!==-1?(it=parseFloat(/^WebGL (\d)/.exec(It)[1]),ct=it>=1):It.indexOf("OpenGL ES")!==-1&&(it=parseFloat(/^OpenGL ES (\d)/.exec(It)[1]),ct=it>=2);let zt=null,kt={};const ue=o.getParameter(o.SCISSOR_BOX),te=o.getParameter(o.VIEWPORT),se=new f.IUQ().fromArray(ue),Ze=new f.IUQ().fromArray(te);function $e(k,Rt,Ct,Xt){const Et=new Uint8Array(4),ft=o.createTexture();o.bindTexture(k,ft),o.texParameteri(k,o.TEXTURE_MIN_FILTER,o.NEAREST),o.texParameteri(k,o.TEXTURE_MAG_FILTER,o.NEAREST);for(let $t=0;$t<Ct;$t++)k===o.TEXTURE_3D||k===o.TEXTURE_2D_ARRAY?o.texImage3D(Rt,0,o.RGBA,1,1,Xt,0,o.RGBA,o.UNSIGNED_BYTE,Et):o.texImage2D(Rt+$t,0,o.RGBA,1,1,0,o.RGBA,o.UNSIGNED_BYTE,Et);return ft}const xt={};xt[o.TEXTURE_2D]=$e(o.TEXTURE_2D,o.TEXTURE_2D,1),xt[o.TEXTURE_CUBE_MAP]=$e(o.TEXTURE_CUBE_MAP,o.TEXTURE_CUBE_MAP_POSITIVE_X,6),xt[o.TEXTURE_2D_ARRAY]=$e(o.TEXTURE_2D_ARRAY,o.TEXTURE_2D_ARRAY,1,1),xt[o.TEXTURE_3D]=$e(o.TEXTURE_3D,o.TEXTURE_3D,1,1),S.setClear(0,0,0,1),I.setClear(1),V.setClear(0),Lt(o.DEPTH_TEST),I.setFunc(f.xSv),ae(!1),Je(f.Vb5),Lt(o.CULL_FACE),Pe(f.XIg);function Lt(k){ht[k]!==!0&&(o.enable(k),ht[k]=!0)}function Dt(k){ht[k]!==!1&&(o.disable(k),ht[k]=!1)}function Te(k,Rt){return K[k]!==Rt?(o.bindFramebuffer(k,Rt),K[k]=Rt,k===o.DRAW_FRAMEBUFFER&&(K[o.FRAMEBUFFER]=Rt),k===o.FRAMEBUFFER&&(K[o.DRAW_FRAMEBUFFER]=Rt),!0):!1}function de(k,Rt){let Ct=G,Xt=!1;if(k){Ct=F.get(Rt),Ct===void 0&&(Ct=[],F.set(Rt,Ct));const Et=k.textures;if(Ct.length!==Et.length||Ct[0]!==o.COLOR_ATTACHMENT0){for(let ft=0,$t=Et.length;ft<$t;ft++)Ct[ft]=o.COLOR_ATTACHMENT0+ft;Ct.length=Et.length,Xt=!0}}else Ct[0]!==o.BACK&&(Ct[0]=o.BACK,Xt=!0);Xt&&o.drawBuffers(Ct)}function _e(k){return X!==k?(o.useProgram(k),X=k,!0):!1}const Ut={[f.gO9]:o.FUNC_ADD,[f.FXf]:o.FUNC_SUBTRACT,[f.nST]:o.FUNC_REVERSE_SUBTRACT};Ut[f.znC]=o.MIN,Ut[f.$ei]=o.MAX;const Ie={[f.ojh]:o.ZERO,[f.qad]:o.ONE,[f.f4X]:o.SRC_COLOR,[f.ie2]:o.SRC_ALPHA,[f.hgQ]:o.SRC_ALPHA_SATURATE,[f.wn6]:o.DST_COLOR,[f.hdd]:o.DST_ALPHA,[f.LiQ]:o.ONE_MINUS_SRC_COLOR,[f.OuU]:o.ONE_MINUS_SRC_ALPHA,[f.aEY]:o.ONE_MINUS_DST_COLOR,[f.Nt7]:o.ONE_MINUS_DST_ALPHA,[f.RrE]:o.CONSTANT_COLOR,[f.$Yl]:o.ONE_MINUS_CONSTANT_COLOR,[f.e0p]:o.CONSTANT_ALPHA,[f.ov9]:o.ONE_MINUS_CONSTANT_ALPHA};function Pe(k,Rt,Ct,Xt,Et,ft,$t,me,Ge,ze){if(k===f.XIg){nt===!0&&(Dt(o.BLEND),nt=!1);return}if(nt===!1&&(Lt(o.BLEND),nt=!0),k!==f.bCz){if(k!==A||ze!==U){if((b!==f.gO9||H!==f.gO9)&&(o.blendEquation(o.FUNC_ADD),b=f.gO9,H=f.gO9),ze)switch(k){case f.NTi:o.blendFuncSeparate(o.ONE,o.ONE_MINUS_SRC_ALPHA,o.ONE,o.ONE_MINUS_SRC_ALPHA);break;case f.EZo:o.blendFunc(o.ONE,o.ONE);break;case f.Kwu:o.blendFuncSeparate(o.ZERO,o.ONE_MINUS_SRC_COLOR,o.ZERO,o.ONE);break;case f.EdD:o.blendFuncSeparate(o.DST_COLOR,o.ONE_MINUS_SRC_ALPHA,o.ZERO,o.ONE);break;default:(0,f.z3S)("WebGLState: Invalid blending: ",k);break}else switch(k){case f.NTi:o.blendFuncSeparate(o.SRC_ALPHA,o.ONE_MINUS_SRC_ALPHA,o.ONE,o.ONE_MINUS_SRC_ALPHA);break;case f.EZo:o.blendFuncSeparate(o.SRC_ALPHA,o.ONE,o.ONE,o.ONE);break;case f.Kwu:(0,f.z3S)("WebGLState: SubtractiveBlending requires material.premultipliedAlpha = true");break;case f.EdD:(0,f.z3S)("WebGLState: MultiplyBlending requires material.premultipliedAlpha = true");break;default:(0,f.z3S)("WebGLState: Invalid blending: ",k);break}z=null,q=null,ot=null,J=null,st.set(0,0,0),R=0,A=k,U=ze}return}Et=Et||Rt,ft=ft||Ct,$t=$t||Xt,(Rt!==b||Et!==H)&&(o.blendEquationSeparate(Ut[Rt],Ut[Et]),b=Rt,H=Et),(Ct!==z||Xt!==q||ft!==ot||$t!==J)&&(o.blendFuncSeparate(Ie[Ct],Ie[Xt],Ie[ft],Ie[$t]),z=Ct,q=Xt,ot=ft,J=$t),(me.equals(st)===!1||Ge!==R)&&(o.blendColor(me.r,me.g,me.b,Ge),st.copy(me),R=Ge),A=k,U=!1}function Be(k,Rt){k.side===f.$EB?Dt(o.CULL_FACE):Lt(o.CULL_FACE);let Ct=k.side===f.hsX;Rt&&(Ct=!Ct),ae(Ct),k.blending===f.NTi&&k.transparent===!1?Pe(f.XIg):Pe(k.blending,k.blendEquation,k.blendSrc,k.blendDst,k.blendEquationAlpha,k.blendSrcAlpha,k.blendDstAlpha,k.blendColor,k.blendAlpha,k.premultipliedAlpha),I.setFunc(k.depthFunc),I.setTest(k.depthTest),I.setMask(k.depthWrite),S.setMask(k.colorWrite);const Xt=k.stencilWrite;V.setTest(Xt),Xt&&(V.setMask(k.stencilWriteMask),V.setFunc(k.stencilFunc,k.stencilRef,k.stencilFuncMask),V.setOp(k.stencilFail,k.stencilZFail,k.stencilZPass)),Xe(k.polygonOffset,k.polygonOffsetFactor,k.polygonOffsetUnits),k.alphaToCoverage===!0?Lt(o.SAMPLE_ALPHA_TO_COVERAGE):Dt(o.SAMPLE_ALPHA_TO_COVERAGE)}function ae(k){Tt!==k&&(k?o.frontFace(o.CW):o.frontFace(o.CCW),Tt=k)}function Je(k){k!==f.WNZ?(Lt(o.CULL_FACE),k!==Z&&(k===f.Vb5?o.cullFace(o.BACK):k===f.Jnc?o.cullFace(o.FRONT):o.cullFace(o.FRONT_AND_BACK))):Dt(o.CULL_FACE),Z=k}function B(k){k!==dt&&(ct&&o.lineWidth(k),dt=k)}function Xe(k,Rt,Ct){k?(Lt(o.POLYGON_OFFSET_FILL),(mt!==Rt||gt!==Ct)&&(mt=Rt,gt=Ct,I.getReversed()&&(Rt=-Rt),o.polygonOffset(Rt,Ct))):Dt(o.POLYGON_OFFSET_FILL)}function Ce(k){k?Lt(o.SCISSOR_TEST):Dt(o.SCISSOR_TEST)}function Ee(k){k===void 0&&(k=o.TEXTURE0+_t-1),zt!==k&&(o.activeTexture(k),zt=k)}function Jt(k,Rt,Ct){Ct===void 0&&(zt===null?Ct=o.TEXTURE0+_t-1:Ct=zt);let Xt=kt[Ct];Xt===void 0&&(Xt={type:void 0,texture:void 0},kt[Ct]=Xt),(Xt.type!==k||Xt.texture!==Rt)&&(zt!==Ct&&(o.activeTexture(Ct),zt=Ct),o.bindTexture(k,Rt||xt[k]),Xt.type=k,Xt.texture=Rt)}function P(){const k=kt[zt];k!==void 0&&k.type!==void 0&&(o.bindTexture(k.type,null),k.type=void 0,k.texture=void 0)}function y(){try{o.compressedTexImage2D(...arguments)}catch(k){(0,f.z3S)("WebGLState:",k)}}function W(){try{o.compressedTexImage3D(...arguments)}catch(k){(0,f.z3S)("WebGLState:",k)}}function ut(){try{o.texSubImage2D(...arguments)}catch(k){(0,f.z3S)("WebGLState:",k)}}function yt(){try{o.texSubImage3D(...arguments)}catch(k){(0,f.z3S)("WebGLState:",k)}}function pt(){try{o.compressedTexSubImage2D(...arguments)}catch(k){(0,f.z3S)("WebGLState:",k)}}function Ft(){try{o.compressedTexSubImage3D(...arguments)}catch(k){(0,f.z3S)("WebGLState:",k)}}function Pt(){try{o.texStorage2D(...arguments)}catch(k){(0,f.z3S)("WebGLState:",k)}}function ne(){try{o.texStorage3D(...arguments)}catch(k){(0,f.z3S)("WebGLState:",k)}}function ce(){try{o.texImage2D(...arguments)}catch(k){(0,f.z3S)("WebGLState:",k)}}function bt(){try{o.texImage3D(...arguments)}catch(k){(0,f.z3S)("WebGLState:",k)}}function wt(k){se.equals(k)===!1&&(o.scissor(k.x,k.y,k.z,k.w),se.copy(k))}function qt(k){Ze.equals(k)===!1&&(o.viewport(k.x,k.y,k.z,k.w),Ze.copy(k))}function Zt(k,Rt){let Ct=$.get(Rt);Ct===void 0&&(Ct=new WeakMap,$.set(Rt,Ct));let Xt=Ct.get(k);Xt===void 0&&(Xt=o.getUniformBlockIndex(Rt,k.name),Ct.set(k,Xt))}function Ht(k,Rt){const Xt=$.get(Rt).get(k);j.get(Rt)!==Xt&&(o.uniformBlockBinding(Rt,Xt,k.__bindingPointIndex),j.set(Rt,Xt))}function Se(){o.disable(o.BLEND),o.disable(o.CULL_FACE),o.disable(o.DEPTH_TEST),o.disable(o.POLYGON_OFFSET_FILL),o.disable(o.SCISSOR_TEST),o.disable(o.STENCIL_TEST),o.disable(o.SAMPLE_ALPHA_TO_COVERAGE),o.blendEquation(o.FUNC_ADD),o.blendFunc(o.ONE,o.ZERO),o.blendFuncSeparate(o.ONE,o.ZERO,o.ONE,o.ZERO),o.blendColor(0,0,0,0),o.colorMask(!0,!0,!0,!0),o.clearColor(0,0,0,0),o.depthMask(!0),o.depthFunc(o.LESS),I.setReversed(!1),o.clearDepth(1),o.stencilMask(4294967295),o.stencilFunc(o.ALWAYS,0,4294967295),o.stencilOp(o.KEEP,o.KEEP,o.KEEP),o.clearStencil(0),o.cullFace(o.BACK),o.frontFace(o.CCW),o.polygonOffset(0,0),o.activeTexture(o.TEXTURE0),o.bindFramebuffer(o.FRAMEBUFFER,null),o.bindFramebuffer(o.DRAW_FRAMEBUFFER,null),o.bindFramebuffer(o.READ_FRAMEBUFFER,null),o.useProgram(null),o.lineWidth(1),o.scissor(0,0,o.canvas.width,o.canvas.height),o.viewport(0,0,o.canvas.width,o.canvas.height),ht={},zt=null,kt={},K={},F=new WeakMap,G=[],X=null,nt=!1,A=null,b=null,z=null,q=null,H=null,ot=null,J=null,st=new f.Q1f(0,0,0),R=0,U=!1,Tt=null,Z=null,dt=null,mt=null,gt=null,se.set(0,0,o.canvas.width,o.canvas.height),Ze.set(0,0,o.canvas.width,o.canvas.height),S.reset(),I.reset(),V.reset()}return{buffers:{color:S,depth:I,stencil:V},enable:Lt,disable:Dt,bindFramebuffer:Te,drawBuffers:de,useProgram:_e,setBlending:Pe,setMaterial:Be,setFlipSided:ae,setCullFace:Je,setLineWidth:B,setPolygonOffset:Xe,setScissorTest:Ce,activeTexture:Ee,bindTexture:Jt,unbindTexture:P,compressedTexImage2D:y,compressedTexImage3D:W,texImage2D:ce,texImage3D:bt,updateUBOMapping:Zt,uniformBlockBinding:Ht,texStorage2D:Pt,texStorage3D:ne,texSubImage2D:ut,texSubImage3D:yt,compressedTexSubImage2D:pt,compressedTexSubImage3D:Ft,scissor:wt,viewport:qt,reset:Se}}function Lc(o,p,h,x,E,S,I){const V=p.has("WEBGL_multisampled_render_to_texture")?p.get("WEBGL_multisampled_render_to_texture"):null,j=typeof navigator>"u"?!1:/OculusBrowser/g.test(navigator.userAgent),$=new f.I9Y,ht=new WeakMap;let K;const F=new WeakMap;let G=!1;try{G=typeof OffscreenCanvas<"u"&&new OffscreenCanvas(1,1).getContext("2d")!==null}catch{}function X(P,y){return G?new OffscreenCanvas(P,y):(0,f.qq$)("canvas")}function nt(P,y,W){let ut=1;const yt=Jt(P);if((yt.width>W||yt.height>W)&&(ut=W/Math.max(yt.width,yt.height)),ut<1)if(typeof HTMLImageElement<"u"&&P instanceof HTMLImageElement||typeof HTMLCanvasElement<"u"&&P instanceof HTMLCanvasElement||typeof ImageBitmap<"u"&&P instanceof ImageBitmap||typeof VideoFrame<"u"&&P instanceof VideoFrame){const pt=Math.floor(ut*yt.width),Ft=Math.floor(ut*yt.height);K===void 0&&(K=X(pt,Ft));const Pt=y?X(pt,Ft):K;return Pt.width=pt,Pt.height=Ft,Pt.getContext("2d").drawImage(P,0,0,pt,Ft),(0,f.R8M)("WebGLRenderer: Texture has been resized from ("+yt.width+"x"+yt.height+") to ("+pt+"x"+Ft+")."),Pt}else return"data"in P&&(0,f.R8M)("WebGLRenderer: Image in DataTexture is too big ("+yt.width+"x"+yt.height+")."),P;return P}function A(P){return P.generateMipmaps}function b(P){o.generateMipmap(P)}function z(P){return P.isWebGLCubeRenderTarget?o.TEXTURE_CUBE_MAP:P.isWebGL3DRenderTarget?o.TEXTURE_3D:P.isWebGLArrayRenderTarget||P.isCompressedArrayTexture?o.TEXTURE_2D_ARRAY:o.TEXTURE_2D}function q(P,y,W,ut,yt=!1){if(P!==null){if(o[P]!==void 0)return o[P];(0,f.R8M)("WebGLRenderer: Attempt to use non-existing WebGL internal format '"+P+"'")}let pt=y;if(y===o.RED&&(W===o.FLOAT&&(pt=o.R32F),W===o.HALF_FLOAT&&(pt=o.R16F),W===o.UNSIGNED_BYTE&&(pt=o.R8)),y===o.RED_INTEGER&&(W===o.UNSIGNED_BYTE&&(pt=o.R8UI),W===o.UNSIGNED_SHORT&&(pt=o.R16UI),W===o.UNSIGNED_INT&&(pt=o.R32UI),W===o.BYTE&&(pt=o.R8I),W===o.SHORT&&(pt=o.R16I),W===o.INT&&(pt=o.R32I)),y===o.RG&&(W===o.FLOAT&&(pt=o.RG32F),W===o.HALF_FLOAT&&(pt=o.RG16F),W===o.UNSIGNED_BYTE&&(pt=o.RG8)),y===o.RG_INTEGER&&(W===o.UNSIGNED_BYTE&&(pt=o.RG8UI),W===o.UNSIGNED_SHORT&&(pt=o.RG16UI),W===o.UNSIGNED_INT&&(pt=o.RG32UI),W===o.BYTE&&(pt=o.RG8I),W===o.SHORT&&(pt=o.RG16I),W===o.INT&&(pt=o.RG32I)),y===o.RGB_INTEGER&&(W===o.UNSIGNED_BYTE&&(pt=o.RGB8UI),W===o.UNSIGNED_SHORT&&(pt=o.RGB16UI),W===o.UNSIGNED_INT&&(pt=o.RGB32UI),W===o.BYTE&&(pt=o.RGB8I),W===o.SHORT&&(pt=o.RGB16I),W===o.INT&&(pt=o.RGB32I)),y===o.RGBA_INTEGER&&(W===o.UNSIGNED_BYTE&&(pt=o.RGBA8UI),W===o.UNSIGNED_SHORT&&(pt=o.RGBA16UI),W===o.UNSIGNED_INT&&(pt=o.RGBA32UI),W===o.BYTE&&(pt=o.RGBA8I),W===o.SHORT&&(pt=o.RGBA16I),W===o.INT&&(pt=o.RGBA32I)),y===o.RGB&&(W===o.UNSIGNED_INT_5_9_9_9_REV&&(pt=o.RGB9_E5),W===o.UNSIGNED_INT_10F_11F_11F_REV&&(pt=o.R11F_G11F_B10F)),y===o.RGBA){const Ft=yt?f.VxR:f.ppV.getTransfer(ut);W===o.FLOAT&&(pt=o.RGBA32F),W===o.HALF_FLOAT&&(pt=o.RGBA16F),W===o.UNSIGNED_BYTE&&(pt=Ft===f.KLL?o.SRGB8_ALPHA8:o.RGBA8),W===o.UNSIGNED_SHORT_4_4_4_4&&(pt=o.RGBA4),W===o.UNSIGNED_SHORT_5_5_5_1&&(pt=o.RGB5_A1)}return(pt===o.R16F||pt===o.R32F||pt===o.RG16F||pt===o.RG32F||pt===o.RGBA16F||pt===o.RGBA32F)&&p.get("EXT_color_buffer_float"),pt}function H(P,y){let W;return P?y===null||y===f.bkx||y===f.V3x?W=o.DEPTH24_STENCIL8:y===f.RQf?W=o.DEPTH32F_STENCIL8:y===f.cHt&&(W=o.DEPTH24_STENCIL8,(0,f.R8M)("DepthTexture: 16 bit depth attachment is not supported with stencil. Using 24-bit attachment.")):y===null||y===f.bkx||y===f.V3x?W=o.DEPTH_COMPONENT24:y===f.RQf?W=o.DEPTH_COMPONENT32F:y===f.cHt&&(W=o.DEPTH_COMPONENT16),W}function ot(P,y){return A(P)===!0||P.isFramebufferTexture&&P.minFilter!==f.hxR&&P.minFilter!==f.k6q?Math.log2(Math.max(y.width,y.height))+1:P.mipmaps!==void 0&&P.mipmaps.length>0?P.mipmaps.length:P.isCompressedTexture&&Array.isArray(P.image)?y.mipmaps.length:1}function J(P){const y=P.target;y.removeEventListener("dispose",J),R(y),y.isVideoTexture&&ht.delete(y)}function st(P){const y=P.target;y.removeEventListener("dispose",st),Tt(y)}function R(P){const y=x.get(P);if(y.__webglInit===void 0)return;const W=P.source,ut=F.get(W);if(ut){const yt=ut[y.__cacheKey];yt.usedTimes--,yt.usedTimes===0&&U(P),Object.keys(ut).length===0&&F.delete(W)}x.remove(P)}function U(P){const y=x.get(P);o.deleteTexture(y.__webglTexture);const W=P.source,ut=F.get(W);delete ut[y.__cacheKey],I.memory.textures--}function Tt(P){const y=x.get(P);if(P.depthTexture&&(P.depthTexture.dispose(),x.remove(P.depthTexture)),P.isWebGLCubeRenderTarget)for(let ut=0;ut<6;ut++){if(Array.isArray(y.__webglFramebuffer[ut]))for(let yt=0;yt<y.__webglFramebuffer[ut].length;yt++)o.deleteFramebuffer(y.__webglFramebuffer[ut][yt]);else o.deleteFramebuffer(y.__webglFramebuffer[ut]);y.__webglDepthbuffer&&o.deleteRenderbuffer(y.__webglDepthbuffer[ut])}else{if(Array.isArray(y.__webglFramebuffer))for(let ut=0;ut<y.__webglFramebuffer.length;ut++)o.deleteFramebuffer(y.__webglFramebuffer[ut]);else o.deleteFramebuffer(y.__webglFramebuffer);if(y.__webglDepthbuffer&&o.deleteRenderbuffer(y.__webglDepthbuffer),y.__webglMultisampledFramebuffer&&o.deleteFramebuffer(y.__webglMultisampledFramebuffer),y.__webglColorRenderbuffer)for(let ut=0;ut<y.__webglColorRenderbuffer.length;ut++)y.__webglColorRenderbuffer[ut]&&o.deleteRenderbuffer(y.__webglColorRenderbuffer[ut]);y.__webglDepthRenderbuffer&&o.deleteRenderbuffer(y.__webglDepthRenderbuffer)}const W=P.textures;for(let ut=0,yt=W.length;ut<yt;ut++){const pt=x.get(W[ut]);pt.__webglTexture&&(o.deleteTexture(pt.__webglTexture),I.memory.textures--),x.remove(W[ut])}x.remove(P)}let Z=0;function dt(){Z=0}function mt(){const P=Z;return P>=E.maxTextures&&(0,f.R8M)("WebGLTextures: Trying to use "+P+" texture units while this GPU supports only "+E.maxTextures),Z+=1,P}function gt(P){const y=[];return y.push(P.wrapS),y.push(P.wrapT),y.push(P.wrapR||0),y.push(P.magFilter),y.push(P.minFilter),y.push(P.anisotropy),y.push(P.internalFormat),y.push(P.format),y.push(P.type),y.push(P.generateMipmaps),y.push(P.premultiplyAlpha),y.push(P.flipY),y.push(P.unpackAlignment),y.push(P.colorSpace),y.join()}function _t(P,y){const W=x.get(P);if(P.isVideoTexture&&Ce(P),P.isRenderTargetTexture===!1&&P.isExternalTexture!==!0&&P.version>0&&W.__version!==P.version){const ut=P.image;if(ut===null)(0,f.R8M)("WebGLRenderer: Texture marked for update but no image data found.");else if(ut.complete===!1)(0,f.R8M)("WebGLRenderer: Texture marked for update but image is incomplete");else{xt(W,P,y);return}}else P.isExternalTexture&&(W.__webglTexture=P.sourceTexture?P.sourceTexture:null);h.bindTexture(o.TEXTURE_2D,W.__webglTexture,o.TEXTURE0+y)}function ct(P,y){const W=x.get(P);if(P.isRenderTargetTexture===!1&&P.version>0&&W.__version!==P.version){xt(W,P,y);return}else P.isExternalTexture&&(W.__webglTexture=P.sourceTexture?P.sourceTexture:null);h.bindTexture(o.TEXTURE_2D_ARRAY,W.__webglTexture,o.TEXTURE0+y)}function it(P,y){const W=x.get(P);if(P.isRenderTargetTexture===!1&&P.version>0&&W.__version!==P.version){xt(W,P,y);return}h.bindTexture(o.TEXTURE_3D,W.__webglTexture,o.TEXTURE0+y)}function It(P,y){const W=x.get(P);if(P.isCubeDepthTexture!==!0&&P.version>0&&W.__version!==P.version){Lt(W,P,y);return}h.bindTexture(o.TEXTURE_CUBE_MAP,W.__webglTexture,o.TEXTURE0+y)}const zt={[f.GJx]:o.REPEAT,[f.ghU]:o.CLAMP_TO_EDGE,[f.kTW]:o.MIRRORED_REPEAT},kt={[f.hxR]:o.NEAREST,[f.pHI]:o.NEAREST_MIPMAP_NEAREST,[f.Cfg]:o.NEAREST_MIPMAP_LINEAR,[f.k6q]:o.LINEAR,[f.kRr]:o.LINEAR_MIPMAP_NEAREST,[f.$_I]:o.LINEAR_MIPMAP_LINEAR},ue={[f.amv]:o.NEVER,[f.FFZ]:o.ALWAYS,[f.vim]:o.LESS,[f.TiK]:o.LEQUAL,[f.kO0]:o.EQUAL,[f.gWB]:o.GEQUAL,[f.eoi]:o.GREATER,[f.jzd]:o.NOTEQUAL};function te(P,y){if(y.type===f.RQf&&p.has("OES_texture_float_linear")===!1&&(y.magFilter===f.k6q||y.magFilter===f.kRr||y.magFilter===f.Cfg||y.magFilter===f.$_I||y.minFilter===f.k6q||y.minFilter===f.kRr||y.minFilter===f.Cfg||y.minFilter===f.$_I)&&(0,f.R8M)("WebGLRenderer: Unable to use linear filtering with floating point textures. OES_texture_float_linear not supported on this device."),o.texParameteri(P,o.TEXTURE_WRAP_S,zt[y.wrapS]),o.texParameteri(P,o.TEXTURE_WRAP_T,zt[y.wrapT]),(P===o.TEXTURE_3D||P===o.TEXTURE_2D_ARRAY)&&o.texParameteri(P,o.TEXTURE_WRAP_R,zt[y.wrapR]),o.texParameteri(P,o.TEXTURE_MAG_FILTER,kt[y.magFilter]),o.texParameteri(P,o.TEXTURE_MIN_FILTER,kt[y.minFilter]),y.compareFunction&&(o.texParameteri(P,o.TEXTURE_COMPARE_MODE,o.COMPARE_REF_TO_TEXTURE),o.texParameteri(P,o.TEXTURE_COMPARE_FUNC,ue[y.compareFunction])),p.has("EXT_texture_filter_anisotropic")===!0){if(y.magFilter===f.hxR||y.minFilter!==f.Cfg&&y.minFilter!==f.$_I||y.type===f.RQf&&p.has("OES_texture_float_linear")===!1)return;if(y.anisotropy>1||x.get(y).__currentAnisotropy){const W=p.get("EXT_texture_filter_anisotropic");o.texParameterf(P,W.TEXTURE_MAX_ANISOTROPY_EXT,Math.min(y.anisotropy,E.getMaxAnisotropy())),x.get(y).__currentAnisotropy=y.anisotropy}}}function se(P,y){let W=!1;P.__webglInit===void 0&&(P.__webglInit=!0,y.addEventListener("dispose",J));const ut=y.source;let yt=F.get(ut);yt===void 0&&(yt={},F.set(ut,yt));const pt=gt(y);if(pt!==P.__cacheKey){yt[pt]===void 0&&(yt[pt]={texture:o.createTexture(),usedTimes:0},I.memory.textures++,W=!0),yt[pt].usedTimes++;const Ft=yt[P.__cacheKey];Ft!==void 0&&(yt[P.__cacheKey].usedTimes--,Ft.usedTimes===0&&U(y)),P.__cacheKey=pt,P.__webglTexture=yt[pt].texture}return W}function Ze(P,y,W){return Math.floor(Math.floor(P/W)/y)}function $e(P,y,W,ut){const pt=P.updateRanges;if(pt.length===0)h.texSubImage2D(o.TEXTURE_2D,0,0,0,y.width,y.height,W,ut,y.data);else{pt.sort((bt,wt)=>bt.start-wt.start);let Ft=0;for(let bt=1;bt<pt.length;bt++){const wt=pt[Ft],qt=pt[bt],Zt=wt.start+wt.count,Ht=Ze(qt.start,y.width,4),Se=Ze(wt.start,y.width,4);qt.start<=Zt+1&&Ht===Se&&Ze(qt.start+qt.count-1,y.width,4)===Ht?wt.count=Math.max(wt.count,qt.start+qt.count-wt.start):(++Ft,pt[Ft]=qt)}pt.length=Ft+1;const Pt=o.getParameter(o.UNPACK_ROW_LENGTH),ne=o.getParameter(o.UNPACK_SKIP_PIXELS),ce=o.getParameter(o.UNPACK_SKIP_ROWS);o.pixelStorei(o.UNPACK_ROW_LENGTH,y.width);for(let bt=0,wt=pt.length;bt<wt;bt++){const qt=pt[bt],Zt=Math.floor(qt.start/4),Ht=Math.ceil(qt.count/4),Se=Zt%y.width,k=Math.floor(Zt/y.width),Rt=Ht,Ct=1;o.pixelStorei(o.UNPACK_SKIP_PIXELS,Se),o.pixelStorei(o.UNPACK_SKIP_ROWS,k),h.texSubImage2D(o.TEXTURE_2D,0,Se,k,Rt,Ct,W,ut,y.data)}P.clearUpdateRanges(),o.pixelStorei(o.UNPACK_ROW_LENGTH,Pt),o.pixelStorei(o.UNPACK_SKIP_PIXELS,ne),o.pixelStorei(o.UNPACK_SKIP_ROWS,ce)}}function xt(P,y,W){let ut=o.TEXTURE_2D;(y.isDataArrayTexture||y.isCompressedArrayTexture)&&(ut=o.TEXTURE_2D_ARRAY),y.isData3DTexture&&(ut=o.TEXTURE_3D);const yt=se(P,y),pt=y.source;h.bindTexture(ut,P.__webglTexture,o.TEXTURE0+W);const Ft=x.get(pt);if(pt.version!==Ft.__version||yt===!0){h.activeTexture(o.TEXTURE0+W);const Pt=f.ppV.getPrimaries(f.ppV.workingColorSpace),ne=y.colorSpace===f.jf0?null:f.ppV.getPrimaries(y.colorSpace),ce=y.colorSpace===f.jf0||Pt===ne?o.NONE:o.BROWSER_DEFAULT_WEBGL;o.pixelStorei(o.UNPACK_FLIP_Y_WEBGL,y.flipY),o.pixelStorei(o.UNPACK_PREMULTIPLY_ALPHA_WEBGL,y.premultiplyAlpha),o.pixelStorei(o.UNPACK_ALIGNMENT,y.unpackAlignment),o.pixelStorei(o.UNPACK_COLORSPACE_CONVERSION_WEBGL,ce);let bt=nt(y.image,!1,E.maxTextureSize);bt=Ee(y,bt);const wt=S.convert(y.format,y.colorSpace),qt=S.convert(y.type);let Zt=q(y.internalFormat,wt,qt,y.colorSpace,y.isVideoTexture);te(ut,y);let Ht;const Se=y.mipmaps,k=y.isVideoTexture!==!0,Rt=Ft.__version===void 0||yt===!0,Ct=pt.dataReady,Xt=ot(y,bt);if(y.isDepthTexture)Zt=H(y.format===f.dcC,y.type),Rt&&(k?h.texStorage2D(o.TEXTURE_2D,1,Zt,bt.width,bt.height):h.texImage2D(o.TEXTURE_2D,0,Zt,bt.width,bt.height,0,wt,qt,null));else if(y.isDataTexture)if(Se.length>0){k&&Rt&&h.texStorage2D(o.TEXTURE_2D,Xt,Zt,Se[0].width,Se[0].height);for(let Et=0,ft=Se.length;Et<ft;Et++)Ht=Se[Et],k?Ct&&h.texSubImage2D(o.TEXTURE_2D,Et,0,0,Ht.width,Ht.height,wt,qt,Ht.data):h.texImage2D(o.TEXTURE_2D,Et,Zt,Ht.width,Ht.height,0,wt,qt,Ht.data);y.generateMipmaps=!1}else k?(Rt&&h.texStorage2D(o.TEXTURE_2D,Xt,Zt,bt.width,bt.height),Ct&&$e(y,bt,wt,qt)):h.texImage2D(o.TEXTURE_2D,0,Zt,bt.width,bt.height,0,wt,qt,bt.data);else if(y.isCompressedTexture)if(y.isCompressedArrayTexture){k&&Rt&&h.texStorage3D(o.TEXTURE_2D_ARRAY,Xt,Zt,Se[0].width,Se[0].height,bt.depth);for(let Et=0,ft=Se.length;Et<ft;Et++)if(Ht=Se[Et],y.format!==f.GWd)if(wt!==null)if(k){if(Ct)if(y.layerUpdates.size>0){const $t=(0,f.Nex)(Ht.width,Ht.height,y.format,y.type);for(const me of y.layerUpdates){const Ge=Ht.data.subarray(me*$t/Ht.data.BYTES_PER_ELEMENT,(me+1)*$t/Ht.data.BYTES_PER_ELEMENT);h.compressedTexSubImage3D(o.TEXTURE_2D_ARRAY,Et,0,0,me,Ht.width,Ht.height,1,wt,Ge)}y.clearLayerUpdates()}else h.compressedTexSubImage3D(o.TEXTURE_2D_ARRAY,Et,0,0,0,Ht.width,Ht.height,bt.depth,wt,Ht.data)}else h.compressedTexImage3D(o.TEXTURE_2D_ARRAY,Et,Zt,Ht.width,Ht.height,bt.depth,0,Ht.data,0,0);else(0,f.R8M)("WebGLRenderer: Attempt to load unsupported compressed texture format in .uploadTexture()");else k?Ct&&h.texSubImage3D(o.TEXTURE_2D_ARRAY,Et,0,0,0,Ht.width,Ht.height,bt.depth,wt,qt,Ht.data):h.texImage3D(o.TEXTURE_2D_ARRAY,Et,Zt,Ht.width,Ht.height,bt.depth,0,wt,qt,Ht.data)}else{k&&Rt&&h.texStorage2D(o.TEXTURE_2D,Xt,Zt,Se[0].width,Se[0].height);for(let Et=0,ft=Se.length;Et<ft;Et++)Ht=Se[Et],y.format!==f.GWd?wt!==null?k?Ct&&h.compressedTexSubImage2D(o.TEXTURE_2D,Et,0,0,Ht.width,Ht.height,wt,Ht.data):h.compressedTexImage2D(o.TEXTURE_2D,Et,Zt,Ht.width,Ht.height,0,Ht.data):(0,f.R8M)("WebGLRenderer: Attempt to load unsupported compressed texture format in .uploadTexture()"):k?Ct&&h.texSubImage2D(o.TEXTURE_2D,Et,0,0,Ht.width,Ht.height,wt,qt,Ht.data):h.texImage2D(o.TEXTURE_2D,Et,Zt,Ht.width,Ht.height,0,wt,qt,Ht.data)}else if(y.isDataArrayTexture)if(k){if(Rt&&h.texStorage3D(o.TEXTURE_2D_ARRAY,Xt,Zt,bt.width,bt.height,bt.depth),Ct)if(y.layerUpdates.size>0){const Et=(0,f.Nex)(bt.width,bt.height,y.format,y.type);for(const ft of y.layerUpdates){const $t=bt.data.subarray(ft*Et/bt.data.BYTES_PER_ELEMENT,(ft+1)*Et/bt.data.BYTES_PER_ELEMENT);h.texSubImage3D(o.TEXTURE_2D_ARRAY,0,0,0,ft,bt.width,bt.height,1,wt,qt,$t)}y.clearLayerUpdates()}else h.texSubImage3D(o.TEXTURE_2D_ARRAY,0,0,0,0,bt.width,bt.height,bt.depth,wt,qt,bt.data)}else h.texImage3D(o.TEXTURE_2D_ARRAY,0,Zt,bt.width,bt.height,bt.depth,0,wt,qt,bt.data);else if(y.isData3DTexture)k?(Rt&&h.texStorage3D(o.TEXTURE_3D,Xt,Zt,bt.width,bt.height,bt.depth),Ct&&h.texSubImage3D(o.TEXTURE_3D,0,0,0,0,bt.width,bt.height,bt.depth,wt,qt,bt.data)):h.texImage3D(o.TEXTURE_3D,0,Zt,bt.width,bt.height,bt.depth,0,wt,qt,bt.data);else if(y.isFramebufferTexture){if(Rt)if(k)h.texStorage2D(o.TEXTURE_2D,Xt,Zt,bt.width,bt.height);else{let Et=bt.width,ft=bt.height;for(let $t=0;$t<Xt;$t++)h.texImage2D(o.TEXTURE_2D,$t,Zt,Et,ft,0,wt,qt,null),Et>>=1,ft>>=1}}else if(Se.length>0){if(k&&Rt){const Et=Jt(Se[0]);h.texStorage2D(o.TEXTURE_2D,Xt,Zt,Et.width,Et.height)}for(let Et=0,ft=Se.length;Et<ft;Et++)Ht=Se[Et],k?Ct&&h.texSubImage2D(o.TEXTURE_2D,Et,0,0,wt,qt,Ht):h.texImage2D(o.TEXTURE_2D,Et,Zt,wt,qt,Ht);y.generateMipmaps=!1}else if(k){if(Rt){const Et=Jt(bt);h.texStorage2D(o.TEXTURE_2D,Xt,Zt,Et.width,Et.height)}Ct&&h.texSubImage2D(o.TEXTURE_2D,0,0,0,wt,qt,bt)}else h.texImage2D(o.TEXTURE_2D,0,Zt,wt,qt,bt);A(y)&&b(ut),Ft.__version=pt.version,y.onUpdate&&y.onUpdate(y)}P.__version=y.version}function Lt(P,y,W){if(y.image.length!==6)return;const ut=se(P,y),yt=y.source;h.bindTexture(o.TEXTURE_CUBE_MAP,P.__webglTexture,o.TEXTURE0+W);const pt=x.get(yt);if(yt.version!==pt.__version||ut===!0){h.activeTexture(o.TEXTURE0+W);const Ft=f.ppV.getPrimaries(f.ppV.workingColorSpace),Pt=y.colorSpace===f.jf0?null:f.ppV.getPrimaries(y.colorSpace),ne=y.colorSpace===f.jf0||Ft===Pt?o.NONE:o.BROWSER_DEFAULT_WEBGL;o.pixelStorei(o.UNPACK_FLIP_Y_WEBGL,y.flipY),o.pixelStorei(o.UNPACK_PREMULTIPLY_ALPHA_WEBGL,y.premultiplyAlpha),o.pixelStorei(o.UNPACK_ALIGNMENT,y.unpackAlignment),o.pixelStorei(o.UNPACK_COLORSPACE_CONVERSION_WEBGL,ne);const ce=y.isCompressedTexture||y.image[0].isCompressedTexture,bt=y.image[0]&&y.image[0].isDataTexture,wt=[];for(let ft=0;ft<6;ft++)!ce&&!bt?wt[ft]=nt(y.image[ft],!0,E.maxCubemapSize):wt[ft]=bt?y.image[ft].image:y.image[ft],wt[ft]=Ee(y,wt[ft]);const qt=wt[0],Zt=S.convert(y.format,y.colorSpace),Ht=S.convert(y.type),Se=q(y.internalFormat,Zt,Ht,y.colorSpace),k=y.isVideoTexture!==!0,Rt=pt.__version===void 0||ut===!0,Ct=yt.dataReady;let Xt=ot(y,qt);te(o.TEXTURE_CUBE_MAP,y);let Et;if(ce){k&&Rt&&h.texStorage2D(o.TEXTURE_CUBE_MAP,Xt,Se,qt.width,qt.height);for(let ft=0;ft<6;ft++){Et=wt[ft].mipmaps;for(let $t=0;$t<Et.length;$t++){const me=Et[$t];y.format!==f.GWd?Zt!==null?k?Ct&&h.compressedTexSubImage2D(o.TEXTURE_CUBE_MAP_POSITIVE_X+ft,$t,0,0,me.width,me.height,Zt,me.data):h.compressedTexImage2D(o.TEXTURE_CUBE_MAP_POSITIVE_X+ft,$t,Se,me.width,me.height,0,me.data):(0,f.R8M)("WebGLRenderer: Attempt to load unsupported compressed texture format in .setTextureCube()"):k?Ct&&h.texSubImage2D(o.TEXTURE_CUBE_MAP_POSITIVE_X+ft,$t,0,0,me.width,me.height,Zt,Ht,me.data):h.texImage2D(o.TEXTURE_CUBE_MAP_POSITIVE_X+ft,$t,Se,me.width,me.height,0,Zt,Ht,me.data)}}}else{if(Et=y.mipmaps,k&&Rt){Et.length>0&&Xt++;const ft=Jt(wt[0]);h.texStorage2D(o.TEXTURE_CUBE_MAP,Xt,Se,ft.width,ft.height)}for(let ft=0;ft<6;ft++)if(bt){k?Ct&&h.texSubImage2D(o.TEXTURE_CUBE_MAP_POSITIVE_X+ft,0,0,0,wt[ft].width,wt[ft].height,Zt,Ht,wt[ft].data):h.texImage2D(o.TEXTURE_CUBE_MAP_POSITIVE_X+ft,0,Se,wt[ft].width,wt[ft].height,0,Zt,Ht,wt[ft].data);for(let $t=0;$t<Et.length;$t++){const Ge=Et[$t].image[ft].image;k?Ct&&h.texSubImage2D(o.TEXTURE_CUBE_MAP_POSITIVE_X+ft,$t+1,0,0,Ge.width,Ge.height,Zt,Ht,Ge.data):h.texImage2D(o.TEXTURE_CUBE_MAP_POSITIVE_X+ft,$t+1,Se,Ge.width,Ge.height,0,Zt,Ht,Ge.data)}}else{k?Ct&&h.texSubImage2D(o.TEXTURE_CUBE_MAP_POSITIVE_X+ft,0,0,0,Zt,Ht,wt[ft]):h.texImage2D(o.TEXTURE_CUBE_MAP_POSITIVE_X+ft,0,Se,Zt,Ht,wt[ft]);for(let $t=0;$t<Et.length;$t++){const me=Et[$t];k?Ct&&h.texSubImage2D(o.TEXTURE_CUBE_MAP_POSITIVE_X+ft,$t+1,0,0,Zt,Ht,me.image[ft]):h.texImage2D(o.TEXTURE_CUBE_MAP_POSITIVE_X+ft,$t+1,Se,Zt,Ht,me.image[ft])}}}A(y)&&b(o.TEXTURE_CUBE_MAP),pt.__version=yt.version,y.onUpdate&&y.onUpdate(y)}P.__version=y.version}function Dt(P,y,W,ut,yt,pt){const Ft=S.convert(W.format,W.colorSpace),Pt=S.convert(W.type),ne=q(W.internalFormat,Ft,Pt,W.colorSpace),ce=x.get(y),bt=x.get(W);if(bt.__renderTarget=y,!ce.__hasExternalTextures){const wt=Math.max(1,y.width>>pt),qt=Math.max(1,y.height>>pt);yt===o.TEXTURE_3D||yt===o.TEXTURE_2D_ARRAY?h.texImage3D(yt,pt,ne,wt,qt,y.depth,0,Ft,Pt,null):h.texImage2D(yt,pt,ne,wt,qt,0,Ft,Pt,null)}h.bindFramebuffer(o.FRAMEBUFFER,P),Xe(y)?V.framebufferTexture2DMultisampleEXT(o.FRAMEBUFFER,ut,yt,bt.__webglTexture,0,B(y)):(yt===o.TEXTURE_2D||yt>=o.TEXTURE_CUBE_MAP_POSITIVE_X&&yt<=o.TEXTURE_CUBE_MAP_NEGATIVE_Z)&&o.framebufferTexture2D(o.FRAMEBUFFER,ut,yt,bt.__webglTexture,pt),h.bindFramebuffer(o.FRAMEBUFFER,null)}function Te(P,y,W){if(o.bindRenderbuffer(o.RENDERBUFFER,P),y.depthBuffer){const ut=y.depthTexture,yt=ut&&ut.isDepthTexture?ut.type:null,pt=H(y.stencilBuffer,yt),Ft=y.stencilBuffer?o.DEPTH_STENCIL_ATTACHMENT:o.DEPTH_ATTACHMENT;Xe(y)?V.renderbufferStorageMultisampleEXT(o.RENDERBUFFER,B(y),pt,y.width,y.height):W?o.renderbufferStorageMultisample(o.RENDERBUFFER,B(y),pt,y.width,y.height):o.renderbufferStorage(o.RENDERBUFFER,pt,y.width,y.height),o.framebufferRenderbuffer(o.FRAMEBUFFER,Ft,o.RENDERBUFFER,P)}else{const ut=y.textures;for(let yt=0;yt<ut.length;yt++){const pt=ut[yt],Ft=S.convert(pt.format,pt.colorSpace),Pt=S.convert(pt.type),ne=q(pt.internalFormat,Ft,Pt,pt.colorSpace);Xe(y)?V.renderbufferStorageMultisampleEXT(o.RENDERBUFFER,B(y),ne,y.width,y.height):W?o.renderbufferStorageMultisample(o.RENDERBUFFER,B(y),ne,y.width,y.height):o.renderbufferStorage(o.RENDERBUFFER,ne,y.width,y.height)}}o.bindRenderbuffer(o.RENDERBUFFER,null)}function de(P,y,W){const ut=y.isWebGLCubeRenderTarget===!0;if(h.bindFramebuffer(o.FRAMEBUFFER,P),!(y.depthTexture&&y.depthTexture.isDepthTexture))throw new Error("renderTarget.depthTexture must be an instance of THREE.DepthTexture");const yt=x.get(y.depthTexture);if(yt.__renderTarget=y,(!yt.__webglTexture||y.depthTexture.image.width!==y.width||y.depthTexture.image.height!==y.height)&&(y.depthTexture.image.width=y.width,y.depthTexture.image.height=y.height,y.depthTexture.needsUpdate=!0),ut){if(yt.__webglInit===void 0&&(yt.__webglInit=!0,y.depthTexture.addEventListener("dispose",J)),yt.__webglTexture===void 0){yt.__webglTexture=o.createTexture(),h.bindTexture(o.TEXTURE_CUBE_MAP,yt.__webglTexture),te(o.TEXTURE_CUBE_MAP,y.depthTexture);const ce=S.convert(y.depthTexture.format),bt=S.convert(y.depthTexture.type);let wt;y.depthTexture.format===f.zdS?wt=o.DEPTH_COMPONENT24:y.depthTexture.format===f.dcC&&(wt=o.DEPTH24_STENCIL8);for(let qt=0;qt<6;qt++)o.texImage2D(o.TEXTURE_CUBE_MAP_POSITIVE_X+qt,0,wt,y.width,y.height,0,ce,bt,null)}}else _t(y.depthTexture,0);const pt=yt.__webglTexture,Ft=B(y),Pt=ut?o.TEXTURE_CUBE_MAP_POSITIVE_X+W:o.TEXTURE_2D,ne=y.depthTexture.format===f.dcC?o.DEPTH_STENCIL_ATTACHMENT:o.DEPTH_ATTACHMENT;if(y.depthTexture.format===f.zdS)Xe(y)?V.framebufferTexture2DMultisampleEXT(o.FRAMEBUFFER,ne,Pt,pt,0,Ft):o.framebufferTexture2D(o.FRAMEBUFFER,ne,Pt,pt,0);else if(y.depthTexture.format===f.dcC)Xe(y)?V.framebufferTexture2DMultisampleEXT(o.FRAMEBUFFER,ne,Pt,pt,0,Ft):o.framebufferTexture2D(o.FRAMEBUFFER,ne,Pt,pt,0);else throw new Error("Unknown depthTexture format")}function _e(P){const y=x.get(P),W=P.isWebGLCubeRenderTarget===!0;if(y.__boundDepthTexture!==P.depthTexture){const ut=P.depthTexture;if(y.__depthDisposeCallback&&y.__depthDisposeCallback(),ut){const yt=()=>{delete y.__boundDepthTexture,delete y.__depthDisposeCallback,ut.removeEventListener("dispose",yt)};ut.addEventListener("dispose",yt),y.__depthDisposeCallback=yt}y.__boundDepthTexture=ut}if(P.depthTexture&&!y.__autoAllocateDepthBuffer)if(W)for(let ut=0;ut<6;ut++)de(y.__webglFramebuffer[ut],P,ut);else{const ut=P.texture.mipmaps;ut&&ut.length>0?de(y.__webglFramebuffer[0],P,0):de(y.__webglFramebuffer,P,0)}else if(W){y.__webglDepthbuffer=[];for(let ut=0;ut<6;ut++)if(h.bindFramebuffer(o.FRAMEBUFFER,y.__webglFramebuffer[ut]),y.__webglDepthbuffer[ut]===void 0)y.__webglDepthbuffer[ut]=o.createRenderbuffer(),Te(y.__webglDepthbuffer[ut],P,!1);else{const yt=P.stencilBuffer?o.DEPTH_STENCIL_ATTACHMENT:o.DEPTH_ATTACHMENT,pt=y.__webglDepthbuffer[ut];o.bindRenderbuffer(o.RENDERBUFFER,pt),o.framebufferRenderbuffer(o.FRAMEBUFFER,yt,o.RENDERBUFFER,pt)}}else{const ut=P.texture.mipmaps;if(ut&&ut.length>0?h.bindFramebuffer(o.FRAMEBUFFER,y.__webglFramebuffer[0]):h.bindFramebuffer(o.FRAMEBUFFER,y.__webglFramebuffer),y.__webglDepthbuffer===void 0)y.__webglDepthbuffer=o.createRenderbuffer(),Te(y.__webglDepthbuffer,P,!1);else{const yt=P.stencilBuffer?o.DEPTH_STENCIL_ATTACHMENT:o.DEPTH_ATTACHMENT,pt=y.__webglDepthbuffer;o.bindRenderbuffer(o.RENDERBUFFER,pt),o.framebufferRenderbuffer(o.FRAMEBUFFER,yt,o.RENDERBUFFER,pt)}}h.bindFramebuffer(o.FRAMEBUFFER,null)}function Ut(P,y,W){const ut=x.get(P);y!==void 0&&Dt(ut.__webglFramebuffer,P,P.texture,o.COLOR_ATTACHMENT0,o.TEXTURE_2D,0),W!==void 0&&_e(P)}function Ie(P){const y=P.texture,W=x.get(P),ut=x.get(y);P.addEventListener("dispose",st);const yt=P.textures,pt=P.isWebGLCubeRenderTarget===!0,Ft=yt.length>1;if(Ft||(ut.__webglTexture===void 0&&(ut.__webglTexture=o.createTexture()),ut.__version=y.version,I.memory.textures++),pt){W.__webglFramebuffer=[];for(let Pt=0;Pt<6;Pt++)if(y.mipmaps&&y.mipmaps.length>0){W.__webglFramebuffer[Pt]=[];for(let ne=0;ne<y.mipmaps.length;ne++)W.__webglFramebuffer[Pt][ne]=o.createFramebuffer()}else W.__webglFramebuffer[Pt]=o.createFramebuffer()}else{if(y.mipmaps&&y.mipmaps.length>0){W.__webglFramebuffer=[];for(let Pt=0;Pt<y.mipmaps.length;Pt++)W.__webglFramebuffer[Pt]=o.createFramebuffer()}else W.__webglFramebuffer=o.createFramebuffer();if(Ft)for(let Pt=0,ne=yt.length;Pt<ne;Pt++){const ce=x.get(yt[Pt]);ce.__webglTexture===void 0&&(ce.__webglTexture=o.createTexture(),I.memory.textures++)}if(P.samples>0&&Xe(P)===!1){W.__webglMultisampledFramebuffer=o.createFramebuffer(),W.__webglColorRenderbuffer=[],h.bindFramebuffer(o.FRAMEBUFFER,W.__webglMultisampledFramebuffer);for(let Pt=0;Pt<yt.length;Pt++){const ne=yt[Pt];W.__webglColorRenderbuffer[Pt]=o.createRenderbuffer(),o.bindRenderbuffer(o.RENDERBUFFER,W.__webglColorRenderbuffer[Pt]);const ce=S.convert(ne.format,ne.colorSpace),bt=S.convert(ne.type),wt=q(ne.internalFormat,ce,bt,ne.colorSpace,P.isXRRenderTarget===!0),qt=B(P);o.renderbufferStorageMultisample(o.RENDERBUFFER,qt,wt,P.width,P.height),o.framebufferRenderbuffer(o.FRAMEBUFFER,o.COLOR_ATTACHMENT0+Pt,o.RENDERBUFFER,W.__webglColorRenderbuffer[Pt])}o.bindRenderbuffer(o.RENDERBUFFER,null),P.depthBuffer&&(W.__webglDepthRenderbuffer=o.createRenderbuffer(),Te(W.__webglDepthRenderbuffer,P,!0)),h.bindFramebuffer(o.FRAMEBUFFER,null)}}if(pt){h.bindTexture(o.TEXTURE_CUBE_MAP,ut.__webglTexture),te(o.TEXTURE_CUBE_MAP,y);for(let Pt=0;Pt<6;Pt++)if(y.mipmaps&&y.mipmaps.length>0)for(let ne=0;ne<y.mipmaps.length;ne++)Dt(W.__webglFramebuffer[Pt][ne],P,y,o.COLOR_ATTACHMENT0,o.TEXTURE_CUBE_MAP_POSITIVE_X+Pt,ne);else Dt(W.__webglFramebuffer[Pt],P,y,o.COLOR_ATTACHMENT0,o.TEXTURE_CUBE_MAP_POSITIVE_X+Pt,0);A(y)&&b(o.TEXTURE_CUBE_MAP),h.unbindTexture()}else if(Ft){for(let Pt=0,ne=yt.length;Pt<ne;Pt++){const ce=yt[Pt],bt=x.get(ce);let wt=o.TEXTURE_2D;(P.isWebGL3DRenderTarget||P.isWebGLArrayRenderTarget)&&(wt=P.isWebGL3DRenderTarget?o.TEXTURE_3D:o.TEXTURE_2D_ARRAY),h.bindTexture(wt,bt.__webglTexture),te(wt,ce),Dt(W.__webglFramebuffer,P,ce,o.COLOR_ATTACHMENT0+Pt,wt,0),A(ce)&&b(wt)}h.unbindTexture()}else{let Pt=o.TEXTURE_2D;if((P.isWebGL3DRenderTarget||P.isWebGLArrayRenderTarget)&&(Pt=P.isWebGL3DRenderTarget?o.TEXTURE_3D:o.TEXTURE_2D_ARRAY),h.bindTexture(Pt,ut.__webglTexture),te(Pt,y),y.mipmaps&&y.mipmaps.length>0)for(let ne=0;ne<y.mipmaps.length;ne++)Dt(W.__webglFramebuffer[ne],P,y,o.COLOR_ATTACHMENT0,Pt,ne);else Dt(W.__webglFramebuffer,P,y,o.COLOR_ATTACHMENT0,Pt,0);A(y)&&b(Pt),h.unbindTexture()}P.depthBuffer&&_e(P)}function Pe(P){const y=P.textures;for(let W=0,ut=y.length;W<ut;W++){const yt=y[W];if(A(yt)){const pt=z(P),Ft=x.get(yt).__webglTexture;h.bindTexture(pt,Ft),b(pt),h.unbindTexture()}}}const Be=[],ae=[];function Je(P){if(P.samples>0){if(Xe(P)===!1){const y=P.textures,W=P.width,ut=P.height;let yt=o.COLOR_BUFFER_BIT;const pt=P.stencilBuffer?o.DEPTH_STENCIL_ATTACHMENT:o.DEPTH_ATTACHMENT,Ft=x.get(P),Pt=y.length>1;if(Pt)for(let ce=0;ce<y.length;ce++)h.bindFramebuffer(o.FRAMEBUFFER,Ft.__webglMultisampledFramebuffer),o.framebufferRenderbuffer(o.FRAMEBUFFER,o.COLOR_ATTACHMENT0+ce,o.RENDERBUFFER,null),h.bindFramebuffer(o.FRAMEBUFFER,Ft.__webglFramebuffer),o.framebufferTexture2D(o.DRAW_FRAMEBUFFER,o.COLOR_ATTACHMENT0+ce,o.TEXTURE_2D,null,0);h.bindFramebuffer(o.READ_FRAMEBUFFER,Ft.__webglMultisampledFramebuffer);const ne=P.texture.mipmaps;ne&&ne.length>0?h.bindFramebuffer(o.DRAW_FRAMEBUFFER,Ft.__webglFramebuffer[0]):h.bindFramebuffer(o.DRAW_FRAMEBUFFER,Ft.__webglFramebuffer);for(let ce=0;ce<y.length;ce++){if(P.resolveDepthBuffer&&(P.depthBuffer&&(yt|=o.DEPTH_BUFFER_BIT),P.stencilBuffer&&P.resolveStencilBuffer&&(yt|=o.STENCIL_BUFFER_BIT)),Pt){o.framebufferRenderbuffer(o.READ_FRAMEBUFFER,o.COLOR_ATTACHMENT0,o.RENDERBUFFER,Ft.__webglColorRenderbuffer[ce]);const bt=x.get(y[ce]).__webglTexture;o.framebufferTexture2D(o.DRAW_FRAMEBUFFER,o.COLOR_ATTACHMENT0,o.TEXTURE_2D,bt,0)}o.blitFramebuffer(0,0,W,ut,0,0,W,ut,yt,o.NEAREST),j===!0&&(Be.length=0,ae.length=0,Be.push(o.COLOR_ATTACHMENT0+ce),P.depthBuffer&&P.resolveDepthBuffer===!1&&(Be.push(pt),ae.push(pt),o.invalidateFramebuffer(o.DRAW_FRAMEBUFFER,ae)),o.invalidateFramebuffer(o.READ_FRAMEBUFFER,Be))}if(h.bindFramebuffer(o.READ_FRAMEBUFFER,null),h.bindFramebuffer(o.DRAW_FRAMEBUFFER,null),Pt)for(let ce=0;ce<y.length;ce++){h.bindFramebuffer(o.FRAMEBUFFER,Ft.__webglMultisampledFramebuffer),o.framebufferRenderbuffer(o.FRAMEBUFFER,o.COLOR_ATTACHMENT0+ce,o.RENDERBUFFER,Ft.__webglColorRenderbuffer[ce]);const bt=x.get(y[ce]).__webglTexture;h.bindFramebuffer(o.FRAMEBUFFER,Ft.__webglFramebuffer),o.framebufferTexture2D(o.DRAW_FRAMEBUFFER,o.COLOR_ATTACHMENT0+ce,o.TEXTURE_2D,bt,0)}h.bindFramebuffer(o.DRAW_FRAMEBUFFER,Ft.__webglMultisampledFramebuffer)}else if(P.depthBuffer&&P.resolveDepthBuffer===!1&&j){const y=P.stencilBuffer?o.DEPTH_STENCIL_ATTACHMENT:o.DEPTH_ATTACHMENT;o.invalidateFramebuffer(o.DRAW_FRAMEBUFFER,[y])}}}function B(P){return Math.min(E.maxSamples,P.samples)}function Xe(P){const y=x.get(P);return P.samples>0&&p.has("WEBGL_multisampled_render_to_texture")===!0&&y.__useRenderToTexture!==!1}function Ce(P){const y=I.render.frame;ht.get(P)!==y&&(ht.set(P,y),P.update())}function Ee(P,y){const W=P.colorSpace,ut=P.format,yt=P.type;return P.isCompressedTexture===!0||P.isVideoTexture===!0||W!==f.Zr2&&W!==f.jf0&&(f.ppV.getTransfer(W)===f.KLL?(ut!==f.GWd||yt!==f.OUM)&&(0,f.R8M)("WebGLTextures: sRGB encoded textures have to use RGBAFormat and UnsignedByteType."):(0,f.z3S)("WebGLTextures: Unsupported texture color space:",W)),y}function Jt(P){return typeof HTMLImageElement<"u"&&P instanceof HTMLImageElement?($.width=P.naturalWidth||P.width,$.height=P.naturalHeight||P.height):typeof VideoFrame<"u"&&P instanceof VideoFrame?($.width=P.displayWidth,$.height=P.displayHeight):($.width=P.width,$.height=P.height),$}this.allocateTextureUnit=mt,this.resetTextureUnits=dt,this.setTexture2D=_t,this.setTexture2DArray=ct,this.setTexture3D=it,this.setTextureCube=It,this.rebindTextures=Ut,this.setupRenderTarget=Ie,this.updateRenderTargetMipmap=Pe,this.updateMultisampleRenderTarget=Je,this.setupDepthRenderbuffer=_e,this.setupFrameBufferTexture=Dt,this.useMultisampledRTT=Xe,this.isReversedDepthBuffer=function(){return h.buffers.depth.getReversed()}}function zi(o,p){function h(x,E=f.jf0){let S;const I=f.ppV.getTransfer(E);if(x===f.OUM)return o.UNSIGNED_BYTE;if(x===f.Wew)return o.UNSIGNED_SHORT_4_4_4_4;if(x===f.gJ2)return o.UNSIGNED_SHORT_5_5_5_1;if(x===f.Dmk)return o.UNSIGNED_INT_5_9_9_9_REV;if(x===f.yT7)return o.UNSIGNED_INT_10F_11F_11F_REV;if(x===f.tJf)return o.BYTE;if(x===f.fBL)return o.SHORT;if(x===f.cHt)return o.UNSIGNED_SHORT;if(x===f.Yuy)return o.INT;if(x===f.bkx)return o.UNSIGNED_INT;if(x===f.RQf)return o.FLOAT;if(x===f.ix0)return o.HALF_FLOAT;if(x===f.wrO)return o.ALPHA;if(x===f.HIg)return o.RGB;if(x===f.GWd)return o.RGBA;if(x===f.zdS)return o.DEPTH_COMPONENT;if(x===f.dcC)return o.DEPTH_STENCIL;if(x===f.VT0)return o.RED;if(x===f.ZQM)return o.RED_INTEGER;if(x===f.paN)return o.RG;if(x===f.TkQ)return o.RG_INTEGER;if(x===f.c90)return o.RGBA_INTEGER;if(x===f.IE4||x===f.Nz6||x===f.jR7||x===f.BXX)if(I===f.KLL)if(S=p.get("WEBGL_compressed_texture_s3tc_srgb"),S!==null){if(x===f.IE4)return S.COMPRESSED_SRGB_S3TC_DXT1_EXT;if(x===f.Nz6)return S.COMPRESSED_SRGB_ALPHA_S3TC_DXT1_EXT;if(x===f.jR7)return S.COMPRESSED_SRGB_ALPHA_S3TC_DXT3_EXT;if(x===f.BXX)return S.COMPRESSED_SRGB_ALPHA_S3TC_DXT5_EXT}else return null;else if(S=p.get("WEBGL_compressed_texture_s3tc"),S!==null){if(x===f.IE4)return S.COMPRESSED_RGB_S3TC_DXT1_EXT;if(x===f.Nz6)return S.COMPRESSED_RGBA_S3TC_DXT1_EXT;if(x===f.jR7)return S.COMPRESSED_RGBA_S3TC_DXT3_EXT;if(x===f.BXX)return S.COMPRESSED_RGBA_S3TC_DXT5_EXT}else return null;if(x===f.k6Q||x===f.kTp||x===f.HXV||x===f.pBf)if(S=p.get("WEBGL_compressed_texture_pvrtc"),S!==null){if(x===f.k6Q)return S.COMPRESSED_RGB_PVRTC_4BPPV1_IMG;if(x===f.kTp)return S.COMPRESSED_RGB_PVRTC_2BPPV1_IMG;if(x===f.HXV)return S.COMPRESSED_RGBA_PVRTC_4BPPV1_IMG;if(x===f.pBf)return S.COMPRESSED_RGBA_PVRTC_2BPPV1_IMG}else return null;if(x===f.CVz||x===f.Riy||x===f.KDk||x===f.BVL||x===f.gZr||x===f.OtU||x===f.jSS)if(S=p.get("WEBGL_compressed_texture_etc"),S!==null){if(x===f.CVz||x===f.Riy)return I===f.KLL?S.COMPRESSED_SRGB8_ETC2:S.COMPRESSED_RGB8_ETC2;if(x===f.KDk)return I===f.KLL?S.COMPRESSED_SRGB8_ALPHA8_ETC2_EAC:S.COMPRESSED_RGBA8_ETC2_EAC;if(x===f.BVL)return S.COMPRESSED_R11_EAC;if(x===f.gZr)return S.COMPRESSED_SIGNED_R11_EAC;if(x===f.OtU)return S.COMPRESSED_RG11_EAC;if(x===f.jSS)return S.COMPRESSED_SIGNED_RG11_EAC}else return null;if(x===f.qa3||x===f.B_h||x===f.czI||x===f.rSH||x===f.Qrf||x===f.psI||x===f.a5J||x===f._QJ||x===f.uB5||x===f.lyL||x===f.bC7||x===f.y3Z||x===f.ojs||x===f.S$4)if(S=p.get("WEBGL_compressed_texture_astc"),S!==null){if(x===f.qa3)return I===f.KLL?S.COMPRESSED_SRGB8_ALPHA8_ASTC_4x4_KHR:S.COMPRESSED_RGBA_ASTC_4x4_KHR;if(x===f.B_h)return I===f.KLL?S.COMPRESSED_SRGB8_ALPHA8_ASTC_5x4_KHR:S.COMPRESSED_RGBA_ASTC_5x4_KHR;if(x===f.czI)return I===f.KLL?S.COMPRESSED_SRGB8_ALPHA8_ASTC_5x5_KHR:S.COMPRESSED_RGBA_ASTC_5x5_KHR;if(x===f.rSH)return I===f.KLL?S.COMPRESSED_SRGB8_ALPHA8_ASTC_6x5_KHR:S.COMPRESSED_RGBA_ASTC_6x5_KHR;if(x===f.Qrf)return I===f.KLL?S.COMPRESSED_SRGB8_ALPHA8_ASTC_6x6_KHR:S.COMPRESSED_RGBA_ASTC_6x6_KHR;if(x===f.psI)return I===f.KLL?S.COMPRESSED_SRGB8_ALPHA8_ASTC_8x5_KHR:S.COMPRESSED_RGBA_ASTC_8x5_KHR;if(x===f.a5J)return I===f.KLL?S.COMPRESSED_SRGB8_ALPHA8_ASTC_8x6_KHR:S.COMPRESSED_RGBA_ASTC_8x6_KHR;if(x===f._QJ)return I===f.KLL?S.COMPRESSED_SRGB8_ALPHA8_ASTC_8x8_KHR:S.COMPRESSED_RGBA_ASTC_8x8_KHR;if(x===f.uB5)return I===f.KLL?S.COMPRESSED_SRGB8_ALPHA8_ASTC_10x5_KHR:S.COMPRESSED_RGBA_ASTC_10x5_KHR;if(x===f.lyL)return I===f.KLL?S.COMPRESSED_SRGB8_ALPHA8_ASTC_10x6_KHR:S.COMPRESSED_RGBA_ASTC_10x6_KHR;if(x===f.bC7)return I===f.KLL?S.COMPRESSED_SRGB8_ALPHA8_ASTC_10x8_KHR:S.COMPRESSED_RGBA_ASTC_10x8_KHR;if(x===f.y3Z)return I===f.KLL?S.COMPRESSED_SRGB8_ALPHA8_ASTC_10x10_KHR:S.COMPRESSED_RGBA_ASTC_10x10_KHR;if(x===f.ojs)return I===f.KLL?S.COMPRESSED_SRGB8_ALPHA8_ASTC_12x10_KHR:S.COMPRESSED_RGBA_ASTC_12x10_KHR;if(x===f.S$4)return I===f.KLL?S.COMPRESSED_SRGB8_ALPHA8_ASTC_12x12_KHR:S.COMPRESSED_RGBA_ASTC_12x12_KHR}else return null;if(x===f.Fn||x===f.H23||x===f.W9U)if(S=p.get("EXT_texture_compression_bptc"),S!==null){if(x===f.Fn)return I===f.KLL?S.COMPRESSED_SRGB_ALPHA_BPTC_UNORM_EXT:S.COMPRESSED_RGBA_BPTC_UNORM_EXT;if(x===f.H23)return S.COMPRESSED_RGB_BPTC_SIGNED_FLOAT_EXT;if(x===f.W9U)return S.COMPRESSED_RGB_BPTC_UNSIGNED_FLOAT_EXT}else return null;if(x===f.Kef||x===f.XG_||x===f.HO_||x===f.CWW)if(S=p.get("EXT_texture_compression_rgtc"),S!==null){if(x===f.Kef)return S.COMPRESSED_RED_RGTC1_EXT;if(x===f.XG_)return S.COMPRESSED_SIGNED_RED_RGTC1_EXT;if(x===f.HO_)return S.COMPRESSED_RED_GREEN_RGTC2_EXT;if(x===f.CWW)return S.COMPRESSED_SIGNED_RED_GREEN_RGTC2_EXT}else return null;return x===f.V3x?o.UNSIGNED_INT_24_8:o[x]!==void 0?o[x]:null}return{convert:h}}const Rr=`
void main() {

	gl_Position = vec4( position, 1.0 );

}`,De=`
uniform sampler2DArray depthColor;
uniform float depthWidth;
uniform float depthHeight;

void main() {

	vec2 coord = vec2( gl_FragCoord.x / depthWidth, gl_FragCoord.y / depthHeight );

	if ( coord.x >= 1.0 ) {

		gl_FragDepth = texture( depthColor, vec3( coord.x - 1.0, coord.y, 1 ) ).r;

	} else {

		gl_FragDepth = texture( depthColor, vec3( coord.x, coord.y, 0 ) ).r;

	}

}`;class rs{constructor(){this.texture=null,this.mesh=null,this.depthNear=0,this.depthFar=0}init(p,h){if(this.texture===null){const x=new f.rjZ(p.texture);(p.depthNear!==h.depthNear||p.depthFar!==h.depthFar)&&(this.depthNear=p.depthNear,this.depthFar=p.depthFar),this.texture=x}}getMesh(p){if(this.texture!==null&&this.mesh===null){const h=p.cameras[0].viewport,x=new f.BKk({vertexShader:Rr,fragmentShader:De,uniforms:{depthColor:{value:this.texture},depthWidth:{value:h.z},depthHeight:{value:h.w}}});this.mesh=new f.eaF(new f.bdM(20,20),x)}return this.mesh}reset(){this.texture=null,this.mesh=null}getDepthTexture(){return this.texture}}class Dc extends f.Qev{constructor(p,h){super();const x=this;let E=null,S=1,I=null,V="local-floor",j=1,$=null,ht=null,K=null,F=null,G=null,X=null;const nt=typeof XRWebGLBinding<"u",A=new rs,b={},z=h.getContextAttributes();let q=null,H=null;const ot=[],J=[],st=new f.I9Y;let R=null;const U=new f.ubm;U.viewport=new f.IUQ;const Tt=new f.ubm;Tt.viewport=new f.IUQ;const Z=[U,Tt],dt=new f.nZQ;let mt=null,gt=null;this.cameraAutoUpdate=!0,this.enabled=!1,this.isPresenting=!1,this.getController=function(xt){let Lt=ot[xt];return Lt===void 0&&(Lt=new f.R3r,ot[xt]=Lt),Lt.getTargetRaySpace()},this.getControllerGrip=function(xt){let Lt=ot[xt];return Lt===void 0&&(Lt=new f.R3r,ot[xt]=Lt),Lt.getGripSpace()},this.getHand=function(xt){let Lt=ot[xt];return Lt===void 0&&(Lt=new f.R3r,ot[xt]=Lt),Lt.getHandSpace()};function _t(xt){const Lt=J.indexOf(xt.inputSource);if(Lt===-1)return;const Dt=ot[Lt];Dt!==void 0&&(Dt.update(xt.inputSource,xt.frame,$||I),Dt.dispatchEvent({type:xt.type,data:xt.inputSource}))}function ct(){E.removeEventListener("select",_t),E.removeEventListener("selectstart",_t),E.removeEventListener("selectend",_t),E.removeEventListener("squeeze",_t),E.removeEventListener("squeezestart",_t),E.removeEventListener("squeezeend",_t),E.removeEventListener("end",ct),E.removeEventListener("inputsourceschange",it);for(let xt=0;xt<ot.length;xt++){const Lt=J[xt];Lt!==null&&(J[xt]=null,ot[xt].disconnect(Lt))}mt=null,gt=null,A.reset();for(const xt in b)delete b[xt];p.setRenderTarget(q),G=null,F=null,K=null,E=null,H=null,$e.stop(),x.isPresenting=!1,p.setPixelRatio(R),p.setSize(st.width,st.height,!1),x.dispatchEvent({type:"sessionend"})}this.setFramebufferScaleFactor=function(xt){S=xt,x.isPresenting===!0&&(0,f.R8M)("WebXRManager: Cannot change framebuffer scale while presenting.")},this.setReferenceSpaceType=function(xt){V=xt,x.isPresenting===!0&&(0,f.R8M)("WebXRManager: Cannot change reference space type while presenting.")},this.getReferenceSpace=function(){return $||I},this.setReferenceSpace=function(xt){$=xt},this.getBaseLayer=function(){return F!==null?F:G},this.getBinding=function(){return K===null&&nt&&(K=new XRWebGLBinding(E,h)),K},this.getFrame=function(){return X},this.getSession=function(){return E},this.setSession=async function(xt){if(E=xt,E!==null){if(q=p.getRenderTarget(),E.addEventListener("select",_t),E.addEventListener("selectstart",_t),E.addEventListener("selectend",_t),E.addEventListener("squeeze",_t),E.addEventListener("squeezestart",_t),E.addEventListener("squeezeend",_t),E.addEventListener("end",ct),E.addEventListener("inputsourceschange",it),z.xrCompatible!==!0&&await h.makeXRCompatible(),R=p.getPixelRatio(),p.getSize(st),nt&&"createProjectionLayer"in XRWebGLBinding.prototype){let Dt=null,Te=null,de=null;z.depth&&(de=z.stencil?h.DEPTH24_STENCIL8:h.DEPTH_COMPONENT24,Dt=z.stencil?f.dcC:f.zdS,Te=z.stencil?f.V3x:f.bkx);const _e={colorFormat:h.RGBA8,depthFormat:de,scaleFactor:S};K=this.getBinding(),F=K.createProjectionLayer(_e),E.updateRenderState({layers:[F]}),p.setPixelRatio(1),p.setSize(F.textureWidth,F.textureHeight,!1),H=new f.nWS(F.textureWidth,F.textureHeight,{format:f.GWd,type:f.OUM,depthTexture:new f.VCu(F.textureWidth,F.textureHeight,Te,void 0,void 0,void 0,void 0,void 0,void 0,Dt),stencilBuffer:z.stencil,colorSpace:p.outputColorSpace,samples:z.antialias?4:0,resolveDepthBuffer:F.ignoreDepthValues===!1,resolveStencilBuffer:F.ignoreDepthValues===!1})}else{const Dt={antialias:z.antialias,alpha:!0,depth:z.depth,stencil:z.stencil,framebufferScaleFactor:S};G=new XRWebGLLayer(E,h,Dt),E.updateRenderState({baseLayer:G}),p.setPixelRatio(1),p.setSize(G.framebufferWidth,G.framebufferHeight,!1),H=new f.nWS(G.framebufferWidth,G.framebufferHeight,{format:f.GWd,type:f.OUM,colorSpace:p.outputColorSpace,stencilBuffer:z.stencil,resolveDepthBuffer:G.ignoreDepthValues===!1,resolveStencilBuffer:G.ignoreDepthValues===!1})}H.isXRRenderTarget=!0,this.setFoveation(j),$=null,I=await E.requestReferenceSpace(V),$e.setContext(E),$e.start(),x.isPresenting=!0,x.dispatchEvent({type:"sessionstart"})}},this.getEnvironmentBlendMode=function(){if(E!==null)return E.environmentBlendMode},this.getDepthTexture=function(){return A.getDepthTexture()};function it(xt){for(let Lt=0;Lt<xt.removed.length;Lt++){const Dt=xt.removed[Lt],Te=J.indexOf(Dt);Te>=0&&(J[Te]=null,ot[Te].disconnect(Dt))}for(let Lt=0;Lt<xt.added.length;Lt++){const Dt=xt.added[Lt];let Te=J.indexOf(Dt);if(Te===-1){for(let _e=0;_e<ot.length;_e++)if(_e>=J.length){J.push(Dt),Te=_e;break}else if(J[_e]===null){J[_e]=Dt,Te=_e;break}if(Te===-1)break}const de=ot[Te];de&&de.connect(Dt)}}const It=new f.Pq0,zt=new f.Pq0;function kt(xt,Lt,Dt){It.setFromMatrixPosition(Lt.matrixWorld),zt.setFromMatrixPosition(Dt.matrixWorld);const Te=It.distanceTo(zt),de=Lt.projectionMatrix.elements,_e=Dt.projectionMatrix.elements,Ut=de[14]/(de[10]-1),Ie=de[14]/(de[10]+1),Pe=(de[9]+1)/de[5],Be=(de[9]-1)/de[5],ae=(de[8]-1)/de[0],Je=(_e[8]+1)/_e[0],B=Ut*ae,Xe=Ut*Je,Ce=Te/(-ae+Je),Ee=Ce*-ae;if(Lt.matrixWorld.decompose(xt.position,xt.quaternion,xt.scale),xt.translateX(Ee),xt.translateZ(Ce),xt.matrixWorld.compose(xt.position,xt.quaternion,xt.scale),xt.matrixWorldInverse.copy(xt.matrixWorld).invert(),de[10]===-1)xt.projectionMatrix.copy(Lt.projectionMatrix),xt.projectionMatrixInverse.copy(Lt.projectionMatrixInverse);else{const Jt=Ut+Ce,P=Ie+Ce,y=B-Ee,W=Xe+(Te-Ee),ut=Pe*Ie/P*Jt,yt=Be*Ie/P*Jt;xt.projectionMatrix.makePerspective(y,W,ut,yt,Jt,P),xt.projectionMatrixInverse.copy(xt.projectionMatrix).invert()}}function ue(xt,Lt){Lt===null?xt.matrixWorld.copy(xt.matrix):xt.matrixWorld.multiplyMatrices(Lt.matrixWorld,xt.matrix),xt.matrixWorldInverse.copy(xt.matrixWorld).invert()}this.updateCamera=function(xt){if(E===null)return;let Lt=xt.near,Dt=xt.far;A.texture!==null&&(A.depthNear>0&&(Lt=A.depthNear),A.depthFar>0&&(Dt=A.depthFar)),dt.near=Tt.near=U.near=Lt,dt.far=Tt.far=U.far=Dt,(mt!==dt.near||gt!==dt.far)&&(E.updateRenderState({depthNear:dt.near,depthFar:dt.far}),mt=dt.near,gt=dt.far),dt.layers.mask=xt.layers.mask|6,U.layers.mask=dt.layers.mask&-5,Tt.layers.mask=dt.layers.mask&-3;const Te=xt.parent,de=dt.cameras;ue(dt,Te);for(let _e=0;_e<de.length;_e++)ue(de[_e],Te);de.length===2?kt(dt,U,Tt):dt.projectionMatrix.copy(U.projectionMatrix),te(xt,dt,Te)};function te(xt,Lt,Dt){Dt===null?xt.matrix.copy(Lt.matrixWorld):(xt.matrix.copy(Dt.matrixWorld),xt.matrix.invert(),xt.matrix.multiply(Lt.matrixWorld)),xt.matrix.decompose(xt.position,xt.quaternion,xt.scale),xt.updateMatrixWorld(!0),xt.projectionMatrix.copy(Lt.projectionMatrix),xt.projectionMatrixInverse.copy(Lt.projectionMatrixInverse),xt.isPerspectiveCamera&&(xt.fov=f.a55*2*Math.atan(1/xt.projectionMatrix.elements[5]),xt.zoom=1)}this.getCamera=function(){return dt},this.getFoveation=function(){if(!(F===null&&G===null))return j},this.setFoveation=function(xt){j=xt,F!==null&&(F.fixedFoveation=xt),G!==null&&G.fixedFoveation!==void 0&&(G.fixedFoveation=xt)},this.hasDepthSensing=function(){return A.texture!==null},this.getDepthSensingMesh=function(){return A.getMesh(dt)},this.getCameraTexture=function(xt){return b[xt]};let se=null;function Ze(xt,Lt){if(ht=Lt.getViewerPose($||I),X=Lt,ht!==null){const Dt=ht.views;G!==null&&(p.setRenderTargetFramebuffer(H,G.framebuffer),p.setRenderTarget(H));let Te=!1;Dt.length!==dt.cameras.length&&(dt.cameras.length=0,Te=!0);for(let Ie=0;Ie<Dt.length;Ie++){const Pe=Dt[Ie];let Be=null;if(G!==null)Be=G.getViewport(Pe);else{const Je=K.getViewSubImage(F,Pe);Be=Je.viewport,Ie===0&&(p.setRenderTargetTextures(H,Je.colorTexture,Je.depthStencilTexture),p.setRenderTarget(H))}let ae=Z[Ie];ae===void 0&&(ae=new f.ubm,ae.layers.enable(Ie),ae.viewport=new f.IUQ,Z[Ie]=ae),ae.matrix.fromArray(Pe.transform.matrix),ae.matrix.decompose(ae.position,ae.quaternion,ae.scale),ae.projectionMatrix.fromArray(Pe.projectionMatrix),ae.projectionMatrixInverse.copy(ae.projectionMatrix).invert(),ae.viewport.set(Be.x,Be.y,Be.width,Be.height),Ie===0&&(dt.matrix.copy(ae.matrix),dt.matrix.decompose(dt.position,dt.quaternion,dt.scale)),Te===!0&&dt.cameras.push(ae)}const de=E.enabledFeatures;if(de&&de.includes("depth-sensing")&&E.depthUsage=="gpu-optimized"&&nt){K=x.getBinding();const Ie=K.getDepthInformation(Dt[0]);Ie&&Ie.isValid&&Ie.texture&&A.init(Ie,E.renderState)}if(de&&de.includes("camera-access")&&nt){p.state.unbindTexture(),K=x.getBinding();for(let Ie=0;Ie<Dt.length;Ie++){const Pe=Dt[Ie].camera;if(Pe){let Be=b[Pe];Be||(Be=new f.rjZ,b[Pe]=Be);const ae=K.getCameraImage(Pe);Be.sourceTexture=ae}}}}for(let Dt=0;Dt<ot.length;Dt++){const Te=J[Dt],de=ot[Dt];Te!==null&&de!==void 0&&de.update(Te,Lt,$||I)}se&&se(xt,Lt),Lt.detectedPlanes&&x.dispatchEvent({type:"planesdetected",data:Lt}),X=null}const $e=new Oo;$e.setAnimationLoop(Ze),this.setAnimationLoop=function(xt){se=xt},this.dispose=function(){}}}const vi=new f.O9p,$a=new f.kn4;function li(o,p){function h(A,b){A.matrixAutoUpdate===!0&&A.updateMatrix(),b.value.copy(A.matrix)}function x(A,b){b.color.getRGB(A.fogColor.value,(0,f._Ut)(o)),b.isFog?(A.fogNear.value=b.near,A.fogFar.value=b.far):b.isFogExp2&&(A.fogDensity.value=b.density)}function E(A,b,z,q,H){b.isMeshBasicMaterial?S(A,b):b.isMeshLambertMaterial?(S(A,b),b.envMap&&(A.envMapIntensity.value=b.envMapIntensity)):b.isMeshToonMaterial?(S(A,b),K(A,b)):b.isMeshPhongMaterial?(S(A,b),ht(A,b),b.envMap&&(A.envMapIntensity.value=b.envMapIntensity)):b.isMeshStandardMaterial?(S(A,b),F(A,b),b.isMeshPhysicalMaterial&&G(A,b,H)):b.isMeshMatcapMaterial?(S(A,b),X(A,b)):b.isMeshDepthMaterial?S(A,b):b.isMeshDistanceMaterial?(S(A,b),nt(A,b)):b.isMeshNormalMaterial?S(A,b):b.isLineBasicMaterial?(I(A,b),b.isLineDashedMaterial&&V(A,b)):b.isPointsMaterial?j(A,b,z,q):b.isSpriteMaterial?$(A,b):b.isShadowMaterial?(A.color.value.copy(b.color),A.opacity.value=b.opacity):b.isShaderMaterial&&(b.uniformsNeedUpdate=!1)}function S(A,b){A.opacity.value=b.opacity,b.color&&A.diffuse.value.copy(b.color),b.emissive&&A.emissive.value.copy(b.emissive).multiplyScalar(b.emissiveIntensity),b.map&&(A.map.value=b.map,h(b.map,A.mapTransform)),b.alphaMap&&(A.alphaMap.value=b.alphaMap,h(b.alphaMap,A.alphaMapTransform)),b.bumpMap&&(A.bumpMap.value=b.bumpMap,h(b.bumpMap,A.bumpMapTransform),A.bumpScale.value=b.bumpScale,b.side===f.hsX&&(A.bumpScale.value*=-1)),b.normalMap&&(A.normalMap.value=b.normalMap,h(b.normalMap,A.normalMapTransform),A.normalScale.value.copy(b.normalScale),b.side===f.hsX&&A.normalScale.value.negate()),b.displacementMap&&(A.displacementMap.value=b.displacementMap,h(b.displacementMap,A.displacementMapTransform),A.displacementScale.value=b.displacementScale,A.displacementBias.value=b.displacementBias),b.emissiveMap&&(A.emissiveMap.value=b.emissiveMap,h(b.emissiveMap,A.emissiveMapTransform)),b.specularMap&&(A.specularMap.value=b.specularMap,h(b.specularMap,A.specularMapTransform)),b.alphaTest>0&&(A.alphaTest.value=b.alphaTest);const z=p.get(b),q=z.envMap,H=z.envMapRotation;q&&(A.envMap.value=q,vi.copy(H),vi.x*=-1,vi.y*=-1,vi.z*=-1,q.isCubeTexture&&q.isRenderTargetTexture===!1&&(vi.y*=-1,vi.z*=-1),A.envMapRotation.value.setFromMatrix4($a.makeRotationFromEuler(vi)),A.flipEnvMap.value=q.isCubeTexture&&q.isRenderTargetTexture===!1?-1:1,A.reflectivity.value=b.reflectivity,A.ior.value=b.ior,A.refractionRatio.value=b.refractionRatio),b.lightMap&&(A.lightMap.value=b.lightMap,A.lightMapIntensity.value=b.lightMapIntensity,h(b.lightMap,A.lightMapTransform)),b.aoMap&&(A.aoMap.value=b.aoMap,A.aoMapIntensity.value=b.aoMapIntensity,h(b.aoMap,A.aoMapTransform))}function I(A,b){A.diffuse.value.copy(b.color),A.opacity.value=b.opacity,b.map&&(A.map.value=b.map,h(b.map,A.mapTransform))}function V(A,b){A.dashSize.value=b.dashSize,A.totalSize.value=b.dashSize+b.gapSize,A.scale.value=b.scale}function j(A,b,z,q){A.diffuse.value.copy(b.color),A.opacity.value=b.opacity,A.size.value=b.size*z,A.scale.value=q*.5,b.map&&(A.map.value=b.map,h(b.map,A.uvTransform)),b.alphaMap&&(A.alphaMap.value=b.alphaMap,h(b.alphaMap,A.alphaMapTransform)),b.alphaTest>0&&(A.alphaTest.value=b.alphaTest)}function $(A,b){A.diffuse.value.copy(b.color),A.opacity.value=b.opacity,A.rotation.value=b.rotation,b.map&&(A.map.value=b.map,h(b.map,A.mapTransform)),b.alphaMap&&(A.alphaMap.value=b.alphaMap,h(b.alphaMap,A.alphaMapTransform)),b.alphaTest>0&&(A.alphaTest.value=b.alphaTest)}function ht(A,b){A.specular.value.copy(b.specular),A.shininess.value=Math.max(b.shininess,1e-4)}function K(A,b){b.gradientMap&&(A.gradientMap.value=b.gradientMap)}function F(A,b){A.metalness.value=b.metalness,b.metalnessMap&&(A.metalnessMap.value=b.metalnessMap,h(b.metalnessMap,A.metalnessMapTransform)),A.roughness.value=b.roughness,b.roughnessMap&&(A.roughnessMap.value=b.roughnessMap,h(b.roughnessMap,A.roughnessMapTransform)),b.envMap&&(A.envMapIntensity.value=b.envMapIntensity)}function G(A,b,z){A.ior.value=b.ior,b.sheen>0&&(A.sheenColor.value.copy(b.sheenColor).multiplyScalar(b.sheen),A.sheenRoughness.value=b.sheenRoughness,b.sheenColorMap&&(A.sheenColorMap.value=b.sheenColorMap,h(b.sheenColorMap,A.sheenColorMapTransform)),b.sheenRoughnessMap&&(A.sheenRoughnessMap.value=b.sheenRoughnessMap,h(b.sheenRoughnessMap,A.sheenRoughnessMapTransform))),b.clearcoat>0&&(A.clearcoat.value=b.clearcoat,A.clearcoatRoughness.value=b.clearcoatRoughness,b.clearcoatMap&&(A.clearcoatMap.value=b.clearcoatMap,h(b.clearcoatMap,A.clearcoatMapTransform)),b.clearcoatRoughnessMap&&(A.clearcoatRoughnessMap.value=b.clearcoatRoughnessMap,h(b.clearcoatRoughnessMap,A.clearcoatRoughnessMapTransform)),b.clearcoatNormalMap&&(A.clearcoatNormalMap.value=b.clearcoatNormalMap,h(b.clearcoatNormalMap,A.clearcoatNormalMapTransform),A.clearcoatNormalScale.value.copy(b.clearcoatNormalScale),b.side===f.hsX&&A.clearcoatNormalScale.value.negate())),b.dispersion>0&&(A.dispersion.value=b.dispersion),b.iridescence>0&&(A.iridescence.value=b.iridescence,A.iridescenceIOR.value=b.iridescenceIOR,A.iridescenceThicknessMinimum.value=b.iridescenceThicknessRange[0],A.iridescenceThicknessMaximum.value=b.iridescenceThicknessRange[1],b.iridescenceMap&&(A.iridescenceMap.value=b.iridescenceMap,h(b.iridescenceMap,A.iridescenceMapTransform)),b.iridescenceThicknessMap&&(A.iridescenceThicknessMap.value=b.iridescenceThicknessMap,h(b.iridescenceThicknessMap,A.iridescenceThicknessMapTransform))),b.transmission>0&&(A.transmission.value=b.transmission,A.transmissionSamplerMap.value=z.texture,A.transmissionSamplerSize.value.set(z.width,z.height),b.transmissionMap&&(A.transmissionMap.value=b.transmissionMap,h(b.transmissionMap,A.transmissionMapTransform)),A.thickness.value=b.thickness,b.thicknessMap&&(A.thicknessMap.value=b.thicknessMap,h(b.thicknessMap,A.thicknessMapTransform)),A.attenuationDistance.value=b.attenuationDistance,A.attenuationColor.value.copy(b.attenuationColor)),b.anisotropy>0&&(A.anisotropyVector.value.set(b.anisotropy*Math.cos(b.anisotropyRotation),b.anisotropy*Math.sin(b.anisotropyRotation)),b.anisotropyMap&&(A.anisotropyMap.value=b.anisotropyMap,h(b.anisotropyMap,A.anisotropyMapTransform))),A.specularIntensity.value=b.specularIntensity,A.specularColor.value.copy(b.specularColor),b.specularColorMap&&(A.specularColorMap.value=b.specularColorMap,h(b.specularColorMap,A.specularColorMapTransform)),b.specularIntensityMap&&(A.specularIntensityMap.value=b.specularIntensityMap,h(b.specularIntensityMap,A.specularIntensityMapTransform))}function X(A,b){b.matcap&&(A.matcap.value=b.matcap)}function nt(A,b){const z=p.get(b).light;A.referencePosition.value.setFromMatrixPosition(z.matrixWorld),A.nearDistance.value=z.shadow.camera.near,A.farDistance.value=z.shadow.camera.far}return{refreshFogUniforms:x,refreshMaterialUniforms:E}}function Os(o,p,h,x){let E={},S={},I=[];const V=o.getParameter(o.MAX_UNIFORM_BUFFER_BINDINGS);function j(z,q){const H=q.program;x.uniformBlockBinding(z,H)}function $(z,q){let H=E[z.id];H===void 0&&(X(z),H=ht(z),E[z.id]=H,z.addEventListener("dispose",A));const ot=q.program;x.updateUBOMapping(z,ot);const J=p.render.frame;S[z.id]!==J&&(F(z),S[z.id]=J)}function ht(z){const q=K();z.__bindingPointIndex=q;const H=o.createBuffer(),ot=z.__size,J=z.usage;return o.bindBuffer(o.UNIFORM_BUFFER,H),o.bufferData(o.UNIFORM_BUFFER,ot,J),o.bindBuffer(o.UNIFORM_BUFFER,null),o.bindBufferBase(o.UNIFORM_BUFFER,q,H),H}function K(){for(let z=0;z<V;z++)if(I.indexOf(z)===-1)return I.push(z),z;return(0,f.z3S)("WebGLRenderer: Maximum number of simultaneously usable uniforms groups reached."),0}function F(z){const q=E[z.id],H=z.uniforms,ot=z.__cache;o.bindBuffer(o.UNIFORM_BUFFER,q);for(let J=0,st=H.length;J<st;J++){const R=Array.isArray(H[J])?H[J]:[H[J]];for(let U=0,Tt=R.length;U<Tt;U++){const Z=R[U];if(G(Z,J,U,ot)===!0){const dt=Z.__offset,mt=Array.isArray(Z.value)?Z.value:[Z.value];let gt=0;for(let _t=0;_t<mt.length;_t++){const ct=mt[_t],it=nt(ct);typeof ct=="number"||typeof ct=="boolean"?(Z.__data[0]=ct,o.bufferSubData(o.UNIFORM_BUFFER,dt+gt,Z.__data)):ct.isMatrix3?(Z.__data[0]=ct.elements[0],Z.__data[1]=ct.elements[1],Z.__data[2]=ct.elements[2],Z.__data[3]=0,Z.__data[4]=ct.elements[3],Z.__data[5]=ct.elements[4],Z.__data[6]=ct.elements[5],Z.__data[7]=0,Z.__data[8]=ct.elements[6],Z.__data[9]=ct.elements[7],Z.__data[10]=ct.elements[8],Z.__data[11]=0):(ct.toArray(Z.__data,gt),gt+=it.storage/Float32Array.BYTES_PER_ELEMENT)}o.bufferSubData(o.UNIFORM_BUFFER,dt,Z.__data)}}}o.bindBuffer(o.UNIFORM_BUFFER,null)}function G(z,q,H,ot){const J=z.value,st=q+"_"+H;if(ot[st]===void 0)return typeof J=="number"||typeof J=="boolean"?ot[st]=J:ot[st]=J.clone(),!0;{const R=ot[st];if(typeof J=="number"||typeof J=="boolean"){if(R!==J)return ot[st]=J,!0}else if(R.equals(J)===!1)return R.copy(J),!0}return!1}function X(z){const q=z.uniforms;let H=0;const ot=16;for(let st=0,R=q.length;st<R;st++){const U=Array.isArray(q[st])?q[st]:[q[st]];for(let Tt=0,Z=U.length;Tt<Z;Tt++){const dt=U[Tt],mt=Array.isArray(dt.value)?dt.value:[dt.value];for(let gt=0,_t=mt.length;gt<_t;gt++){const ct=mt[gt],it=nt(ct),It=H%ot,zt=It%it.boundary,kt=It+zt;H+=zt,kt!==0&&ot-kt<it.storage&&(H+=ot-kt),dt.__data=new Float32Array(it.storage/Float32Array.BYTES_PER_ELEMENT),dt.__offset=H,H+=it.storage}}}const J=H%ot;return J>0&&(H+=ot-J),z.__size=H,z.__cache={},this}function nt(z){const q={boundary:0,storage:0};return typeof z=="number"||typeof z=="boolean"?(q.boundary=4,q.storage=4):z.isVector2?(q.boundary=8,q.storage=8):z.isVector3||z.isColor?(q.boundary=16,q.storage=12):z.isVector4?(q.boundary=16,q.storage=16):z.isMatrix3?(q.boundary=48,q.storage=48):z.isMatrix4?(q.boundary=64,q.storage=64):z.isTexture?(0,f.R8M)("WebGLRenderer: Texture samplers can not be part of an uniforms group."):(0,f.R8M)("WebGLRenderer: Unsupported uniform value type.",z),q}function A(z){const q=z.target;q.removeEventListener("dispose",A);const H=I.indexOf(q.__bindingPointIndex);I.splice(H,1),o.deleteBuffer(E[q.id]),delete E[q.id],delete S[q.id]}function b(){for(const z in E)o.deleteBuffer(E[z]);I=[],E={},S={}}return{bind:j,update:$,dispose:b}}const Ir=new Uint16Array([12469,15057,12620,14925,13266,14620,13807,14376,14323,13990,14545,13625,14713,13328,14840,12882,14931,12528,14996,12233,15039,11829,15066,11525,15080,11295,15085,10976,15082,10705,15073,10495,13880,14564,13898,14542,13977,14430,14158,14124,14393,13732,14556,13410,14702,12996,14814,12596,14891,12291,14937,11834,14957,11489,14958,11194,14943,10803,14921,10506,14893,10278,14858,9960,14484,14039,14487,14025,14499,13941,14524,13740,14574,13468,14654,13106,14743,12678,14818,12344,14867,11893,14889,11509,14893,11180,14881,10751,14852,10428,14812,10128,14765,9754,14712,9466,14764,13480,14764,13475,14766,13440,14766,13347,14769,13070,14786,12713,14816,12387,14844,11957,14860,11549,14868,11215,14855,10751,14825,10403,14782,10044,14729,9651,14666,9352,14599,9029,14967,12835,14966,12831,14963,12804,14954,12723,14936,12564,14917,12347,14900,11958,14886,11569,14878,11247,14859,10765,14828,10401,14784,10011,14727,9600,14660,9289,14586,8893,14508,8533,15111,12234,15110,12234,15104,12216,15092,12156,15067,12010,15028,11776,14981,11500,14942,11205,14902,10752,14861,10393,14812,9991,14752,9570,14682,9252,14603,8808,14519,8445,14431,8145,15209,11449,15208,11451,15202,11451,15190,11438,15163,11384,15117,11274,15055,10979,14994,10648,14932,10343,14871,9936,14803,9532,14729,9218,14645,8742,14556,8381,14461,8020,14365,7603,15273,10603,15272,10607,15267,10619,15256,10631,15231,10614,15182,10535,15118,10389,15042,10167,14963,9787,14883,9447,14800,9115,14710,8665,14615,8318,14514,7911,14411,7507,14279,7198,15314,9675,15313,9683,15309,9712,15298,9759,15277,9797,15229,9773,15166,9668,15084,9487,14995,9274,14898,8910,14800,8539,14697,8234,14590,7790,14479,7409,14367,7067,14178,6621,15337,8619,15337,8631,15333,8677,15325,8769,15305,8871,15264,8940,15202,8909,15119,8775,15022,8565,14916,8328,14804,8009,14688,7614,14569,7287,14448,6888,14321,6483,14088,6171,15350,7402,15350,7419,15347,7480,15340,7613,15322,7804,15287,7973,15229,8057,15148,8012,15046,7846,14933,7611,14810,7357,14682,7069,14552,6656,14421,6316,14251,5948,14007,5528,15356,5942,15356,5977,15353,6119,15348,6294,15332,6551,15302,6824,15249,7044,15171,7122,15070,7050,14949,6861,14818,6611,14679,6349,14538,6067,14398,5651,14189,5311,13935,4958,15359,4123,15359,4153,15356,4296,15353,4646,15338,5160,15311,5508,15263,5829,15188,6042,15088,6094,14966,6001,14826,5796,14678,5543,14527,5287,14377,4985,14133,4586,13869,4257,15360,1563,15360,1642,15358,2076,15354,2636,15341,3350,15317,4019,15273,4429,15203,4732,15105,4911,14981,4932,14836,4818,14679,4621,14517,4386,14359,4156,14083,3795,13808,3437,15360,122,15360,137,15358,285,15355,636,15344,1274,15322,2177,15281,2765,15215,3223,15120,3451,14995,3569,14846,3567,14681,3466,14511,3305,14344,3121,14037,2800,13753,2467,15360,0,15360,1,15359,21,15355,89,15346,253,15325,479,15287,796,15225,1148,15133,1492,15008,1749,14856,1882,14685,1886,14506,1783,14324,1608,13996,1398,13702,1183]);let re=null;function xn(){return re===null&&(re=new f.GYF(Ir,16,16,f.paN,f.ix0),re.name="DFG_LUT",re.minFilter=f.k6q,re.magFilter=f.k6q,re.wrapS=f.ghU,re.wrapT=f.ghU,re.generateMipmaps=!1,re.needsUpdate=!0),re}class Pr{constructor(p={}){const{canvas:h=(0,f.lPF)(),context:x=null,depth:E=!0,stencil:S=!1,alpha:I=!1,antialias:V=!1,premultipliedAlpha:j=!0,preserveDrawingBuffer:$=!1,powerPreference:ht="default",failIfMajorPerformanceCaveat:K=!1,reversedDepthBuffer:F=!1,outputBufferType:G=f.OUM}=p;this.isWebGLRenderer=!0;let X;if(x!==null){if(typeof WebGLRenderingContext<"u"&&x instanceof WebGLRenderingContext)throw new Error("THREE.WebGLRenderer: WebGL 1 is not supported since r163.");X=x.getContextAttributes().alpha}else X=I;const nt=G,A=new Set([f.c90,f.TkQ,f.ZQM]),b=new Set([f.OUM,f.bkx,f.cHt,f.V3x,f.Wew,f.gJ2]),z=new Uint32Array(4),q=new Int32Array(4);let H=null,ot=null;const J=[],st=[];let R=null;this.domElement=h,this.debug={checkShaderErrors:!0,onShaderError:null},this.autoClear=!0,this.autoClearColor=!0,this.autoClearDepth=!0,this.autoClearStencil=!0,this.sortObjects=!0,this.clippingPlanes=[],this.localClippingEnabled=!1,this.toneMapping=f.y_p,this.toneMappingExposure=1,this.transmissionResolutionScale=1;const U=this;let Tt=!1;this._outputColorSpace=f.er$;let Z=0,dt=0,mt=null,gt=-1,_t=null;const ct=new f.IUQ,it=new f.IUQ;let It=null;const zt=new f.Q1f(0);let kt=0,ue=h.width,te=h.height,se=1,Ze=null,$e=null;const xt=new f.IUQ(0,0,ue,te),Lt=new f.IUQ(0,0,ue,te);let Dt=!1;const Te=new f.PPD;let de=!1,_e=!1;const Ut=new f.kn4,Ie=new f.Pq0,Pe=new f.IUQ,Be={background:null,fog:null,environment:null,overrideMaterial:null,isScene:!0};let ae=!1;function Je(){return mt===null?se:1}let B=x;function Xe(T,Y){return h.getContext(T,Y)}try{const T={alpha:!0,depth:E,stencil:S,antialias:V,premultipliedAlpha:j,preserveDrawingBuffer:$,powerPreference:ht,failIfMajorPerformanceCaveat:K};if("setAttribute"in h&&h.setAttribute("data-engine",`three.js r${f.sPf}`),h.addEventListener("webglcontextlost",$t,!1),h.addEventListener("webglcontextrestored",me,!1),h.addEventListener("webglcontextcreationerror",Ge,!1),B===null){const Y="webgl2";if(B=Xe(Y,T),B===null)throw Xe(Y)?new Error("Error creating WebGL context with your selected attributes."):new Error("Error creating WebGL context.")}}catch(T){throw(0,f.z3S)("WebGLRenderer: "+T.message),T}let Ce,Ee,Jt,P,y,W,ut,yt,pt,Ft,Pt,ne,ce,bt,wt,qt,Zt,Ht,Se,k,Rt,Ct,Xt;function Et(){Ce=new Hh(B),Ce.init(),Rt=new zi(B,Ce),Ee=new Oh(B,Ce,p,Rt),Jt=new Za(B,Ce),Ee.reversedDepthBuffer&&F&&Jt.buffers.depth.setReversed(!0),P=new qh(B),y=new Cc,W=new Lc(B,Ce,Jt,y,Ee,Rt,P),ut=new Gh(U),yt=new Bo(B),Ct=new ba(B,yt),pt=new Wh(B,yt,P,Ct),Ft=new Zh(B,pt,yt,Ct,P),Ht=new Yh(B,Ee,W),wt=new Bh(y),Pt=new On(U,ut,Ce,Ee,Ct,wt),ne=new li(U,y),ce=new wn,bt=new Oi(Ce),Zt=new fr(U,ut,Jt,Ft,X,j),qt=new Ya(U,Ft,Ee),Xt=new Os(B,P,Ee,Jt),Se=new Fh(B,Ce,P),k=new Xh(B,Ce,P),P.programs=Pt.programs,U.capabilities=Ee,U.extensions=Ce,U.properties=y,U.renderLists=ce,U.shadowMap=qt,U.state=Jt,U.info=P}Et(),nt!==f.OUM&&(R=new Aa(nt,h.width,h.height,E,S));const ft=new Dc(U,B);this.xr=ft,this.getContext=function(){return B},this.getContextAttributes=function(){return B.getContextAttributes()},this.forceContextLoss=function(){const T=Ce.get("WEBGL_lose_context");T&&T.loseContext()},this.forceContextRestore=function(){const T=Ce.get("WEBGL_lose_context");T&&T.restoreContext()},this.getPixelRatio=function(){return se},this.setPixelRatio=function(T){T!==void 0&&(se=T,this.setSize(ue,te,!1))},this.getSize=function(T){return T.set(ue,te)},this.setSize=function(T,Y,lt=!0){if(ft.isPresenting){(0,f.R8M)("WebGLRenderer: Can't change size while VR device is presenting.");return}ue=T,te=Y,h.width=Math.floor(T*se),h.height=Math.floor(Y*se),lt===!0&&(h.style.width=T+"px",h.style.height=Y+"px"),R!==null&&R.setSize(h.width,h.height),this.setViewport(0,0,T,Y)},this.getDrawingBufferSize=function(T){return T.set(ue*se,te*se).floor()},this.setDrawingBufferSize=function(T,Y,lt){ue=T,te=Y,se=lt,h.width=Math.floor(T*lt),h.height=Math.floor(Y*lt),this.setViewport(0,0,T,Y)},this.setEffects=function(T){if(nt===f.OUM){console.error("THREE.WebGLRenderer: setEffects() requires outputBufferType set to HalfFloatType or FloatType.");return}if(T){for(let Y=0;Y<T.length;Y++)if(T[Y].isOutputPass===!0){console.warn("THREE.WebGLRenderer: OutputPass is not needed in setEffects(). Tone mapping and color space conversion are applied automatically.");break}}R.setEffects(T||[])},this.getCurrentViewport=function(T){return T.copy(ct)},this.getViewport=function(T){return T.copy(xt)},this.setViewport=function(T,Y,lt,tt){T.isVector4?xt.set(T.x,T.y,T.z,T.w):xt.set(T,Y,lt,tt),Jt.viewport(ct.copy(xt).multiplyScalar(se).round())},this.getScissor=function(T){return T.copy(Lt)},this.setScissor=function(T,Y,lt,tt){T.isVector4?Lt.set(T.x,T.y,T.z,T.w):Lt.set(T,Y,lt,tt),Jt.scissor(it.copy(Lt).multiplyScalar(se).round())},this.getScissorTest=function(){return Dt},this.setScissorTest=function(T){Jt.setScissorTest(Dt=T)},this.setOpaqueSort=function(T){Ze=T},this.setTransparentSort=function(T){$e=T},this.getClearColor=function(T){return T.copy(Zt.getClearColor())},this.setClearColor=function(){Zt.setClearColor(...arguments)},this.getClearAlpha=function(){return Zt.getClearAlpha()},this.setClearAlpha=function(){Zt.setClearAlpha(...arguments)},this.clear=function(T=!0,Y=!0,lt=!0){let tt=0;if(T){let Q=!1;if(mt!==null){const Vt=mt.texture.format;Q=A.has(Vt)}if(Q){const Vt=mt.texture.type,Wt=b.has(Vt),Ot=Zt.getClearColor(),jt=Zt.getClearAlpha(),ie=Ot.r,be=Ot.g,Ae=Ot.b;Wt?(z[0]=ie,z[1]=be,z[2]=Ae,z[3]=jt,B.clearBufferuiv(B.COLOR,0,z)):(q[0]=ie,q[1]=be,q[2]=Ae,q[3]=jt,B.clearBufferiv(B.COLOR,0,q))}else tt|=B.COLOR_BUFFER_BIT}Y&&(tt|=B.DEPTH_BUFFER_BIT),lt&&(tt|=B.STENCIL_BUFFER_BIT,this.state.buffers.stencil.setMask(4294967295)),tt!==0&&B.clear(tt)},this.clearColor=function(){this.clear(!0,!1,!1)},this.clearDepth=function(){this.clear(!1,!0,!1)},this.clearStencil=function(){this.clear(!1,!1,!0)},this.dispose=function(){h.removeEventListener("webglcontextlost",$t,!1),h.removeEventListener("webglcontextrestored",me,!1),h.removeEventListener("webglcontextcreationerror",Ge,!1),Zt.dispose(),ce.dispose(),bt.dispose(),y.dispose(),ut.dispose(),Ft.dispose(),Ct.dispose(),Xt.dispose(),Pt.dispose(),ft.dispose(),ft.removeEventListener("sessionstart",os),ft.removeEventListener("sessionend",Bs),Vn.stop()};function $t(T){T.preventDefault(),(0,f.Rm2)("WebGLRenderer: Context Lost."),Tt=!0}function me(){(0,f.Rm2)("WebGLRenderer: Context Restored."),Tt=!1;const T=P.autoReset,Y=qt.enabled,lt=qt.autoUpdate,tt=qt.needsUpdate,Q=qt.type;Et(),P.autoReset=T,qt.enabled=Y,qt.autoUpdate=lt,qt.needsUpdate=tt,qt.type=Q}function Ge(T){(0,f.z3S)("WebGLRenderer: A WebGL context could not be created. Reason: ",T.statusMessage)}function ze(T){const Y=T.target;Y.removeEventListener("dispose",ze),fn(Y)}function fn(T){zn(T),y.remove(T)}function zn(T){const Y=y.get(T).programs;Y!==void 0&&(Y.forEach(function(lt){Pt.releaseProgram(lt)}),T.isShaderMaterial&&Pt.releaseShaderCache(T))}this.renderBufferDirect=function(T,Y,lt,tt,Q,Vt){Y===null&&(Y=Be);const Wt=Q.isMesh&&Q.matrixWorld.determinant()<0,Ot=Nr(T,Y,lt,tt,Q);Jt.setMaterial(tt,Wt);let jt=lt.index,ie=1;if(tt.wireframe===!0){if(jt=pt.getWireframeAttribute(lt),jt===void 0)return;ie=2}const be=lt.drawRange,Ae=lt.attributes.position;let Yt=be.start*ie,Ve=(be.start+be.count)*ie;Vt!==null&&(Yt=Math.max(Yt,Vt.start*ie),Ve=Math.min(Ve,(Vt.start+Vt.count)*ie)),jt!==null?(Yt=Math.max(Yt,0),Ve=Math.min(Ve,jt.count)):Ae!=null&&(Yt=Math.max(Yt,0),Ve=Math.min(Ve,Ae.count));const Ke=Ve-Yt;if(Ke<0||Ke===1/0)return;Ct.setup(Q,tt,Ot,lt,jt);let He,Ue=Se;if(jt!==null&&(He=yt.get(jt),Ue=k,Ue.setIndex(He)),Q.isMesh)tt.wireframe===!0?(Jt.setLineWidth(tt.wireframeLinewidth*Je()),Ue.setMode(B.LINES)):Ue.setMode(B.TRIANGLES);else if(Q.isLine){let Qe=tt.linewidth;Qe===void 0&&(Qe=1),Jt.setLineWidth(Qe*Je()),Q.isLineSegments?Ue.setMode(B.LINES):Q.isLineLoop?Ue.setMode(B.LINE_LOOP):Ue.setMode(B.LINE_STRIP)}else Q.isPoints?Ue.setMode(B.POINTS):Q.isSprite&&Ue.setMode(B.TRIANGLES);if(Q.isBatchedMesh)if(Q._multiDrawInstances!==null)(0,f.mcG)("WebGLRenderer: renderMultiDrawInstances has been deprecated and will be removed in r184. Append to renderMultiDraw arguments and use indirection."),Ue.renderMultiDrawInstances(Q._multiDrawStarts,Q._multiDrawCounts,Q._multiDrawCount,Q._multiDrawInstances);else if(Ce.get("WEBGL_multi_draw"))Ue.renderMultiDraw(Q._multiDrawStarts,Q._multiDrawCounts,Q._multiDrawCount);else{const Qe=Q._multiDrawStarts,ee=Q._multiDrawCounts,vn=Q._multiDrawCount,Le=jt?yt.get(jt).bytesPerElement:1,Gn=y.get(tt).currentProgram.getUniforms();for(let Sn=0;Sn<vn;Sn++)Gn.setValue(B,"_gl_DrawID",Sn),Ue.render(Qe[Sn]/Le,ee[Sn])}else if(Q.isInstancedMesh)Ue.renderInstances(Yt,Ke,Q.count);else if(lt.isInstancedBufferGeometry){const Qe=lt._maxInstanceCount!==void 0?lt._maxInstanceCount:1/0,ee=Math.min(lt.instanceCount,Qe);Ue.renderInstances(Yt,Ke,ee)}else Ue.render(Yt,Ke)};function as(T,Y,lt){T.transparent===!0&&T.side===f.$EB&&T.forceSinglePass===!1?(T.side=f.hsX,T.needsUpdate=!0,Mi(T,Y,lt),T.side=f.hB5,T.needsUpdate=!0,Mi(T,Y,lt),T.side=f.$EB):Mi(T,Y,lt)}this.compile=function(T,Y,lt=null){lt===null&&(lt=T),ot=bt.get(lt),ot.init(Y),st.push(ot),lt.traverseVisible(function(Q){Q.isLight&&Q.layers.test(Y.layers)&&(ot.pushLight(Q),Q.castShadow&&ot.pushShadow(Q))}),T!==lt&&T.traverseVisible(function(Q){Q.isLight&&Q.layers.test(Y.layers)&&(ot.pushLight(Q),Q.castShadow&&ot.pushShadow(Q))}),ot.setupLights();const tt=new Set;return T.traverse(function(Q){if(!(Q.isMesh||Q.isPoints||Q.isLine||Q.isSprite))return;const Vt=Q.material;if(Vt)if(Array.isArray(Vt))for(let Wt=0;Wt<Vt.length;Wt++){const Ot=Vt[Wt];as(Ot,lt,Q),tt.add(Ot)}else as(Vt,lt,Q),tt.add(Vt)}),ot=st.pop(),tt},this.compileAsync=function(T,Y,lt=null){const tt=this.compile(T,Y,lt);return new Promise(Q=>{function Vt(){if(tt.forEach(function(Wt){y.get(Wt).currentProgram.isReady()&&tt.delete(Wt)}),tt.size===0){Q(T);return}setTimeout(Vt,10)}Ce.get("KHR_parallel_shader_compile")!==null?Vt():setTimeout(Vt,10)})};let Zn=null;function Lr(T){Zn&&Zn(T)}function os(){Vn.stop()}function Bs(){Vn.start()}const Vn=new Oo;Vn.setAnimationLoop(Lr),typeof self<"u"&&Vn.setContext(self),this.setAnimationLoop=function(T){Zn=T,ft.setAnimationLoop(T),T===null?Vn.stop():Vn.start()},ft.addEventListener("sessionstart",os),ft.addEventListener("sessionend",Bs),this.render=function(T,Y){if(Y!==void 0&&Y.isCamera!==!0){(0,f.z3S)("WebGLRenderer.render: camera is not an instance of THREE.Camera.");return}if(Tt===!0)return;const lt=ft.enabled===!0&&ft.isPresenting===!0,tt=R!==null&&(mt===null||lt)&&R.begin(U,mt);if(T.matrixWorldAutoUpdate===!0&&T.updateMatrixWorld(),Y.parent===null&&Y.matrixWorldAutoUpdate===!0&&Y.updateMatrixWorld(),ft.enabled===!0&&ft.isPresenting===!0&&(R===null||R.isCompositing()===!1)&&(ft.cameraAutoUpdate===!0&&ft.updateCamera(Y),Y=ft.getCamera()),T.isScene===!0&&T.onBeforeRender(U,T,Y,mt),ot=bt.get(T,st.length),ot.init(Y),st.push(ot),Ut.multiplyMatrices(Y.projectionMatrix,Y.matrixWorldInverse),Te.setFromProjectionMatrix(Ut,f.TdN,Y.reversedDepth),_e=this.localClippingEnabled,de=wt.init(this.clippingPlanes,_e),H=ce.get(T,J.length),H.init(),J.push(H),ft.enabled===!0&&ft.isPresenting===!0){const Wt=U.xr.getDepthSensingMesh();Wt!==null&&ei(Wt,Y,-1/0,U.sortObjects)}ei(T,Y,0,U.sortObjects),H.finish(),U.sortObjects===!0&&H.sort(Ze,$e),ae=ft.enabled===!1||ft.isPresenting===!1||ft.hasDepthSensing()===!1,ae&&Zt.addToRenderList(H,T),this.info.render.frame++,de===!0&&wt.beginShadows();const Q=ot.state.shadowsArray;if(qt.render(Q,T,Y),de===!0&&wt.endShadows(),this.info.autoReset===!0&&this.info.reset(),(tt&&R.hasRenderPass())===!1){const Wt=H.opaque,Ot=H.transmissive;if(ot.setupLights(),Y.isArrayCamera){const jt=Y.cameras;if(Ot.length>0)for(let ie=0,be=jt.length;ie<be;ie++){const Ae=jt[ie];ci(Wt,Ot,T,Ae)}ae&&Zt.render(T);for(let ie=0,be=jt.length;ie<be;ie++){const Ae=jt[ie];Dr(H,T,Ae,Ae.viewport)}}else Ot.length>0&&ci(Wt,Ot,T,Y),ae&&Zt.render(T),Dr(H,T,Y)}mt!==null&&dt===0&&(W.updateMultisampleRenderTarget(mt),W.updateRenderTargetMipmap(mt)),tt&&R.end(U),T.isScene===!0&&T.onAfterRender(U,T,Y),Ct.resetDefaultState(),gt=-1,_t=null,st.pop(),st.length>0?(ot=st[st.length-1],de===!0&&wt.setGlobalState(U.clippingPlanes,ot.state.camera)):ot=null,J.pop(),J.length>0?H=J[J.length-1]:H=null};function ei(T,Y,lt,tt){if(T.visible===!1)return;if(T.layers.test(Y.layers)){if(T.isGroup)lt=T.renderOrder;else if(T.isLOD)T.autoUpdate===!0&&T.update(Y);else if(T.isLight)ot.pushLight(T),T.castShadow&&ot.pushShadow(T);else if(T.isSprite){if(!T.frustumCulled||Te.intersectsSprite(T)){tt&&Pe.setFromMatrixPosition(T.matrixWorld).applyMatrix4(Ut);const Wt=Ft.update(T),Ot=T.material;Ot.visible&&H.push(T,Wt,Ot,lt,Pe.z,null)}}else if((T.isMesh||T.isLine||T.isPoints)&&(!T.frustumCulled||Te.intersectsObject(T))){const Wt=Ft.update(T),Ot=T.material;if(tt&&(T.boundingSphere!==void 0?(T.boundingSphere===null&&T.computeBoundingSphere(),Pe.copy(T.boundingSphere.center)):(Wt.boundingSphere===null&&Wt.computeBoundingSphere(),Pe.copy(Wt.boundingSphere.center)),Pe.applyMatrix4(T.matrixWorld).applyMatrix4(Ut)),Array.isArray(Ot)){const jt=Wt.groups;for(let ie=0,be=jt.length;ie<be;ie++){const Ae=jt[ie],Yt=Ot[Ae.materialIndex];Yt&&Yt.visible&&H.push(T,Wt,Yt,lt,Pe.z,Ae)}}else Ot.visible&&H.push(T,Wt,Ot,lt,Pe.z,null)}}const Vt=T.children;for(let Wt=0,Ot=Vt.length;Wt<Ot;Wt++)ei(Vt[Wt],Y,lt,tt)}function Dr(T,Y,lt,tt){const{opaque:Q,transmissive:Vt,transparent:Wt}=T;ot.setupLightsView(lt),de===!0&&wt.setGlobalState(U.clippingPlanes,lt),tt&&Jt.viewport(ct.copy(tt)),Q.length>0&&yi(Q,Y,lt),Vt.length>0&&yi(Vt,Y,lt),Wt.length>0&&yi(Wt,Y,lt),Jt.buffers.depth.setTest(!0),Jt.buffers.depth.setMask(!0),Jt.buffers.color.setMask(!0),Jt.setPolygonOffset(!1)}function ci(T,Y,lt,tt){if((lt.isScene===!0?lt.overrideMaterial:null)!==null)return;if(ot.state.transmissionRenderTarget[tt.id]===void 0){const Yt=Ce.has("EXT_color_buffer_half_float")||Ce.has("EXT_color_buffer_float");ot.state.transmissionRenderTarget[tt.id]=new f.nWS(1,1,{generateMipmaps:!0,type:Yt?f.ix0:f.OUM,minFilter:f.$_I,samples:Math.max(4,Ee.samples),stencilBuffer:S,resolveDepthBuffer:!1,resolveStencilBuffer:!1,colorSpace:f.ppV.workingColorSpace})}const Vt=ot.state.transmissionRenderTarget[tt.id],Wt=tt.viewport||ct;Vt.setSize(Wt.z*U.transmissionResolutionScale,Wt.w*U.transmissionResolutionScale);const Ot=U.getRenderTarget(),jt=U.getActiveCubeFace(),ie=U.getActiveMipmapLevel();U.setRenderTarget(Vt),U.getClearColor(zt),kt=U.getClearAlpha(),kt<1&&U.setClearColor(16777215,.5),U.clear(),ae&&Zt.render(lt);const be=U.toneMapping;U.toneMapping=f.y_p;const Ae=tt.viewport;if(tt.viewport!==void 0&&(tt.viewport=void 0),ot.setupLightsView(tt),de===!0&&wt.setGlobalState(U.clippingPlanes,tt),yi(T,lt,tt),W.updateMultisampleRenderTarget(Vt),W.updateRenderTargetMipmap(Vt),Ce.has("WEBGL_multisampled_render_to_texture")===!1){let Yt=!1;for(let Ve=0,Ke=Y.length;Ve<Ke;Ve++){const He=Y[Ve],{object:Ue,geometry:Qe,material:ee,group:vn}=He;if(ee.side===f.$EB&&Ue.layers.test(tt.layers)){const Le=ee.side;ee.side=f.hsX,ee.needsUpdate=!0,Ur(Ue,lt,tt,Qe,ee,vn),ee.side=Le,ee.needsUpdate=!0,Yt=!0}}Yt===!0&&(W.updateMultisampleRenderTarget(Vt),W.updateRenderTargetMipmap(Vt))}U.setRenderTarget(Ot,jt,ie),U.setClearColor(zt,kt),Ae!==void 0&&(tt.viewport=Ae),U.toneMapping=be}function yi(T,Y,lt){const tt=Y.isScene===!0?Y.overrideMaterial:null;for(let Q=0,Vt=T.length;Q<Vt;Q++){const Wt=T[Q],{object:Ot,geometry:jt,group:ie}=Wt;let be=Wt.material;be.allowOverride===!0&&tt!==null&&(be=tt),Ot.layers.test(lt.layers)&&Ur(Ot,Y,lt,jt,be,ie)}}function Ur(T,Y,lt,tt,Q,Vt){T.onBeforeRender(U,Y,lt,tt,Q,Vt),T.modelViewMatrix.multiplyMatrices(lt.matrixWorldInverse,T.matrixWorld),T.normalMatrix.getNormalMatrix(T.modelViewMatrix),Q.onBeforeRender(U,Y,lt,tt,T,Vt),Q.transparent===!0&&Q.side===f.$EB&&Q.forceSinglePass===!1?(Q.side=f.hsX,Q.needsUpdate=!0,U.renderBufferDirect(lt,Y,tt,Q,T,Vt),Q.side=f.hB5,Q.needsUpdate=!0,U.renderBufferDirect(lt,Y,tt,Q,T,Vt),Q.side=f.$EB):U.renderBufferDirect(lt,Y,tt,Q,T,Vt),T.onAfterRender(U,Y,lt,tt,Q,Vt)}function Mi(T,Y,lt){Y.isScene!==!0&&(Y=Be);const tt=y.get(T),Q=ot.state.lights,Vt=ot.state.shadowsArray,Wt=Q.state.version,Ot=Pt.getParameters(T,Q.state,Vt,Y,lt),jt=Pt.getProgramCacheKey(Ot);let ie=tt.programs;tt.environment=T.isMeshStandardMaterial||T.isMeshLambertMaterial||T.isMeshPhongMaterial?Y.environment:null,tt.fog=Y.fog;const be=T.isMeshStandardMaterial||T.isMeshLambertMaterial&&!T.envMap||T.isMeshPhongMaterial&&!T.envMap;tt.envMap=ut.get(T.envMap||tt.environment,be),tt.envMapRotation=tt.environment!==null&&T.envMap===null?Y.environmentRotation:T.envMapRotation,ie===void 0&&(T.addEventListener("dispose",ze),ie=new Map,tt.programs=ie);let Ae=ie.get(jt);if(Ae!==void 0){if(tt.currentProgram===Ae&&tt.lightsStateVersion===Wt)return cs(T,Ot),Ae}else Ot.uniforms=Pt.getUniforms(T),T.onBeforeCompile(Ot,U),Ae=Pt.acquireProgram(Ot,jt),ie.set(jt,Ae),tt.uniforms=Ot.uniforms;const Yt=tt.uniforms;return(!T.isShaderMaterial&&!T.isRawShaderMaterial||T.clipping===!0)&&(Yt.clippingPlanes=wt.uniform),cs(T,Ot),tt.needsLights=Ja(T),tt.lightsStateVersion=Wt,tt.needsLights&&(Yt.ambientLightColor.value=Q.state.ambient,Yt.lightProbe.value=Q.state.probe,Yt.directionalLights.value=Q.state.directional,Yt.directionalLightShadows.value=Q.state.directionalShadow,Yt.spotLights.value=Q.state.spot,Yt.spotLightShadows.value=Q.state.spotShadow,Yt.rectAreaLights.value=Q.state.rectArea,Yt.ltc_1.value=Q.state.rectAreaLTC1,Yt.ltc_2.value=Q.state.rectAreaLTC2,Yt.pointLights.value=Q.state.point,Yt.pointLightShadows.value=Q.state.pointShadow,Yt.hemisphereLights.value=Q.state.hemi,Yt.directionalShadowMatrix.value=Q.state.directionalShadowMatrix,Yt.spotLightMatrix.value=Q.state.spotLightMatrix,Yt.spotLightMap.value=Q.state.spotLightMap,Yt.pointShadowMatrix.value=Q.state.pointShadowMatrix),tt.currentProgram=Ae,tt.uniformsList=null,Ae}function ls(T){if(T.uniformsList===null){const Y=T.currentProgram.getUniforms();T.uniformsList=Us.seqWithValue(Y.seq,T.uniforms)}return T.uniformsList}function cs(T,Y){const lt=y.get(T);lt.outputColorSpace=Y.outputColorSpace,lt.batching=Y.batching,lt.batchingColor=Y.batchingColor,lt.instancing=Y.instancing,lt.instancingColor=Y.instancingColor,lt.instancingMorph=Y.instancingMorph,lt.skinning=Y.skinning,lt.morphTargets=Y.morphTargets,lt.morphNormals=Y.morphNormals,lt.morphColors=Y.morphColors,lt.morphTargetsCount=Y.morphTargetsCount,lt.numClippingPlanes=Y.numClippingPlanes,lt.numIntersection=Y.numClipIntersection,lt.vertexAlphas=Y.vertexAlphas,lt.vertexTangents=Y.vertexTangents,lt.toneMapping=Y.toneMapping}function Nr(T,Y,lt,tt,Q){Y.isScene!==!0&&(Y=Be),W.resetTextureUnits();const Vt=Y.fog,Wt=tt.isMeshStandardMaterial||tt.isMeshLambertMaterial||tt.isMeshPhongMaterial?Y.environment:null,Ot=mt===null?U.outputColorSpace:mt.isXRRenderTarget===!0?mt.texture.colorSpace:f.Zr2,jt=tt.isMeshStandardMaterial||tt.isMeshLambertMaterial&&!tt.envMap||tt.isMeshPhongMaterial&&!tt.envMap,ie=ut.get(tt.envMap||Wt,jt),be=tt.vertexColors===!0&&!!lt.attributes.color&&lt.attributes.color.itemSize===4,Ae=!!lt.attributes.tangent&&(!!tt.normalMap||tt.anisotropy>0),Yt=!!lt.morphAttributes.position,Ve=!!lt.morphAttributes.normal,Ke=!!lt.morphAttributes.color;let He=f.y_p;tt.toneMapped&&(mt===null||mt.isXRRenderTarget===!0)&&(He=U.toneMapping);const Ue=lt.morphAttributes.position||lt.morphAttributes.normal||lt.morphAttributes.color,Qe=Ue!==void 0?Ue.length:0,ee=y.get(tt),vn=ot.state.lights;if(de===!0&&(_e===!0||T!==_t)){const We=T===_t&&tt.id===gt;wt.setState(tt,T,We)}let Le=!1;tt.version===ee.__version?(ee.needsLights&&ee.lightsStateVersion!==vn.state.version||ee.outputColorSpace!==Ot||Q.isBatchedMesh&&ee.batching===!1||!Q.isBatchedMesh&&ee.batching===!0||Q.isBatchedMesh&&ee.batchingColor===!0&&Q.colorTexture===null||Q.isBatchedMesh&&ee.batchingColor===!1&&Q.colorTexture!==null||Q.isInstancedMesh&&ee.instancing===!1||!Q.isInstancedMesh&&ee.instancing===!0||Q.isSkinnedMesh&&ee.skinning===!1||!Q.isSkinnedMesh&&ee.skinning===!0||Q.isInstancedMesh&&ee.instancingColor===!0&&Q.instanceColor===null||Q.isInstancedMesh&&ee.instancingColor===!1&&Q.instanceColor!==null||Q.isInstancedMesh&&ee.instancingMorph===!0&&Q.morphTexture===null||Q.isInstancedMesh&&ee.instancingMorph===!1&&Q.morphTexture!==null||ee.envMap!==ie||tt.fog===!0&&ee.fog!==Vt||ee.numClippingPlanes!==void 0&&(ee.numClippingPlanes!==wt.numPlanes||ee.numIntersection!==wt.numIntersection)||ee.vertexAlphas!==be||ee.vertexTangents!==Ae||ee.morphTargets!==Yt||ee.morphNormals!==Ve||ee.morphColors!==Ke||ee.toneMapping!==He||ee.morphTargetsCount!==Qe)&&(Le=!0):(Le=!0,ee.__version=tt.version);let Gn=ee.currentProgram;Le===!0&&(Gn=Mi(tt,Y,Q));let Sn=!1,Hn=!1,Vi=!1;const ke=Gn.getUniforms(),dn=ee.uniforms;if(Jt.useProgram(Gn.program)&&(Sn=!0,Hn=!0,Vi=!0),tt.id!==gt&&(gt=tt.id,Hn=!0),Sn||_t!==T){Jt.buffers.depth.getReversed()&&T.reversedDepth!==!0&&(T._reversedDepth=!0,T.updateProjectionMatrix()),ke.setValue(B,"projectionMatrix",T.projectionMatrix),ke.setValue(B,"viewMatrix",T.matrixWorldInverse);const bn=ke.map.cameraPosition;bn!==void 0&&bn.setValue(B,Ie.setFromMatrixPosition(T.matrixWorld)),Ee.logarithmicDepthBuffer&&ke.setValue(B,"logDepthBufFC",2/(Math.log(T.far+1)/Math.LN2)),(tt.isMeshPhongMaterial||tt.isMeshToonMaterial||tt.isMeshLambertMaterial||tt.isMeshBasicMaterial||tt.isMeshStandardMaterial||tt.isShaderMaterial)&&ke.setValue(B,"isOrthographic",T.isOrthographicCamera===!0),_t!==T&&(_t=T,Hn=!0,Vi=!0)}if(ee.needsLights&&(vn.state.directionalShadowMap.length>0&&ke.setValue(B,"directionalShadowMap",vn.state.directionalShadowMap,W),vn.state.spotShadowMap.length>0&&ke.setValue(B,"spotShadowMap",vn.state.spotShadowMap,W),vn.state.pointShadowMap.length>0&&ke.setValue(B,"pointShadowMap",vn.state.pointShadowMap,W)),Q.isSkinnedMesh){ke.setOptional(B,Q,"bindMatrix"),ke.setOptional(B,Q,"bindMatrixInverse");const We=Q.skeleton;We&&(We.boneTexture===null&&We.computeBoneTexture(),ke.setValue(B,"boneTexture",We.boneTexture,W))}Q.isBatchedMesh&&(ke.setOptional(B,Q,"batchingTexture"),ke.setValue(B,"batchingTexture",Q._matricesTexture,W),ke.setOptional(B,Q,"batchingIdTexture"),ke.setValue(B,"batchingIdTexture",Q._indirectTexture,W),ke.setOptional(B,Q,"batchingColorTexture"),Q._colorsTexture!==null&&ke.setValue(B,"batchingColorTexture",Q._colorsTexture,W));const hi=lt.morphAttributes;if((hi.position!==void 0||hi.normal!==void 0||hi.color!==void 0)&&Ht.update(Q,lt,Gn),(Hn||ee.receiveShadow!==Q.receiveShadow)&&(ee.receiveShadow=Q.receiveShadow,ke.setValue(B,"receiveShadow",Q.receiveShadow)),(tt.isMeshStandardMaterial||tt.isMeshLambertMaterial||tt.isMeshPhongMaterial)&&tt.envMap===null&&Y.environment!==null&&(dn.envMapIntensity.value=Y.environmentIntensity),dn.dfgLUT!==void 0&&(dn.dfgLUT.value=xn()),Hn&&(ke.setValue(B,"toneMappingExposure",U.toneMappingExposure),ee.needsLights&&zs(dn,Vi),Vt&&tt.fog===!0&&ne.refreshFogUniforms(dn,Vt),ne.refreshMaterialUniforms(dn,tt,se,te,ot.state.transmissionRenderTarget[T.id]),Us.upload(B,ls(ee),dn,W)),tt.isShaderMaterial&&tt.uniformsNeedUpdate===!0&&(Us.upload(B,ls(ee),dn,W),tt.uniformsNeedUpdate=!1),tt.isSpriteMaterial&&ke.setValue(B,"center",Q.center),ke.setValue(B,"modelViewMatrix",Q.modelViewMatrix),ke.setValue(B,"normalMatrix",Q.normalMatrix),ke.setValue(B,"modelMatrix",Q.matrixWorld),tt.isShaderMaterial||tt.isRawShaderMaterial){const We=tt.uniformsGroups;for(let bn=0,ki=We.length;bn<ki;bn++){const hs=We[bn];Xt.update(hs,Gn),Xt.bind(hs,Gn)}}return Gn}function zs(T,Y){T.ambientLightColor.needsUpdate=Y,T.lightProbe.needsUpdate=Y,T.directionalLights.needsUpdate=Y,T.directionalLightShadows.needsUpdate=Y,T.pointLights.needsUpdate=Y,T.pointLightShadows.needsUpdate=Y,T.spotLights.needsUpdate=Y,T.spotLightShadows.needsUpdate=Y,T.rectAreaLights.needsUpdate=Y,T.hemisphereLights.needsUpdate=Y}function Ja(T){return T.isMeshLambertMaterial||T.isMeshToonMaterial||T.isMeshPhongMaterial||T.isMeshStandardMaterial||T.isShadowMaterial||T.isShaderMaterial&&T.lights===!0}this.getActiveCubeFace=function(){return Z},this.getActiveMipmapLevel=function(){return dt},this.getRenderTarget=function(){return mt},this.setRenderTargetTextures=function(T,Y,lt){const tt=y.get(T);tt.__autoAllocateDepthBuffer=T.resolveDepthBuffer===!1,tt.__autoAllocateDepthBuffer===!1&&(tt.__useRenderToTexture=!1),y.get(T.texture).__webglTexture=Y,y.get(T.depthTexture).__webglTexture=tt.__autoAllocateDepthBuffer?void 0:lt,tt.__hasExternalTextures=!0},this.setRenderTargetFramebuffer=function(T,Y){const lt=y.get(T);lt.__webglFramebuffer=Y,lt.__useDefaultFramebuffer=Y===void 0};const Vs=B.createFramebuffer();this.setRenderTarget=function(T,Y=0,lt=0){mt=T,Z=Y,dt=lt;let tt=null,Q=!1,Vt=!1;if(T){const Ot=y.get(T);if(Ot.__useDefaultFramebuffer!==void 0){Jt.bindFramebuffer(B.FRAMEBUFFER,Ot.__webglFramebuffer),ct.copy(T.viewport),it.copy(T.scissor),It=T.scissorTest,Jt.viewport(ct),Jt.scissor(it),Jt.setScissorTest(It),gt=-1;return}else if(Ot.__webglFramebuffer===void 0)W.setupRenderTarget(T);else if(Ot.__hasExternalTextures)W.rebindTextures(T,y.get(T.texture).__webglTexture,y.get(T.depthTexture).__webglTexture);else if(T.depthBuffer){const be=T.depthTexture;if(Ot.__boundDepthTexture!==be){if(be!==null&&y.has(be)&&(T.width!==be.image.width||T.height!==be.image.height))throw new Error("WebGLRenderTarget: Attached DepthTexture is initialized to the incorrect size.");W.setupDepthRenderbuffer(T)}}const jt=T.texture;(jt.isData3DTexture||jt.isDataArrayTexture||jt.isCompressedArrayTexture)&&(Vt=!0);const ie=y.get(T).__webglFramebuffer;T.isWebGLCubeRenderTarget?(Array.isArray(ie[Y])?tt=ie[Y][lt]:tt=ie[Y],Q=!0):T.samples>0&&W.useMultisampledRTT(T)===!1?tt=y.get(T).__webglMultisampledFramebuffer:Array.isArray(ie)?tt=ie[lt]:tt=ie,ct.copy(T.viewport),it.copy(T.scissor),It=T.scissorTest}else ct.copy(xt).multiplyScalar(se).floor(),it.copy(Lt).multiplyScalar(se).floor(),It=Dt;if(lt!==0&&(tt=Vs),Jt.bindFramebuffer(B.FRAMEBUFFER,tt)&&Jt.drawBuffers(T,tt),Jt.viewport(ct),Jt.scissor(it),Jt.setScissorTest(It),Q){const Ot=y.get(T.texture);B.framebufferTexture2D(B.FRAMEBUFFER,B.COLOR_ATTACHMENT0,B.TEXTURE_CUBE_MAP_POSITIVE_X+Y,Ot.__webglTexture,lt)}else if(Vt){const Ot=Y;for(let jt=0;jt<T.textures.length;jt++){const ie=y.get(T.textures[jt]);B.framebufferTextureLayer(B.FRAMEBUFFER,B.COLOR_ATTACHMENT0+jt,ie.__webglTexture,lt,Ot)}}else if(T!==null&&lt!==0){const Ot=y.get(T.texture);B.framebufferTexture2D(B.FRAMEBUFFER,B.COLOR_ATTACHMENT0,B.TEXTURE_2D,Ot.__webglTexture,lt)}gt=-1},this.readRenderTargetPixels=function(T,Y,lt,tt,Q,Vt,Wt,Ot=0){if(!(T&&T.isWebGLRenderTarget)){(0,f.z3S)("WebGLRenderer.readRenderTargetPixels: renderTarget is not THREE.WebGLRenderTarget.");return}let jt=y.get(T).__webglFramebuffer;if(T.isWebGLCubeRenderTarget&&Wt!==void 0&&(jt=jt[Wt]),jt){Jt.bindFramebuffer(B.FRAMEBUFFER,jt);try{const ie=T.textures[Ot],be=ie.format,Ae=ie.type;if(T.textures.length>1&&B.readBuffer(B.COLOR_ATTACHMENT0+Ot),!Ee.textureFormatReadable(be)){(0,f.z3S)("WebGLRenderer.readRenderTargetPixels: renderTarget is not in RGBA or implementation defined format.");return}if(!Ee.textureTypeReadable(Ae)){(0,f.z3S)("WebGLRenderer.readRenderTargetPixels: renderTarget is not in UnsignedByteType or implementation defined type.");return}Y>=0&&Y<=T.width-tt&&lt>=0&&lt<=T.height-Q&&B.readPixels(Y,lt,tt,Q,Rt.convert(be),Rt.convert(Ae),Vt)}finally{const ie=mt!==null?y.get(mt).__webglFramebuffer:null;Jt.bindFramebuffer(B.FRAMEBUFFER,ie)}}},this.readRenderTargetPixelsAsync=async function(T,Y,lt,tt,Q,Vt,Wt,Ot=0){if(!(T&&T.isWebGLRenderTarget))throw new Error("THREE.WebGLRenderer.readRenderTargetPixels: renderTarget is not THREE.WebGLRenderTarget.");let jt=y.get(T).__webglFramebuffer;if(T.isWebGLCubeRenderTarget&&Wt!==void 0&&(jt=jt[Wt]),jt)if(Y>=0&&Y<=T.width-tt&&lt>=0&&lt<=T.height-Q){Jt.bindFramebuffer(B.FRAMEBUFFER,jt);const ie=T.textures[Ot],be=ie.format,Ae=ie.type;if(T.textures.length>1&&B.readBuffer(B.COLOR_ATTACHMENT0+Ot),!Ee.textureFormatReadable(be))throw new Error("THREE.WebGLRenderer.readRenderTargetPixelsAsync: renderTarget is not in RGBA or implementation defined format.");if(!Ee.textureTypeReadable(Ae))throw new Error("THREE.WebGLRenderer.readRenderTargetPixelsAsync: renderTarget is not in UnsignedByteType or implementation defined type.");const Yt=B.createBuffer();B.bindBuffer(B.PIXEL_PACK_BUFFER,Yt),B.bufferData(B.PIXEL_PACK_BUFFER,Vt.byteLength,B.STREAM_READ),B.readPixels(Y,lt,tt,Q,Rt.convert(be),Rt.convert(Ae),0);const Ve=mt!==null?y.get(mt).__webglFramebuffer:null;Jt.bindFramebuffer(B.FRAMEBUFFER,Ve);const Ke=B.fenceSync(B.SYNC_GPU_COMMANDS_COMPLETE,0);return B.flush(),await(0,f.jej)(B,Ke,4),B.bindBuffer(B.PIXEL_PACK_BUFFER,Yt),B.getBufferSubData(B.PIXEL_PACK_BUFFER,0,Vt),B.deleteBuffer(Yt),B.deleteSync(Ke),Vt}else throw new Error("THREE.WebGLRenderer.readRenderTargetPixelsAsync: requested read bounds are out of range.")},this.copyFramebufferToTexture=function(T,Y=null,lt=0){const tt=Math.pow(2,-lt),Q=Math.floor(T.image.width*tt),Vt=Math.floor(T.image.height*tt),Wt=Y!==null?Y.x:0,Ot=Y!==null?Y.y:0;W.setTexture2D(T,0),B.copyTexSubImage2D(B.TEXTURE_2D,lt,0,0,Wt,Ot,Q,Vt),Jt.unbindTexture()};const kn=B.createFramebuffer(),Uc=B.createFramebuffer();this.copyTextureToTexture=function(T,Y,lt=null,tt=null,Q=0,Vt=0){let Wt,Ot,jt,ie,be,Ae,Yt,Ve,Ke;const He=T.isCompressedTexture?T.mipmaps[Vt]:T.image;if(lt!==null)Wt=lt.max.x-lt.min.x,Ot=lt.max.y-lt.min.y,jt=lt.isBox3?lt.max.z-lt.min.z:1,ie=lt.min.x,be=lt.min.y,Ae=lt.isBox3?lt.min.z:0;else{const dn=Math.pow(2,-Q);Wt=Math.floor(He.width*dn),Ot=Math.floor(He.height*dn),T.isDataArrayTexture?jt=He.depth:T.isData3DTexture?jt=Math.floor(He.depth*dn):jt=1,ie=0,be=0,Ae=0}tt!==null?(Yt=tt.x,Ve=tt.y,Ke=tt.z):(Yt=0,Ve=0,Ke=0);const Ue=Rt.convert(Y.format),Qe=Rt.convert(Y.type);let ee;Y.isData3DTexture?(W.setTexture3D(Y,0),ee=B.TEXTURE_3D):Y.isDataArrayTexture||Y.isCompressedArrayTexture?(W.setTexture2DArray(Y,0),ee=B.TEXTURE_2D_ARRAY):(W.setTexture2D(Y,0),ee=B.TEXTURE_2D),B.pixelStorei(B.UNPACK_FLIP_Y_WEBGL,Y.flipY),B.pixelStorei(B.UNPACK_PREMULTIPLY_ALPHA_WEBGL,Y.premultiplyAlpha),B.pixelStorei(B.UNPACK_ALIGNMENT,Y.unpackAlignment);const vn=B.getParameter(B.UNPACK_ROW_LENGTH),Le=B.getParameter(B.UNPACK_IMAGE_HEIGHT),Gn=B.getParameter(B.UNPACK_SKIP_PIXELS),Sn=B.getParameter(B.UNPACK_SKIP_ROWS),Hn=B.getParameter(B.UNPACK_SKIP_IMAGES);B.pixelStorei(B.UNPACK_ROW_LENGTH,He.width),B.pixelStorei(B.UNPACK_IMAGE_HEIGHT,He.height),B.pixelStorei(B.UNPACK_SKIP_PIXELS,ie),B.pixelStorei(B.UNPACK_SKIP_ROWS,be),B.pixelStorei(B.UNPACK_SKIP_IMAGES,Ae);const Vi=T.isDataArrayTexture||T.isData3DTexture,ke=Y.isDataArrayTexture||Y.isData3DTexture;if(T.isDepthTexture){const dn=y.get(T),hi=y.get(Y),We=y.get(dn.__renderTarget),bn=y.get(hi.__renderTarget);Jt.bindFramebuffer(B.READ_FRAMEBUFFER,We.__webglFramebuffer),Jt.bindFramebuffer(B.DRAW_FRAMEBUFFER,bn.__webglFramebuffer);for(let ki=0;ki<jt;ki++)Vi&&(B.framebufferTextureLayer(B.READ_FRAMEBUFFER,B.COLOR_ATTACHMENT0,y.get(T).__webglTexture,Q,Ae+ki),B.framebufferTextureLayer(B.DRAW_FRAMEBUFFER,B.COLOR_ATTACHMENT0,y.get(Y).__webglTexture,Vt,Ke+ki)),B.blitFramebuffer(ie,be,Wt,Ot,Yt,Ve,Wt,Ot,B.DEPTH_BUFFER_BIT,B.NEAREST);Jt.bindFramebuffer(B.READ_FRAMEBUFFER,null),Jt.bindFramebuffer(B.DRAW_FRAMEBUFFER,null)}else if(Q!==0||T.isRenderTargetTexture||y.has(T)){const dn=y.get(T),hi=y.get(Y);Jt.bindFramebuffer(B.READ_FRAMEBUFFER,kn),Jt.bindFramebuffer(B.DRAW_FRAMEBUFFER,Uc);for(let We=0;We<jt;We++)Vi?B.framebufferTextureLayer(B.READ_FRAMEBUFFER,B.COLOR_ATTACHMENT0,dn.__webglTexture,Q,Ae+We):B.framebufferTexture2D(B.READ_FRAMEBUFFER,B.COLOR_ATTACHMENT0,B.TEXTURE_2D,dn.__webglTexture,Q),ke?B.framebufferTextureLayer(B.DRAW_FRAMEBUFFER,B.COLOR_ATTACHMENT0,hi.__webglTexture,Vt,Ke+We):B.framebufferTexture2D(B.DRAW_FRAMEBUFFER,B.COLOR_ATTACHMENT0,B.TEXTURE_2D,hi.__webglTexture,Vt),Q!==0?B.blitFramebuffer(ie,be,Wt,Ot,Yt,Ve,Wt,Ot,B.COLOR_BUFFER_BIT,B.NEAREST):ke?B.copyTexSubImage3D(ee,Vt,Yt,Ve,Ke+We,ie,be,Wt,Ot):B.copyTexSubImage2D(ee,Vt,Yt,Ve,ie,be,Wt,Ot);Jt.bindFramebuffer(B.READ_FRAMEBUFFER,null),Jt.bindFramebuffer(B.DRAW_FRAMEBUFFER,null)}else ke?T.isDataTexture||T.isData3DTexture?B.texSubImage3D(ee,Vt,Yt,Ve,Ke,Wt,Ot,jt,Ue,Qe,He.data):Y.isCompressedArrayTexture?B.compressedTexSubImage3D(ee,Vt,Yt,Ve,Ke,Wt,Ot,jt,Ue,He.data):B.texSubImage3D(ee,Vt,Yt,Ve,Ke,Wt,Ot,jt,Ue,Qe,He):T.isDataTexture?B.texSubImage2D(B.TEXTURE_2D,Vt,Yt,Ve,Wt,Ot,Ue,Qe,He.data):T.isCompressedTexture?B.compressedTexSubImage2D(B.TEXTURE_2D,Vt,Yt,Ve,He.width,He.height,Ue,He.data):B.texSubImage2D(B.TEXTURE_2D,Vt,Yt,Ve,Wt,Ot,Ue,Qe,He);B.pixelStorei(B.UNPACK_ROW_LENGTH,vn),B.pixelStorei(B.UNPACK_IMAGE_HEIGHT,Le),B.pixelStorei(B.UNPACK_SKIP_PIXELS,Gn),B.pixelStorei(B.UNPACK_SKIP_ROWS,Sn),B.pixelStorei(B.UNPACK_SKIP_IMAGES,Hn),Vt===0&&Y.generateMipmaps&&B.generateMipmap(ee),Jt.unbindTexture()},this.initRenderTarget=function(T){y.get(T).__webglFramebuffer===void 0&&W.setupRenderTarget(T)},this.initTexture=function(T){T.isCubeTexture?W.setTextureCube(T,0):T.isData3DTexture?W.setTexture3D(T,0):T.isDataArrayTexture||T.isCompressedArrayTexture?W.setTexture2DArray(T,0):W.setTexture2D(T,0),Jt.unbindTexture()},this.resetState=function(){Z=0,dt=0,mt=null,Jt.reset(),Ct.reset()},typeof __THREE_DEVTOOLS__<"u"&&__THREE_DEVTOOLS__.dispatchEvent(new CustomEvent("observe",{detail:this}))}get coordinateSystem(){return f.TdN}get outputColorSpace(){return this._outputColorSpace}set outputColorSpace(p){this._outputColorSpace=p;const h=this.getContext();h.drawingBufferColorSpace=f.ppV._getDrawingBufferColorSpace(p),h.unpackColorSpace=f.ppV._getUnpackColorSpace()}}}),81072:(function(bf,Fo,Xi){Xi.d(Fo,{$EB:function(){return Wo},$Yl:function(){return hl},$_I:function(){return Rs},$ei:function(){return jo},A$4:function(){return Dt},BKk:function(){return ch},BVL:function(){return Il},BXX:function(){return ya},B_h:function(){return Nl},CSG:function(){return Ou},CVz:function(){return wl},CWW:function(){return tc},Cfg:function(){return ta},D$Q:function(){return Nu},Dmk:function(){return ca},EZo:function(){return qo},EdD:function(){return Zo},F1T:function(){return ep},FFZ:function(){return La},FV:function(){return vl},FXf:function(){return Jo},Fn:function(){return Zl},GJx:function(){return ws},GWd:function(){return Ci},GYF:function(){return Yt},Gc6:function(){return Ff},Gwm:function(){return rr},H23:function(){return $l},HIg:function(){return fa},HO_:function(){return jl},HXV:function(){return Al},I9Y:function(){return At},IE4:function(){return _a},IUQ:function(){return En},Iit:function(){return Uu},Jnc:function(){return Vo},K52:function(){return ar},KDk:function(){return Rl},KLL:function(){return Ds},KRh:function(){return dl},Kef:function(){return Kl},Kwu:function(){return Yo},LAk:function(){return Ml},LiQ:function(){return il},LlO:function(){return Sd},LoY:function(){return y},MW4:function(){return de},Mjd:function(){return _l},NTi:function(){return Ks},Nex:function(){return gf},Nt7:function(){return rl},Nz6:function(){return xa},O9p:function(){return Bn},OUM:function(){return Is},Om:function(){return Qr},OtU:function(){return Ll},OuU:function(){return tr},PPD:function(){return Ka},Pq0:function(){return O},Q1f:function(){return re},QP0:function(){return ko},Qev:function(){return Kn},Qrf:function(){return Bl},R3r:function(){return vi},R8M:function(){return le},RQf:function(){return gi},Riy:function(){return Cl},Rm2:function(){return gr},RrE:function(){return cl},RyA:function(){return Ho},S$4:function(){return Yl},THS:function(){return se},TdN:function(){return Nn},TiK:function(){return Ca},TkQ:function(){return ma},U3G:function(){return sr},V3x:function(){return Tl},V9B:function(){return ei},VCu:function(){return xu},VT0:function(){return cr},Vb5:function(){return zo},VxR:function(){return Ki},W9U:function(){return Jl},WNZ:function(){return Bo},Wdf:function(){return lc},Wew:function(){return oa},Wk7:function(){return Go},XG_:function(){return Ql},XIg:function(){return Xo},XrR:function(){return pl},Yuy:function(){return ra},Z58:function(){return p},ZQM:function(){return hr},Zcv:function(){return We},Zr2:function(){return Ji},_QJ:function(){return kl},_Ut:function(){return Md},a55:function(){return Di},a5J:function(){return Vl},aEY:function(){return ol},aJ8:function(){return Sl},aVO:function(){return Bu},amv:function(){return Ea},b4q:function(){return kc},bC7:function(){return Wl},bCz:function(){return $o},bI3:function(){return Un},bdM:function(){return xo},bkx:function(){return Zi},brA:function(){return ir},bw0:function(){return or},c90:function(){return ga},cHt:function(){return sa},caT:function(){return Es},czI:function(){return Fl},dYF:function(){return Er},dcC:function(){return da},dwI:function(){return Fn},e0p:function(){return ul},eHc:function(){return er},eaF:function(){return kn},eoi:function(){return Ra},er$:function(){return yn},f4X:function(){return nl},fBL:function(){return ia},g7M:function(){return yl},gJ2:function(){return la},gO9:function(){return Qs},gPd:function(){return Ye},gWB:function(){return Pa},gZr:function(){return Pl},ghU:function(){return Ln},hB5:function(){return As},hdd:function(){return sl},hgQ:function(){return ll},hsX:function(){return Yr},hxR:function(){return Mn},hy7:function(){return Yi},iNn:function(){return ao},ie2:function(){return js},ix0:function(){return aa},jR7:function(){return va},jSS:function(){return Dl},jej:function(){return fc},jf0:function(){return $i},jzd:function(){return Ia},k6Q:function(){return Ma},k6q:function(){return Cn},kO0:function(){return wa},kRr:function(){return ea},kTW:function(){return Cs},kTp:function(){return Sa},kn4:function(){return Me},kyO:function(){return gl},lGu:function(){return nr},lPF:function(){return uc},lxW:function(){return yo},lyL:function(){return Hl},mcG:function(){return _r},nNL:function(){return xl},nST:function(){return Ko},nWS:function(){return Ar},nZQ:function(){return np},ojh:function(){return tl},ojs:function(){return ql},ov9:function(){return fl},pBf:function(){return El},pHI:function(){return jr},paN:function(){return pa},ppV:function(){return _n},psI:function(){return zl},qUd:function(){return mh},qa3:function(){return Ul},qad:function(){return el},qq$:function(){return es},qtW:function(){return Ut},rFo:function(){return Ha},rSH:function(){return Ol},ri6:function(){return dc},rjZ:function(){return Of},sPf:function(){return qr},tJf:function(){return na},uB5:function(){return Gl},uV5:function(){return Kr},ubm:function(){return Xn},vim:function(){return pr},vyJ:function(){return Ta},wfO:function(){return Jr},wn6:function(){return al},wrO:function(){return ua},xFO:function(){return $r},xSv:function(){return qi},y3Z:function(){return Xl},yT7:function(){return ha},y_p:function(){return ml},z3S:function(){return Re},zdS:function(){return Ps},zgK:function(){return wr},znC:function(){return Qo}});var Tf=Xi(82759);/**
 * @license
 * Copyright 2010-2026 Three.js Authors
 * SPDX-License-Identifier: MIT
 */const qr="183",f={LEFT:0,MIDDLE:1,RIGHT:2,ROTATE:0,DOLLY:1,PAN:2},Oo={ROTATE:0,PAN:1,DOLLY_PAN:2,DOLLY_ROTATE:3},Bo=0,zo=1,Vo=2,Ah=3,Eh=0,ko=1,Go=2,Ho=3,As=0,Yr=1,Wo=2,Xo=0,Ks=1,qo=2,Yo=3,Zo=4,$o=5,wh=6,Qs=100,Jo=101,Ko=102,Qo=103,jo=104,tl=200,el=201,nl=202,il=203,js=204,tr=205,sl=206,rl=207,al=208,ol=209,ll=210,cl=211,hl=212,ul=213,fl=214,er=0,nr=1,ir=2,qi=3,sr=4,rr=5,ar=6,or=7,Es=0,dl=1,pl=2,ml=0,gl=1,_l=2,xl=3,vl=4,yl=5,Ml=6,Sl=7,Zr="attached",bl="detached",lr=300,Yi=301,$r=302,Jr=303,Kr=304,Qr=306,ws=1e3,Ln=1001,Cs=1002,Mn=1003,jr=1004,Ch=1004,ta=1005,Rh=1005,Cn=1006,ea=1007,Ih=1007,Rs=1008,Ph=1008,Is=1009,na=1010,ia=1011,sa=1012,ra=1013,Zi=1014,gi=1015,aa=1016,oa=1017,la=1018,Tl=1020,ca=35902,ha=35899,ua=1021,fa=1022,Ci=1023,Ps=1026,da=1027,cr=1028,hr=1029,pa=1030,ma=1031,Lh=1032,ga=1033,_a=33776,xa=33777,va=33778,ya=33779,Ma=35840,Sa=35841,Al=35842,El=35843,wl=36196,Cl=37492,Rl=37496,Il=37488,Pl=37489,Ll=37490,Dl=37491,Ul=37808,Nl=37809,Fl=37810,Ol=37811,Bl=37812,zl=37813,Vl=37814,kl=37815,Gl=37816,Hl=37817,Wl=37818,Xl=37819,ql=37820,Yl=37821,Zl=36492,$l=36494,Jl=36495,Kl=36283,Ql=36284,jl=36285,tc=36286,Dh=2200,Uh=2201,Nh=2202,ur=2300,ve=2301,Nt=2302,Dn=2303,Jn=2400,Rn=2401,Ls=2402,fr=2500,ba=2501,Fh=0,Oh=1,Bh=2,ri=3200,ec=3201,Ri=3202,zh=3203,Un=0,Ta=1,$i="",yn="srgb",Ji="srgb-linear",Ki="linear",Ds="srgb",nc="",Vh="rg",ic="ga",Qi=0,Ii=7680,kh=7681,sc=7682,rc=7683,dr=34055,ac=34056,Gh=5386,Hh=512,Wh=513,Xh=514,qh=515,Yh=516,Zh=517,$h=518,Aa=519,Ea=512,pr=513,wa=514,Ca=515,Ra=516,Ia=517,Pa=518,La=519,ji=35044,oc=35048,ts=35040,on=35045,ln=35049,mr=35041,Jh=35046,Kh=35050,Qh=35042,jh="100",lc="300 es",Nn=2e3,Pi=2001,tu={COMPUTE:"compute",RENDER:"render"},eu={PERSPECTIVE:"perspective",LINEAR:"linear",FLAT:"flat"},nu={NORMAL:"normal",CENTROID:"centroid",SAMPLE:"sample",FIRST:"first",EITHER:"either"},iu={TEXTURE_COMPARE:"depthTextureCompare"};function cc(u){for(let t=u.length-1;t>=0;--t)if(u[t]>=65535)return!0;return!1}const hc={Int8Array,Uint8Array,Uint8ClampedArray,Int16Array,Uint16Array,Int32Array,Uint32Array,Float32Array,Float64Array};function Li(u,t){return new hc[u](t)}function Da(u){return ArrayBuffer.isView(u)&&!(u instanceof DataView)}function es(u){return document.createElementNS("http://www.w3.org/1999/xhtml",u)}function uc(){const u=es("canvas");return u.style.display="block",u}const Ua={};let ai=null;function su(u){ai=u}function ru(){return ai}function gr(...u){const t="THREE."+u.shift();ai?ai("log",t,...u):console.log(t,...u)}function Na(u){const t=u[0];if(typeof t=="string"&&t.startsWith("TSL:")){const e=u[1];e&&e.isStackTrace?u[0]+=" "+e.getLocation():u[1]='Stack trace not available. Enable "THREE.Node.captureStackTrace" to capture stack traces.'}return u}function le(...u){u=Na(u);const t="THREE."+u.shift();if(ai)ai("warn",t,...u);else{const e=u[0];e&&e.isStackTrace?console.warn(e.getError(t)):console.warn(t,...u)}}function Re(...u){u=Na(u);const t="THREE."+u.shift();if(ai)ai("error",t,...u);else{const e=u[0];e&&e.isStackTrace?console.error(e.getError(t)):console.error(t,...u)}}function _r(...u){const t=u.join(" ");t in Ua||(Ua[t]=!0,le(...u))}function fc(u,t,e){return new Promise(function(n,i){function s(){switch(u.clientWaitSync(t,u.SYNC_FLUSH_COMMANDS_BIT,0)){case u.WAIT_FAILED:i();break;case u.TIMEOUT_EXPIRED:setTimeout(s,e);break;default:n()}}setTimeout(s,e)})}const dc={[er]:nr,[ir]:ar,[sr]:or,[qi]:rr,[nr]:er,[ar]:ir,[or]:sr,[rr]:qi};class Kn{addEventListener(t,e){this._listeners===void 0&&(this._listeners={});const n=this._listeners;n[t]===void 0&&(n[t]=[]),n[t].indexOf(e)===-1&&n[t].push(e)}hasEventListener(t,e){const n=this._listeners;return n===void 0?!1:n[t]!==void 0&&n[t].indexOf(e)!==-1}removeEventListener(t,e){const n=this._listeners;if(n===void 0)return;const i=n[t];if(i!==void 0){const s=i.indexOf(e);s!==-1&&i.splice(s,1)}}dispatchEvent(t){const e=this._listeners;if(e===void 0)return;const n=e[t.type];if(n!==void 0){t.target=this;const i=n.slice(0);for(let s=0,r=i.length;s<r;s++)i[s].call(this,t);t.target=null}}}const mn=["00","01","02","03","04","05","06","07","08","09","0a","0b","0c","0d","0e","0f","10","11","12","13","14","15","16","17","18","19","1a","1b","1c","1d","1e","1f","20","21","22","23","24","25","26","27","28","29","2a","2b","2c","2d","2e","2f","30","31","32","33","34","35","36","37","38","39","3a","3b","3c","3d","3e","3f","40","41","42","43","44","45","46","47","48","49","4a","4b","4c","4d","4e","4f","50","51","52","53","54","55","56","57","58","59","5a","5b","5c","5d","5e","5f","60","61","62","63","64","65","66","67","68","69","6a","6b","6c","6d","6e","6f","70","71","72","73","74","75","76","77","78","79","7a","7b","7c","7d","7e","7f","80","81","82","83","84","85","86","87","88","89","8a","8b","8c","8d","8e","8f","90","91","92","93","94","95","96","97","98","99","9a","9b","9c","9d","9e","9f","a0","a1","a2","a3","a4","a5","a6","a7","a8","a9","aa","ab","ac","ad","ae","af","b0","b1","b2","b3","b4","b5","b6","b7","b8","b9","ba","bb","bc","bd","be","bf","c0","c1","c2","c3","c4","c5","c6","c7","c8","c9","ca","cb","cc","cd","ce","cf","d0","d1","d2","d3","d4","d5","d6","d7","d8","d9","da","db","dc","dd","de","df","e0","e1","e2","e3","e4","e5","e6","e7","e8","e9","ea","eb","ec","ed","ee","ef","f0","f1","f2","f3","f4","f5","f6","f7","f8","f9","fa","fb","fc","fd","fe","ff"];let Fa=1234567;const _i=Math.PI/180,Di=180/Math.PI;function An(){const u=Math.random()*4294967295|0,t=Math.random()*4294967295|0,e=Math.random()*4294967295|0,n=Math.random()*4294967295|0;return(mn[u&255]+mn[u>>8&255]+mn[u>>16&255]+mn[u>>24&255]+"-"+mn[t&255]+mn[t>>8&255]+"-"+mn[t>>16&15|64]+mn[t>>24&255]+"-"+mn[e&63|128]+mn[e>>8&255]+"-"+mn[e>>16&255]+mn[e>>24&255]+mn[n&255]+mn[n>>8&255]+mn[n>>16&255]+mn[n>>24&255]).toLowerCase()}function pe(u,t,e){return Math.max(t,Math.min(e,u))}function xr(u,t){return(u%t+t)%t}function pc(u,t,e,n,i){return n+(u-t)*(i-n)/(e-t)}function mc(u,t,e){return u!==t?(e-u)/(t-u):0}function ns(u,t,e){return(1-e)*u+e*t}function gc(u,t,e,n){return ns(u,t,1-Math.exp(-e*n))}function _c(u,t=1){return t-Math.abs(xr(u,t*2)-t)}function xc(u,t,e){return u<=t?0:u>=e?1:(u=(u-t)/(e-t),u*u*(3-2*u))}function vc(u,t,e){return u<=t?0:u>=e?1:(u=(u-t)/(e-t),u*u*u*(u*(u*6-15)+10))}function vr(u,t){return u+Math.floor(Math.random()*(t-u+1))}function Oa(u,t){return u+Math.random()*(t-u)}function yc(u){return u*(.5-Math.random())}function Us(u){u!==void 0&&(Fa=u);let t=Fa+=1831565813;return t=Math.imul(t^t>>>15,t|1),t^=t+Math.imul(t^t>>>7,t|61),((t^t>>>14)>>>0)/4294967296}function Ba(u){return u*_i}function Mc(u){return u*Di}function Sc(u){return(u&u-1)===0&&u!==0}function bc(u){return Math.pow(2,Math.ceil(Math.log(u)/Math.LN2))}function za(u){return Math.pow(2,Math.floor(Math.log(u)/Math.LN2))}function Tc(u,t,e,n,i){const s=Math.cos,r=Math.sin,a=s(e/2),l=r(e/2),c=s((t+n)/2),d=r((t+n)/2),m=s((t-n)/2),g=r((t-n)/2),_=s((n-t)/2),v=r((n-t)/2);switch(i){case"XYX":u.set(a*d,l*m,l*g,a*c);break;case"YZY":u.set(l*g,a*d,l*m,a*c);break;case"ZXZ":u.set(l*m,l*g,a*d,a*c);break;case"XZX":u.set(a*d,l*v,l*_,a*c);break;case"YXY":u.set(l*_,a*d,l*v,a*c);break;case"ZYZ":u.set(l*v,l*_,a*d,a*c);break;default:le("MathUtils: .setQuaternionFromProperEuler() encountered an unknown order: "+i)}}function gn(u,t){switch(t.constructor){case Float32Array:return u;case Uint32Array:return u/4294967295;case Uint16Array:return u/65535;case Uint8Array:return u/255;case Int32Array:return Math.max(u/2147483647,-1);case Int16Array:return Math.max(u/32767,-1);case Int8Array:return Math.max(u/127,-1);default:throw new Error("Invalid component type.")}}function ye(u,t){switch(t.constructor){case Float32Array:return u;case Uint32Array:return Math.round(u*4294967295);case Uint16Array:return Math.round(u*65535);case Uint8Array:return Math.round(u*255);case Int32Array:return Math.round(u*2147483647);case Int16Array:return Math.round(u*32767);case Int8Array:return Math.round(u*127);default:throw new Error("Invalid component type.")}}const au={DEG2RAD:_i,RAD2DEG:Di,generateUUID:An,clamp:pe,euclideanModulo:xr,mapLinear:pc,inverseLerp:mc,lerp:ns,damp:gc,pingpong:_c,smoothstep:xc,smootherstep:vc,randInt:vr,randFloat:Oa,randFloatSpread:yc,seededRandom:Us,degToRad:Ba,radToDeg:Mc,isPowerOfTwo:Sc,ceilPowerOfTwo:bc,floorPowerOfTwo:za,setQuaternionFromProperEuler:Tc,normalize:ye,denormalize:gn};class At{constructor(t=0,e=0){At.prototype.isVector2=!0,this.x=t,this.y=e}get width(){return this.x}set width(t){this.x=t}get height(){return this.y}set height(t){this.y=t}set(t,e){return this.x=t,this.y=e,this}setScalar(t){return this.x=t,this.y=t,this}setX(t){return this.x=t,this}setY(t){return this.y=t,this}setComponent(t,e){switch(t){case 0:this.x=e;break;case 1:this.y=e;break;default:throw new Error("index is out of range: "+t)}return this}getComponent(t){switch(t){case 0:return this.x;case 1:return this.y;default:throw new Error("index is out of range: "+t)}}clone(){return new this.constructor(this.x,this.y)}copy(t){return this.x=t.x,this.y=t.y,this}add(t){return this.x+=t.x,this.y+=t.y,this}addScalar(t){return this.x+=t,this.y+=t,this}addVectors(t,e){return this.x=t.x+e.x,this.y=t.y+e.y,this}addScaledVector(t,e){return this.x+=t.x*e,this.y+=t.y*e,this}sub(t){return this.x-=t.x,this.y-=t.y,this}subScalar(t){return this.x-=t,this.y-=t,this}subVectors(t,e){return this.x=t.x-e.x,this.y=t.y-e.y,this}multiply(t){return this.x*=t.x,this.y*=t.y,this}multiplyScalar(t){return this.x*=t,this.y*=t,this}divide(t){return this.x/=t.x,this.y/=t.y,this}divideScalar(t){return this.multiplyScalar(1/t)}applyMatrix3(t){const e=this.x,n=this.y,i=t.elements;return this.x=i[0]*e+i[3]*n+i[6],this.y=i[1]*e+i[4]*n+i[7],this}min(t){return this.x=Math.min(this.x,t.x),this.y=Math.min(this.y,t.y),this}max(t){return this.x=Math.max(this.x,t.x),this.y=Math.max(this.y,t.y),this}clamp(t,e){return this.x=pe(this.x,t.x,e.x),this.y=pe(this.y,t.y,e.y),this}clampScalar(t,e){return this.x=pe(this.x,t,e),this.y=pe(this.y,t,e),this}clampLength(t,e){const n=this.length();return this.divideScalar(n||1).multiplyScalar(pe(n,t,e))}floor(){return this.x=Math.floor(this.x),this.y=Math.floor(this.y),this}ceil(){return this.x=Math.ceil(this.x),this.y=Math.ceil(this.y),this}round(){return this.x=Math.round(this.x),this.y=Math.round(this.y),this}roundToZero(){return this.x=Math.trunc(this.x),this.y=Math.trunc(this.y),this}negate(){return this.x=-this.x,this.y=-this.y,this}dot(t){return this.x*t.x+this.y*t.y}cross(t){return this.x*t.y-this.y*t.x}lengthSq(){return this.x*this.x+this.y*this.y}length(){return Math.sqrt(this.x*this.x+this.y*this.y)}manhattanLength(){return Math.abs(this.x)+Math.abs(this.y)}normalize(){return this.divideScalar(this.length()||1)}angle(){return Math.atan2(-this.y,-this.x)+Math.PI}angleTo(t){const e=Math.sqrt(this.lengthSq()*t.lengthSq());if(e===0)return Math.PI/2;const n=this.dot(t)/e;return Math.acos(pe(n,-1,1))}distanceTo(t){return Math.sqrt(this.distanceToSquared(t))}distanceToSquared(t){const e=this.x-t.x,n=this.y-t.y;return e*e+n*n}manhattanDistanceTo(t){return Math.abs(this.x-t.x)+Math.abs(this.y-t.y)}setLength(t){return this.normalize().multiplyScalar(t)}lerp(t,e){return this.x+=(t.x-this.x)*e,this.y+=(t.y-this.y)*e,this}lerpVectors(t,e,n){return this.x=t.x+(e.x-t.x)*n,this.y=t.y+(e.y-t.y)*n,this}equals(t){return t.x===this.x&&t.y===this.y}fromArray(t,e=0){return this.x=t[e],this.y=t[e+1],this}toArray(t=[],e=0){return t[e]=this.x,t[e+1]=this.y,t}fromBufferAttribute(t,e){return this.x=t.getX(e),this.y=t.getY(e),this}rotateAround(t,e){const n=Math.cos(e),i=Math.sin(e),s=this.x-t.x,r=this.y-t.y;return this.x=s*n-r*i+t.x,this.y=s*i+r*n+t.y,this}random(){return this.x=Math.random(),this.y=Math.random(),this}*[Symbol.iterator](){yield this.x,yield this.y}}class un{constructor(t=0,e=0,n=0,i=1){this.isQuaternion=!0,this._x=t,this._y=e,this._z=n,this._w=i}static slerpFlat(t,e,n,i,s,r,a){let l=n[i+0],c=n[i+1],d=n[i+2],m=n[i+3],g=s[r+0],_=s[r+1],v=s[r+2],M=s[r+3];if(m!==M||l!==g||c!==_||d!==v){let C=l*g+c*_+d*v+m*M;C<0&&(g=-g,_=-_,v=-v,M=-M,C=-C);let w=1-a;if(C<.9995){const L=Math.acos(C),D=Math.sin(L);w=Math.sin(w*L)/D,a=Math.sin(a*L)/D,l=l*w+g*a,c=c*w+_*a,d=d*w+v*a,m=m*w+M*a}else{l=l*w+g*a,c=c*w+_*a,d=d*w+v*a,m=m*w+M*a;const L=1/Math.sqrt(l*l+c*c+d*d+m*m);l*=L,c*=L,d*=L,m*=L}}t[e]=l,t[e+1]=c,t[e+2]=d,t[e+3]=m}static multiplyQuaternionsFlat(t,e,n,i,s,r){const a=n[i],l=n[i+1],c=n[i+2],d=n[i+3],m=s[r],g=s[r+1],_=s[r+2],v=s[r+3];return t[e]=a*v+d*m+l*_-c*g,t[e+1]=l*v+d*g+c*m-a*_,t[e+2]=c*v+d*_+a*g-l*m,t[e+3]=d*v-a*m-l*g-c*_,t}get x(){return this._x}set x(t){this._x=t,this._onChangeCallback()}get y(){return this._y}set y(t){this._y=t,this._onChangeCallback()}get z(){return this._z}set z(t){this._z=t,this._onChangeCallback()}get w(){return this._w}set w(t){this._w=t,this._onChangeCallback()}set(t,e,n,i){return this._x=t,this._y=e,this._z=n,this._w=i,this._onChangeCallback(),this}clone(){return new this.constructor(this._x,this._y,this._z,this._w)}copy(t){return this._x=t.x,this._y=t.y,this._z=t.z,this._w=t.w,this._onChangeCallback(),this}setFromEuler(t,e=!0){const n=t._x,i=t._y,s=t._z,r=t._order,a=Math.cos,l=Math.sin,c=a(n/2),d=a(i/2),m=a(s/2),g=l(n/2),_=l(i/2),v=l(s/2);switch(r){case"XYZ":this._x=g*d*m+c*_*v,this._y=c*_*m-g*d*v,this._z=c*d*v+g*_*m,this._w=c*d*m-g*_*v;break;case"YXZ":this._x=g*d*m+c*_*v,this._y=c*_*m-g*d*v,this._z=c*d*v-g*_*m,this._w=c*d*m+g*_*v;break;case"ZXY":this._x=g*d*m-c*_*v,this._y=c*_*m+g*d*v,this._z=c*d*v+g*_*m,this._w=c*d*m-g*_*v;break;case"ZYX":this._x=g*d*m-c*_*v,this._y=c*_*m+g*d*v,this._z=c*d*v-g*_*m,this._w=c*d*m+g*_*v;break;case"YZX":this._x=g*d*m+c*_*v,this._y=c*_*m+g*d*v,this._z=c*d*v-g*_*m,this._w=c*d*m-g*_*v;break;case"XZY":this._x=g*d*m-c*_*v,this._y=c*_*m-g*d*v,this._z=c*d*v+g*_*m,this._w=c*d*m+g*_*v;break;default:le("Quaternion: .setFromEuler() encountered an unknown order: "+r)}return e===!0&&this._onChangeCallback(),this}setFromAxisAngle(t,e){const n=e/2,i=Math.sin(n);return this._x=t.x*i,this._y=t.y*i,this._z=t.z*i,this._w=Math.cos(n),this._onChangeCallback(),this}setFromRotationMatrix(t){const e=t.elements,n=e[0],i=e[4],s=e[8],r=e[1],a=e[5],l=e[9],c=e[2],d=e[6],m=e[10],g=n+a+m;if(g>0){const _=.5/Math.sqrt(g+1);this._w=.25/_,this._x=(d-l)*_,this._y=(s-c)*_,this._z=(r-i)*_}else if(n>a&&n>m){const _=2*Math.sqrt(1+n-a-m);this._w=(d-l)/_,this._x=.25*_,this._y=(i+r)/_,this._z=(s+c)/_}else if(a>m){const _=2*Math.sqrt(1+a-n-m);this._w=(s-c)/_,this._x=(i+r)/_,this._y=.25*_,this._z=(l+d)/_}else{const _=2*Math.sqrt(1+m-n-a);this._w=(r-i)/_,this._x=(s+c)/_,this._y=(l+d)/_,this._z=.25*_}return this._onChangeCallback(),this}setFromUnitVectors(t,e){let n=t.dot(e)+1;return n<1e-8?(n=0,Math.abs(t.x)>Math.abs(t.z)?(this._x=-t.y,this._y=t.x,this._z=0,this._w=n):(this._x=0,this._y=-t.z,this._z=t.y,this._w=n)):(this._x=t.y*e.z-t.z*e.y,this._y=t.z*e.x-t.x*e.z,this._z=t.x*e.y-t.y*e.x,this._w=n),this.normalize()}angleTo(t){return 2*Math.acos(Math.abs(pe(this.dot(t),-1,1)))}rotateTowards(t,e){const n=this.angleTo(t);if(n===0)return this;const i=Math.min(1,e/n);return this.slerp(t,i),this}identity(){return this.set(0,0,0,1)}invert(){return this.conjugate()}conjugate(){return this._x*=-1,this._y*=-1,this._z*=-1,this._onChangeCallback(),this}dot(t){return this._x*t._x+this._y*t._y+this._z*t._z+this._w*t._w}lengthSq(){return this._x*this._x+this._y*this._y+this._z*this._z+this._w*this._w}length(){return Math.sqrt(this._x*this._x+this._y*this._y+this._z*this._z+this._w*this._w)}normalize(){let t=this.length();return t===0?(this._x=0,this._y=0,this._z=0,this._w=1):(t=1/t,this._x=this._x*t,this._y=this._y*t,this._z=this._z*t,this._w=this._w*t),this._onChangeCallback(),this}multiply(t){return this.multiplyQuaternions(this,t)}premultiply(t){return this.multiplyQuaternions(t,this)}multiplyQuaternions(t,e){const n=t._x,i=t._y,s=t._z,r=t._w,a=e._x,l=e._y,c=e._z,d=e._w;return this._x=n*d+r*a+i*c-s*l,this._y=i*d+r*l+s*a-n*c,this._z=s*d+r*c+n*l-i*a,this._w=r*d-n*a-i*l-s*c,this._onChangeCallback(),this}slerp(t,e){let n=t._x,i=t._y,s=t._z,r=t._w,a=this.dot(t);a<0&&(n=-n,i=-i,s=-s,r=-r,a=-a);let l=1-e;if(a<.9995){const c=Math.acos(a),d=Math.sin(c);l=Math.sin(l*c)/d,e=Math.sin(e*c)/d,this._x=this._x*l+n*e,this._y=this._y*l+i*e,this._z=this._z*l+s*e,this._w=this._w*l+r*e,this._onChangeCallback()}else this._x=this._x*l+n*e,this._y=this._y*l+i*e,this._z=this._z*l+s*e,this._w=this._w*l+r*e,this.normalize();return this}slerpQuaternions(t,e,n){return this.copy(t).slerp(e,n)}random(){const t=2*Math.PI*Math.random(),e=2*Math.PI*Math.random(),n=Math.random(),i=Math.sqrt(1-n),s=Math.sqrt(n);return this.set(i*Math.sin(t),i*Math.cos(t),s*Math.sin(e),s*Math.cos(e))}equals(t){return t._x===this._x&&t._y===this._y&&t._z===this._z&&t._w===this._w}fromArray(t,e=0){return this._x=t[e],this._y=t[e+1],this._z=t[e+2],this._w=t[e+3],this._onChangeCallback(),this}toArray(t=[],e=0){return t[e]=this._x,t[e+1]=this._y,t[e+2]=this._z,t[e+3]=this._w,t}fromBufferAttribute(t,e){return this._x=t.getX(e),this._y=t.getY(e),this._z=t.getZ(e),this._w=t.getW(e),this._onChangeCallback(),this}toJSON(){return this.toArray()}_onChange(t){return this._onChangeCallback=t,this}_onChangeCallback(){}*[Symbol.iterator](){yield this._x,yield this._y,yield this._z,yield this._w}}class O{constructor(t=0,e=0,n=0){O.prototype.isVector3=!0,this.x=t,this.y=e,this.z=n}set(t,e,n){return n===void 0&&(n=this.z),this.x=t,this.y=e,this.z=n,this}setScalar(t){return this.x=t,this.y=t,this.z=t,this}setX(t){return this.x=t,this}setY(t){return this.y=t,this}setZ(t){return this.z=t,this}setComponent(t,e){switch(t){case 0:this.x=e;break;case 1:this.y=e;break;case 2:this.z=e;break;default:throw new Error("index is out of range: "+t)}return this}getComponent(t){switch(t){case 0:return this.x;case 1:return this.y;case 2:return this.z;default:throw new Error("index is out of range: "+t)}}clone(){return new this.constructor(this.x,this.y,this.z)}copy(t){return this.x=t.x,this.y=t.y,this.z=t.z,this}add(t){return this.x+=t.x,this.y+=t.y,this.z+=t.z,this}addScalar(t){return this.x+=t,this.y+=t,this.z+=t,this}addVectors(t,e){return this.x=t.x+e.x,this.y=t.y+e.y,this.z=t.z+e.z,this}addScaledVector(t,e){return this.x+=t.x*e,this.y+=t.y*e,this.z+=t.z*e,this}sub(t){return this.x-=t.x,this.y-=t.y,this.z-=t.z,this}subScalar(t){return this.x-=t,this.y-=t,this.z-=t,this}subVectors(t,e){return this.x=t.x-e.x,this.y=t.y-e.y,this.z=t.z-e.z,this}multiply(t){return this.x*=t.x,this.y*=t.y,this.z*=t.z,this}multiplyScalar(t){return this.x*=t,this.y*=t,this.z*=t,this}multiplyVectors(t,e){return this.x=t.x*e.x,this.y=t.y*e.y,this.z=t.z*e.z,this}applyEuler(t){return this.applyQuaternion(Va.setFromEuler(t))}applyAxisAngle(t,e){return this.applyQuaternion(Va.setFromAxisAngle(t,e))}applyMatrix3(t){const e=this.x,n=this.y,i=this.z,s=t.elements;return this.x=s[0]*e+s[3]*n+s[6]*i,this.y=s[1]*e+s[4]*n+s[7]*i,this.z=s[2]*e+s[5]*n+s[8]*i,this}applyNormalMatrix(t){return this.applyMatrix3(t).normalize()}applyMatrix4(t){const e=this.x,n=this.y,i=this.z,s=t.elements,r=1/(s[3]*e+s[7]*n+s[11]*i+s[15]);return this.x=(s[0]*e+s[4]*n+s[8]*i+s[12])*r,this.y=(s[1]*e+s[5]*n+s[9]*i+s[13])*r,this.z=(s[2]*e+s[6]*n+s[10]*i+s[14])*r,this}applyQuaternion(t){const e=this.x,n=this.y,i=this.z,s=t.x,r=t.y,a=t.z,l=t.w,c=2*(r*i-a*n),d=2*(a*e-s*i),m=2*(s*n-r*e);return this.x=e+l*c+r*m-a*d,this.y=n+l*d+a*c-s*m,this.z=i+l*m+s*d-r*c,this}project(t){return this.applyMatrix4(t.matrixWorldInverse).applyMatrix4(t.projectionMatrix)}unproject(t){return this.applyMatrix4(t.projectionMatrixInverse).applyMatrix4(t.matrixWorld)}transformDirection(t){const e=this.x,n=this.y,i=this.z,s=t.elements;return this.x=s[0]*e+s[4]*n+s[8]*i,this.y=s[1]*e+s[5]*n+s[9]*i,this.z=s[2]*e+s[6]*n+s[10]*i,this.normalize()}divide(t){return this.x/=t.x,this.y/=t.y,this.z/=t.z,this}divideScalar(t){return this.multiplyScalar(1/t)}min(t){return this.x=Math.min(this.x,t.x),this.y=Math.min(this.y,t.y),this.z=Math.min(this.z,t.z),this}max(t){return this.x=Math.max(this.x,t.x),this.y=Math.max(this.y,t.y),this.z=Math.max(this.z,t.z),this}clamp(t,e){return this.x=pe(this.x,t.x,e.x),this.y=pe(this.y,t.y,e.y),this.z=pe(this.z,t.z,e.z),this}clampScalar(t,e){return this.x=pe(this.x,t,e),this.y=pe(this.y,t,e),this.z=pe(this.z,t,e),this}clampLength(t,e){const n=this.length();return this.divideScalar(n||1).multiplyScalar(pe(n,t,e))}floor(){return this.x=Math.floor(this.x),this.y=Math.floor(this.y),this.z=Math.floor(this.z),this}ceil(){return this.x=Math.ceil(this.x),this.y=Math.ceil(this.y),this.z=Math.ceil(this.z),this}round(){return this.x=Math.round(this.x),this.y=Math.round(this.y),this.z=Math.round(this.z),this}roundToZero(){return this.x=Math.trunc(this.x),this.y=Math.trunc(this.y),this.z=Math.trunc(this.z),this}negate(){return this.x=-this.x,this.y=-this.y,this.z=-this.z,this}dot(t){return this.x*t.x+this.y*t.y+this.z*t.z}lengthSq(){return this.x*this.x+this.y*this.y+this.z*this.z}length(){return Math.sqrt(this.x*this.x+this.y*this.y+this.z*this.z)}manhattanLength(){return Math.abs(this.x)+Math.abs(this.y)+Math.abs(this.z)}normalize(){return this.divideScalar(this.length()||1)}setLength(t){return this.normalize().multiplyScalar(t)}lerp(t,e){return this.x+=(t.x-this.x)*e,this.y+=(t.y-this.y)*e,this.z+=(t.z-this.z)*e,this}lerpVectors(t,e,n){return this.x=t.x+(e.x-t.x)*n,this.y=t.y+(e.y-t.y)*n,this.z=t.z+(e.z-t.z)*n,this}cross(t){return this.crossVectors(this,t)}crossVectors(t,e){const n=t.x,i=t.y,s=t.z,r=e.x,a=e.y,l=e.z;return this.x=i*l-s*a,this.y=s*r-n*l,this.z=n*a-i*r,this}projectOnVector(t){const e=t.lengthSq();if(e===0)return this.set(0,0,0);const n=t.dot(this)/e;return this.copy(t).multiplyScalar(n)}projectOnPlane(t){return yr.copy(this).projectOnVector(t),this.sub(yr)}reflect(t){return this.sub(yr.copy(t).multiplyScalar(2*this.dot(t)))}angleTo(t){const e=Math.sqrt(this.lengthSq()*t.lengthSq());if(e===0)return Math.PI/2;const n=this.dot(t)/e;return Math.acos(pe(n,-1,1))}distanceTo(t){return Math.sqrt(this.distanceToSquared(t))}distanceToSquared(t){const e=this.x-t.x,n=this.y-t.y,i=this.z-t.z;return e*e+n*n+i*i}manhattanDistanceTo(t){return Math.abs(this.x-t.x)+Math.abs(this.y-t.y)+Math.abs(this.z-t.z)}setFromSpherical(t){return this.setFromSphericalCoords(t.radius,t.phi,t.theta)}setFromSphericalCoords(t,e,n){const i=Math.sin(e)*t;return this.x=i*Math.sin(n),this.y=Math.cos(e)*t,this.z=i*Math.cos(n),this}setFromCylindrical(t){return this.setFromCylindricalCoords(t.radius,t.theta,t.y)}setFromCylindricalCoords(t,e,n){return this.x=t*Math.sin(e),this.y=n,this.z=t*Math.cos(e),this}setFromMatrixPosition(t){const e=t.elements;return this.x=e[12],this.y=e[13],this.z=e[14],this}setFromMatrixScale(t){const e=this.setFromMatrixColumn(t,0).length(),n=this.setFromMatrixColumn(t,1).length(),i=this.setFromMatrixColumn(t,2).length();return this.x=e,this.y=n,this.z=i,this}setFromMatrixColumn(t,e){return this.fromArray(t.elements,e*4)}setFromMatrix3Column(t,e){return this.fromArray(t.elements,e*3)}setFromEuler(t){return this.x=t._x,this.y=t._y,this.z=t._z,this}setFromColor(t){return this.x=t.r,this.y=t.g,this.z=t.b,this}equals(t){return t.x===this.x&&t.y===this.y&&t.z===this.z}fromArray(t,e=0){return this.x=t[e],this.y=t[e+1],this.z=t[e+2],this}toArray(t=[],e=0){return t[e]=this.x,t[e+1]=this.y,t[e+2]=this.z,t}fromBufferAttribute(t,e){return this.x=t.getX(e),this.y=t.getY(e),this.z=t.getZ(e),this}random(){return this.x=Math.random(),this.y=Math.random(),this.z=Math.random(),this}randomDirection(){const t=Math.random()*Math.PI*2,e=Math.random()*2-1,n=Math.sqrt(1-e*e);return this.x=n*Math.cos(t),this.y=e,this.z=n*Math.sin(t),this}*[Symbol.iterator](){yield this.x,yield this.y,yield this.z}}const yr=new O,Va=new un;class Fn{constructor(t,e,n,i,s,r,a,l,c){Fn.prototype.isMatrix3=!0,this.elements=[1,0,0,0,1,0,0,0,1],t!==void 0&&this.set(t,e,n,i,s,r,a,l,c)}set(t,e,n,i,s,r,a,l,c){const d=this.elements;return d[0]=t,d[1]=i,d[2]=a,d[3]=e,d[4]=s,d[5]=l,d[6]=n,d[7]=r,d[8]=c,this}identity(){return this.set(1,0,0,0,1,0,0,0,1),this}copy(t){const e=this.elements,n=t.elements;return e[0]=n[0],e[1]=n[1],e[2]=n[2],e[3]=n[3],e[4]=n[4],e[5]=n[5],e[6]=n[6],e[7]=n[7],e[8]=n[8],this}extractBasis(t,e,n){return t.setFromMatrix3Column(this,0),e.setFromMatrix3Column(this,1),n.setFromMatrix3Column(this,2),this}setFromMatrix4(t){const e=t.elements;return this.set(e[0],e[4],e[8],e[1],e[5],e[9],e[2],e[6],e[10]),this}multiply(t){return this.multiplyMatrices(this,t)}premultiply(t){return this.multiplyMatrices(t,this)}multiplyMatrices(t,e){const n=t.elements,i=e.elements,s=this.elements,r=n[0],a=n[3],l=n[6],c=n[1],d=n[4],m=n[7],g=n[2],_=n[5],v=n[8],M=i[0],C=i[3],w=i[6],L=i[1],D=i[4],N=i[7],et=i[2],rt=i[5],vt=i[8];return s[0]=r*M+a*L+l*et,s[3]=r*C+a*D+l*rt,s[6]=r*w+a*N+l*vt,s[1]=c*M+d*L+m*et,s[4]=c*C+d*D+m*rt,s[7]=c*w+d*N+m*vt,s[2]=g*M+_*L+v*et,s[5]=g*C+_*D+v*rt,s[8]=g*w+_*N+v*vt,this}multiplyScalar(t){const e=this.elements;return e[0]*=t,e[3]*=t,e[6]*=t,e[1]*=t,e[4]*=t,e[7]*=t,e[2]*=t,e[5]*=t,e[8]*=t,this}determinant(){const t=this.elements,e=t[0],n=t[1],i=t[2],s=t[3],r=t[4],a=t[5],l=t[6],c=t[7],d=t[8];return e*r*d-e*a*c-n*s*d+n*a*l+i*s*c-i*r*l}invert(){const t=this.elements,e=t[0],n=t[1],i=t[2],s=t[3],r=t[4],a=t[5],l=t[6],c=t[7],d=t[8],m=d*r-a*c,g=a*l-d*s,_=c*s-r*l,v=e*m+n*g+i*_;if(v===0)return this.set(0,0,0,0,0,0,0,0,0);const M=1/v;return t[0]=m*M,t[1]=(i*c-d*n)*M,t[2]=(a*n-i*r)*M,t[3]=g*M,t[4]=(d*e-i*l)*M,t[5]=(i*s-a*e)*M,t[6]=_*M,t[7]=(n*l-c*e)*M,t[8]=(r*e-n*s)*M,this}transpose(){let t;const e=this.elements;return t=e[1],e[1]=e[3],e[3]=t,t=e[2],e[2]=e[6],e[6]=t,t=e[5],e[5]=e[7],e[7]=t,this}getNormalMatrix(t){return this.setFromMatrix4(t).invert().transpose()}transposeIntoArray(t){const e=this.elements;return t[0]=e[0],t[1]=e[3],t[2]=e[6],t[3]=e[1],t[4]=e[4],t[5]=e[7],t[6]=e[2],t[7]=e[5],t[8]=e[8],this}setUvTransform(t,e,n,i,s,r,a){const l=Math.cos(s),c=Math.sin(s);return this.set(n*l,n*c,-n*(l*r+c*a)+r+t,-i*c,i*l,-i*(-c*r+l*a)+a+e,0,0,1),this}scale(t,e){return this.premultiply(xi.makeScale(t,e)),this}rotate(t){return this.premultiply(xi.makeRotation(-t)),this}translate(t,e){return this.premultiply(xi.makeTranslation(t,e)),this}makeTranslation(t,e){return t.isVector2?this.set(1,0,t.x,0,1,t.y,0,0,1):this.set(1,0,t,0,1,e,0,0,1),this}makeRotation(t){const e=Math.cos(t),n=Math.sin(t);return this.set(e,-n,0,n,e,0,0,0,1),this}makeScale(t,e){return this.set(t,0,0,0,e,0,0,0,1),this}equals(t){const e=this.elements,n=t.elements;for(let i=0;i<9;i++)if(e[i]!==n[i])return!1;return!0}fromArray(t,e=0){for(let n=0;n<9;n++)this.elements[n]=t[n+e];return this}toArray(t=[],e=0){const n=this.elements;return t[e]=n[0],t[e+1]=n[1],t[e+2]=n[2],t[e+3]=n[3],t[e+4]=n[4],t[e+5]=n[5],t[e+6]=n[6],t[e+7]=n[7],t[e+8]=n[8],t}clone(){return new this.constructor().fromArray(this.elements)}}const xi=new Fn,Mr=new Fn().set(.4123908,.3575843,.1804808,.212639,.7151687,.0721923,.0193308,.1191948,.9505322),Sr=new Fn().set(3.2409699,-1.5373832,-.4986108,-.9692436,1.8759675,.0415551,.0556301,-.203977,1.0569715);function Ac(){const u={enabled:!0,workingColorSpace:Ji,spaces:{},convert:function(i,s,r){return this.enabled===!1||s===r||!s||!r||(this.spaces[s].transfer===Ds&&(i.r=Qn(i.r),i.g=Qn(i.g),i.b=Qn(i.b)),this.spaces[s].primaries!==this.spaces[r].primaries&&(i.applyMatrix3(this.spaces[s].toXYZ),i.applyMatrix3(this.spaces[r].fromXYZ)),this.spaces[r].transfer===Ds&&(i.r=Ui(i.r),i.g=Ui(i.g),i.b=Ui(i.b))),i},workingToColorSpace:function(i,s){return this.convert(i,this.workingColorSpace,s)},colorSpaceToWorking:function(i,s){return this.convert(i,s,this.workingColorSpace)},getPrimaries:function(i){return this.spaces[i].primaries},getTransfer:function(i){return i===$i?Ki:this.spaces[i].transfer},getToneMappingMode:function(i){return this.spaces[i].outputColorSpaceConfig.toneMappingMode||"standard"},getLuminanceCoefficients:function(i,s=this.workingColorSpace){return i.fromArray(this.spaces[s].luminanceCoefficients)},define:function(i){Object.assign(this.spaces,i)},_getMatrix:function(i,s,r){return i.copy(this.spaces[s].toXYZ).multiply(this.spaces[r].fromXYZ)},_getDrawingBufferColorSpace:function(i){return this.spaces[i].outputColorSpaceConfig.drawingBufferColorSpace},_getUnpackColorSpace:function(i=this.workingColorSpace){return this.spaces[i].workingColorSpaceConfig.unpackColorSpace},fromWorkingColorSpace:function(i,s){return _r("ColorManagement: .fromWorkingColorSpace() has been renamed to .workingToColorSpace()."),u.workingToColorSpace(i,s)},toWorkingColorSpace:function(i,s){return _r("ColorManagement: .toWorkingColorSpace() has been renamed to .colorSpaceToWorking()."),u.colorSpaceToWorking(i,s)}},t=[.64,.33,.3,.6,.15,.06],e=[.2126,.7152,.0722],n=[.3127,.329];return u.define({[Ji]:{primaries:t,whitePoint:n,transfer:Ki,toXYZ:Mr,fromXYZ:Sr,luminanceCoefficients:e,workingColorSpaceConfig:{unpackColorSpace:yn},outputColorSpaceConfig:{drawingBufferColorSpace:yn}},[yn]:{primaries:t,whitePoint:n,transfer:Ds,toXYZ:Mr,fromXYZ:Sr,luminanceCoefficients:e,outputColorSpaceConfig:{drawingBufferColorSpace:yn}}}),u}const _n=Ac();function Qn(u){return u<.04045?u*.0773993808:Math.pow(u*.9478672986+.0521327014,2.4)}function Ui(u){return u<.0031308?u*12.92:1.055*Math.pow(u,.41666)-.055}let Ni;class ka{static getDataURL(t,e="image/png"){if(/^data:/i.test(t.src)||typeof HTMLCanvasElement>"u")return t.src;let n;if(t instanceof HTMLCanvasElement)n=t;else{Ni===void 0&&(Ni=es("canvas")),Ni.width=t.width,Ni.height=t.height;const i=Ni.getContext("2d");t instanceof ImageData?i.putImageData(t,0,0):i.drawImage(t,0,0,t.width,t.height),n=Ni}return n.toDataURL(e)}static sRGBToLinear(t){if(typeof HTMLImageElement<"u"&&t instanceof HTMLImageElement||typeof HTMLCanvasElement<"u"&&t instanceof HTMLCanvasElement||typeof ImageBitmap<"u"&&t instanceof ImageBitmap){const e=es("canvas");e.width=t.width,e.height=t.height;const n=e.getContext("2d");n.drawImage(t,0,0,t.width,t.height);const i=n.getImageData(0,0,t.width,t.height),s=i.data;for(let r=0;r<s.length;r++)s[r]=Qn(s[r]/255)*255;return n.putImageData(i,0,0),e}else if(t.data){const e=t.data.slice(0);for(let n=0;n<e.length;n++)e instanceof Uint8Array||e instanceof Uint8ClampedArray?e[n]=Math.floor(Qn(e[n]/255)*255):e[n]=Qn(e[n]);return{data:e,width:t.width,height:t.height}}else return le("ImageUtils.sRGBToLinear(): Unsupported image type. No color space conversion applied."),t}}let Ec=0;class oi{constructor(t=null){this.isSource=!0,Object.defineProperty(this,"id",{value:Ec++}),this.uuid=An(),this.data=t,this.dataReady=!0,this.version=0}getSize(t){const e=this.data;return typeof HTMLVideoElement<"u"&&e instanceof HTMLVideoElement?t.set(e.videoWidth,e.videoHeight,0):typeof VideoFrame<"u"&&e instanceof VideoFrame?t.set(e.displayHeight,e.displayWidth,0):e!==null?t.set(e.width,e.height,e.depth||0):t.set(0,0,0),t}set needsUpdate(t){t===!0&&this.version++}toJSON(t){const e=t===void 0||typeof t=="string";if(!e&&t.images[this.uuid]!==void 0)return t.images[this.uuid];const n={uuid:this.uuid,url:""},i=this.data;if(i!==null){let s;if(Array.isArray(i)){s=[];for(let r=0,a=i.length;r<a;r++)i[r].isDataTexture?s.push(br(i[r].image)):s.push(br(i[r]))}else s=br(i);n.url=s}return e||(t.images[this.uuid]=n),n}}function br(u){return typeof HTMLImageElement<"u"&&u instanceof HTMLImageElement||typeof HTMLCanvasElement<"u"&&u instanceof HTMLCanvasElement||typeof ImageBitmap<"u"&&u instanceof ImageBitmap?ka.getDataURL(u):u.data?{data:Array.from(u.data),width:u.width,height:u.height,type:u.data.constructor.name}:(le("Texture: Unable to serialize Texture."),{})}let wc=0;const Tr=new O;class Ye extends Kn{constructor(t=Ye.DEFAULT_IMAGE,e=Ye.DEFAULT_MAPPING,n=Ln,i=Ln,s=Cn,r=Rs,a=Ci,l=Is,c=Ye.DEFAULT_ANISOTROPY,d=$i){super(),this.isTexture=!0,Object.defineProperty(this,"id",{value:wc++}),this.uuid=An(),this.name="",this.source=new oi(t),this.mipmaps=[],this.mapping=e,this.channel=0,this.wrapS=n,this.wrapT=i,this.magFilter=s,this.minFilter=r,this.anisotropy=c,this.format=a,this.internalFormat=null,this.type=l,this.offset=new At(0,0),this.repeat=new At(1,1),this.center=new At(0,0),this.rotation=0,this.matrixAutoUpdate=!0,this.matrix=new Fn,this.generateMipmaps=!0,this.premultiplyAlpha=!1,this.flipY=!0,this.unpackAlignment=4,this.colorSpace=d,this.userData={},this.updateRanges=[],this.version=0,this.onUpdate=null,this.renderTarget=null,this.isRenderTargetTexture=!1,this.isArrayTexture=!!(t&&t.depth&&t.depth>1),this.pmremVersion=0}get width(){return this.source.getSize(Tr).x}get height(){return this.source.getSize(Tr).y}get depth(){return this.source.getSize(Tr).z}get image(){return this.source.data}set image(t=null){this.source.data=t}updateMatrix(){this.matrix.setUvTransform(this.offset.x,this.offset.y,this.repeat.x,this.repeat.y,this.rotation,this.center.x,this.center.y)}addUpdateRange(t,e){this.updateRanges.push({start:t,count:e})}clearUpdateRanges(){this.updateRanges.length=0}clone(){return new this.constructor().copy(this)}copy(t){return this.name=t.name,this.source=t.source,this.mipmaps=t.mipmaps.slice(0),this.mapping=t.mapping,this.channel=t.channel,this.wrapS=t.wrapS,this.wrapT=t.wrapT,this.magFilter=t.magFilter,this.minFilter=t.minFilter,this.anisotropy=t.anisotropy,this.format=t.format,this.internalFormat=t.internalFormat,this.type=t.type,this.offset.copy(t.offset),this.repeat.copy(t.repeat),this.center.copy(t.center),this.rotation=t.rotation,this.matrixAutoUpdate=t.matrixAutoUpdate,this.matrix.copy(t.matrix),this.generateMipmaps=t.generateMipmaps,this.premultiplyAlpha=t.premultiplyAlpha,this.flipY=t.flipY,this.unpackAlignment=t.unpackAlignment,this.colorSpace=t.colorSpace,this.renderTarget=t.renderTarget,this.isRenderTargetTexture=t.isRenderTargetTexture,this.isArrayTexture=t.isArrayTexture,this.userData=JSON.parse(JSON.stringify(t.userData)),this.needsUpdate=!0,this}setValues(t){for(const e in t){const n=t[e];if(n===void 0){le(`Texture.setValues(): parameter '${e}' has value of undefined.`);continue}const i=this[e];if(i===void 0){le(`Texture.setValues(): property '${e}' does not exist.`);continue}i&&n&&i.isVector2&&n.isVector2||i&&n&&i.isVector3&&n.isVector3||i&&n&&i.isMatrix3&&n.isMatrix3?i.copy(n):this[e]=n}}toJSON(t){const e=t===void 0||typeof t=="string";if(!e&&t.textures[this.uuid]!==void 0)return t.textures[this.uuid];const n={metadata:{version:4.7,type:"Texture",generator:"Texture.toJSON"},uuid:this.uuid,name:this.name,image:this.source.toJSON(t).uuid,mapping:this.mapping,channel:this.channel,repeat:[this.repeat.x,this.repeat.y],offset:[this.offset.x,this.offset.y],center:[this.center.x,this.center.y],rotation:this.rotation,wrap:[this.wrapS,this.wrapT],format:this.format,internalFormat:this.internalFormat,type:this.type,colorSpace:this.colorSpace,minFilter:this.minFilter,magFilter:this.magFilter,anisotropy:this.anisotropy,flipY:this.flipY,generateMipmaps:this.generateMipmaps,premultiplyAlpha:this.premultiplyAlpha,unpackAlignment:this.unpackAlignment};return Object.keys(this.userData).length>0&&(n.userData=this.userData),e||(t.textures[this.uuid]=n),n}dispose(){this.dispatchEvent({type:"dispose"})}transformUv(t){if(this.mapping!==lr)return t;if(t.applyMatrix3(this.matrix),t.x<0||t.x>1)switch(this.wrapS){case ws:t.x=t.x-Math.floor(t.x);break;case Ln:t.x=t.x<0?0:1;break;case Cs:Math.abs(Math.floor(t.x)%2)===1?t.x=Math.ceil(t.x)-t.x:t.x=t.x-Math.floor(t.x);break}if(t.y<0||t.y>1)switch(this.wrapT){case ws:t.y=t.y-Math.floor(t.y);break;case Ln:t.y=t.y<0?0:1;break;case Cs:Math.abs(Math.floor(t.y)%2)===1?t.y=Math.ceil(t.y)-t.y:t.y=t.y-Math.floor(t.y);break}return this.flipY&&(t.y=1-t.y),t}set needsUpdate(t){t===!0&&(this.version++,this.source.needsUpdate=!0)}set needsPMREMUpdate(t){t===!0&&this.pmremVersion++}}Ye.DEFAULT_IMAGE=null,Ye.DEFAULT_MAPPING=lr,Ye.DEFAULT_ANISOTROPY=1;class En{constructor(t=0,e=0,n=0,i=1){En.prototype.isVector4=!0,this.x=t,this.y=e,this.z=n,this.w=i}get width(){return this.z}set width(t){this.z=t}get height(){return this.w}set height(t){this.w=t}set(t,e,n,i){return this.x=t,this.y=e,this.z=n,this.w=i,this}setScalar(t){return this.x=t,this.y=t,this.z=t,this.w=t,this}setX(t){return this.x=t,this}setY(t){return this.y=t,this}setZ(t){return this.z=t,this}setW(t){return this.w=t,this}setComponent(t,e){switch(t){case 0:this.x=e;break;case 1:this.y=e;break;case 2:this.z=e;break;case 3:this.w=e;break;default:throw new Error("index is out of range: "+t)}return this}getComponent(t){switch(t){case 0:return this.x;case 1:return this.y;case 2:return this.z;case 3:return this.w;default:throw new Error("index is out of range: "+t)}}clone(){return new this.constructor(this.x,this.y,this.z,this.w)}copy(t){return this.x=t.x,this.y=t.y,this.z=t.z,this.w=t.w!==void 0?t.w:1,this}add(t){return this.x+=t.x,this.y+=t.y,this.z+=t.z,this.w+=t.w,this}addScalar(t){return this.x+=t,this.y+=t,this.z+=t,this.w+=t,this}addVectors(t,e){return this.x=t.x+e.x,this.y=t.y+e.y,this.z=t.z+e.z,this.w=t.w+e.w,this}addScaledVector(t,e){return this.x+=t.x*e,this.y+=t.y*e,this.z+=t.z*e,this.w+=t.w*e,this}sub(t){return this.x-=t.x,this.y-=t.y,this.z-=t.z,this.w-=t.w,this}subScalar(t){return this.x-=t,this.y-=t,this.z-=t,this.w-=t,this}subVectors(t,e){return this.x=t.x-e.x,this.y=t.y-e.y,this.z=t.z-e.z,this.w=t.w-e.w,this}multiply(t){return this.x*=t.x,this.y*=t.y,this.z*=t.z,this.w*=t.w,this}multiplyScalar(t){return this.x*=t,this.y*=t,this.z*=t,this.w*=t,this}applyMatrix4(t){const e=this.x,n=this.y,i=this.z,s=this.w,r=t.elements;return this.x=r[0]*e+r[4]*n+r[8]*i+r[12]*s,this.y=r[1]*e+r[5]*n+r[9]*i+r[13]*s,this.z=r[2]*e+r[6]*n+r[10]*i+r[14]*s,this.w=r[3]*e+r[7]*n+r[11]*i+r[15]*s,this}divide(t){return this.x/=t.x,this.y/=t.y,this.z/=t.z,this.w/=t.w,this}divideScalar(t){return this.multiplyScalar(1/t)}setAxisAngleFromQuaternion(t){this.w=2*Math.acos(t.w);const e=Math.sqrt(1-t.w*t.w);return e<1e-4?(this.x=1,this.y=0,this.z=0):(this.x=t.x/e,this.y=t.y/e,this.z=t.z/e),this}setAxisAngleFromRotationMatrix(t){let e,n,i,s;const l=t.elements,c=l[0],d=l[4],m=l[8],g=l[1],_=l[5],v=l[9],M=l[2],C=l[6],w=l[10];if(Math.abs(d-g)<.01&&Math.abs(m-M)<.01&&Math.abs(v-C)<.01){if(Math.abs(d+g)<.1&&Math.abs(m+M)<.1&&Math.abs(v+C)<.1&&Math.abs(c+_+w-3)<.1)return this.set(1,0,0,0),this;e=Math.PI;const D=(c+1)/2,N=(_+1)/2,et=(w+1)/2,rt=(d+g)/4,vt=(m+M)/4,at=(v+C)/4;return D>N&&D>et?D<.01?(n=0,i=.707106781,s=.707106781):(n=Math.sqrt(D),i=rt/n,s=vt/n):N>et?N<.01?(n=.707106781,i=0,s=.707106781):(i=Math.sqrt(N),n=rt/i,s=at/i):et<.01?(n=.707106781,i=.707106781,s=0):(s=Math.sqrt(et),n=vt/s,i=at/s),this.set(n,i,s,e),this}let L=Math.sqrt((C-v)*(C-v)+(m-M)*(m-M)+(g-d)*(g-d));return Math.abs(L)<.001&&(L=1),this.x=(C-v)/L,this.y=(m-M)/L,this.z=(g-d)/L,this.w=Math.acos((c+_+w-1)/2),this}setFromMatrixPosition(t){const e=t.elements;return this.x=e[12],this.y=e[13],this.z=e[14],this.w=e[15],this}min(t){return this.x=Math.min(this.x,t.x),this.y=Math.min(this.y,t.y),this.z=Math.min(this.z,t.z),this.w=Math.min(this.w,t.w),this}max(t){return this.x=Math.max(this.x,t.x),this.y=Math.max(this.y,t.y),this.z=Math.max(this.z,t.z),this.w=Math.max(this.w,t.w),this}clamp(t,e){return this.x=pe(this.x,t.x,e.x),this.y=pe(this.y,t.y,e.y),this.z=pe(this.z,t.z,e.z),this.w=pe(this.w,t.w,e.w),this}clampScalar(t,e){return this.x=pe(this.x,t,e),this.y=pe(this.y,t,e),this.z=pe(this.z,t,e),this.w=pe(this.w,t,e),this}clampLength(t,e){const n=this.length();return this.divideScalar(n||1).multiplyScalar(pe(n,t,e))}floor(){return this.x=Math.floor(this.x),this.y=Math.floor(this.y),this.z=Math.floor(this.z),this.w=Math.floor(this.w),this}ceil(){return this.x=Math.ceil(this.x),this.y=Math.ceil(this.y),this.z=Math.ceil(this.z),this.w=Math.ceil(this.w),this}round(){return this.x=Math.round(this.x),this.y=Math.round(this.y),this.z=Math.round(this.z),this.w=Math.round(this.w),this}roundToZero(){return this.x=Math.trunc(this.x),this.y=Math.trunc(this.y),this.z=Math.trunc(this.z),this.w=Math.trunc(this.w),this}negate(){return this.x=-this.x,this.y=-this.y,this.z=-this.z,this.w=-this.w,this}dot(t){return this.x*t.x+this.y*t.y+this.z*t.z+this.w*t.w}lengthSq(){return this.x*this.x+this.y*this.y+this.z*this.z+this.w*this.w}length(){return Math.sqrt(this.x*this.x+this.y*this.y+this.z*this.z+this.w*this.w)}manhattanLength(){return Math.abs(this.x)+Math.abs(this.y)+Math.abs(this.z)+Math.abs(this.w)}normalize(){return this.divideScalar(this.length()||1)}setLength(t){return this.normalize().multiplyScalar(t)}lerp(t,e){return this.x+=(t.x-this.x)*e,this.y+=(t.y-this.y)*e,this.z+=(t.z-this.z)*e,this.w+=(t.w-this.w)*e,this}lerpVectors(t,e,n){return this.x=t.x+(e.x-t.x)*n,this.y=t.y+(e.y-t.y)*n,this.z=t.z+(e.z-t.z)*n,this.w=t.w+(e.w-t.w)*n,this}equals(t){return t.x===this.x&&t.y===this.y&&t.z===this.z&&t.w===this.w}fromArray(t,e=0){return this.x=t[e],this.y=t[e+1],this.z=t[e+2],this.w=t[e+3],this}toArray(t=[],e=0){return t[e]=this.x,t[e+1]=this.y,t[e+2]=this.z,t[e+3]=this.w,t}fromBufferAttribute(t,e){return this.x=t.getX(e),this.y=t.getY(e),this.z=t.getZ(e),this.w=t.getW(e),this}random(){return this.x=Math.random(),this.y=Math.random(),this.z=Math.random(),this.w=Math.random(),this}*[Symbol.iterator](){yield this.x,yield this.y,yield this.z,yield this.w}}class Ga extends Kn{constructor(t=1,e=1,n={}){super(),n=Object.assign({generateMipmaps:!1,internalFormat:null,minFilter:Cn,depthBuffer:!0,stencilBuffer:!1,resolveDepthBuffer:!0,resolveStencilBuffer:!0,depthTexture:null,samples:0,count:1,depth:1,multiview:!1},n),this.isRenderTarget=!0,this.width=t,this.height=e,this.depth=n.depth,this.scissor=new En(0,0,t,e),this.scissorTest=!1,this.viewport=new En(0,0,t,e),this.textures=[];const i={width:t,height:e,depth:n.depth},s=new Ye(i),r=n.count;for(let a=0;a<r;a++)this.textures[a]=s.clone(),this.textures[a].isRenderTargetTexture=!0,this.textures[a].renderTarget=this;this._setTextureOptions(n),this.depthBuffer=n.depthBuffer,this.stencilBuffer=n.stencilBuffer,this.resolveDepthBuffer=n.resolveDepthBuffer,this.resolveStencilBuffer=n.resolveStencilBuffer,this._depthTexture=null,this.depthTexture=n.depthTexture,this.samples=n.samples,this.multiview=n.multiview}_setTextureOptions(t={}){const e={minFilter:Cn,generateMipmaps:!1,flipY:!1,internalFormat:null};t.mapping!==void 0&&(e.mapping=t.mapping),t.wrapS!==void 0&&(e.wrapS=t.wrapS),t.wrapT!==void 0&&(e.wrapT=t.wrapT),t.wrapR!==void 0&&(e.wrapR=t.wrapR),t.magFilter!==void 0&&(e.magFilter=t.magFilter),t.minFilter!==void 0&&(e.minFilter=t.minFilter),t.format!==void 0&&(e.format=t.format),t.type!==void 0&&(e.type=t.type),t.anisotropy!==void 0&&(e.anisotropy=t.anisotropy),t.colorSpace!==void 0&&(e.colorSpace=t.colorSpace),t.flipY!==void 0&&(e.flipY=t.flipY),t.generateMipmaps!==void 0&&(e.generateMipmaps=t.generateMipmaps),t.internalFormat!==void 0&&(e.internalFormat=t.internalFormat);for(let n=0;n<this.textures.length;n++)this.textures[n].setValues(e)}get texture(){return this.textures[0]}set texture(t){this.textures[0]=t}set depthTexture(t){this._depthTexture!==null&&(this._depthTexture.renderTarget=null),t!==null&&(t.renderTarget=this),this._depthTexture=t}get depthTexture(){return this._depthTexture}setSize(t,e,n=1){if(this.width!==t||this.height!==e||this.depth!==n){this.width=t,this.height=e,this.depth=n;for(let i=0,s=this.textures.length;i<s;i++)this.textures[i].image.width=t,this.textures[i].image.height=e,this.textures[i].image.depth=n,this.textures[i].isData3DTexture!==!0&&(this.textures[i].isArrayTexture=this.textures[i].image.depth>1);this.dispose()}this.viewport.set(0,0,t,e),this.scissor.set(0,0,t,e)}clone(){return new this.constructor().copy(this)}copy(t){this.width=t.width,this.height=t.height,this.depth=t.depth,this.scissor.copy(t.scissor),this.scissorTest=t.scissorTest,this.viewport.copy(t.viewport),this.textures.length=0;for(let e=0,n=t.textures.length;e<n;e++){this.textures[e]=t.textures[e].clone(),this.textures[e].isRenderTargetTexture=!0,this.textures[e].renderTarget=this;const i=Object.assign({},t.textures[e].image);this.textures[e].source=new oi(i)}return this.depthBuffer=t.depthBuffer,this.stencilBuffer=t.stencilBuffer,this.resolveDepthBuffer=t.resolveDepthBuffer,this.resolveStencilBuffer=t.resolveStencilBuffer,t.depthTexture!==null&&(this.depthTexture=t.depthTexture.clone()),this.samples=t.samples,this}dispose(){this.dispatchEvent({type:"dispose"})}}class Ar extends Ga{constructor(t=1,e=1,n={}){super(t,e,n),this.isWebGLRenderTarget=!0}}class Ha extends Ye{constructor(t=null,e=1,n=1,i=1){super(null),this.isDataArrayTexture=!0,this.image={data:t,width:e,height:n,depth:i},this.magFilter=Mn,this.minFilter=Mn,this.wrapR=Ln,this.generateMipmaps=!1,this.flipY=!1,this.unpackAlignment=1,this.layerUpdates=new Set}addLayerUpdate(t){this.layerUpdates.add(t)}clearLayerUpdates(){this.layerUpdates.clear()}}class ou extends Ar{constructor(t=1,e=1,n=1,i={}){super(t,e,i),this.isWebGLArrayRenderTarget=!0,this.depth=n,this.texture=new Ha(null,t,e,n),this._setTextureOptions(i),this.texture.isRenderTargetTexture=!0}}class Er extends Ye{constructor(t=null,e=1,n=1,i=1){super(null),this.isData3DTexture=!0,this.image={data:t,width:e,height:n,depth:i},this.magFilter=Mn,this.minFilter=Mn,this.wrapR=Ln,this.generateMipmaps=!1,this.flipY=!1,this.unpackAlignment=1}}class lu extends Ar{constructor(t=1,e=1,n=1,i={}){super(t,e,i),this.isWebGL3DRenderTarget=!0,this.depth=n,this.texture=new Er(null,t,e,n),this._setTextureOptions(i),this.texture.isRenderTargetTexture=!0}}class Me{constructor(t,e,n,i,s,r,a,l,c,d,m,g,_,v,M,C){Me.prototype.isMatrix4=!0,this.elements=[1,0,0,0,0,1,0,0,0,0,1,0,0,0,0,1],t!==void 0&&this.set(t,e,n,i,s,r,a,l,c,d,m,g,_,v,M,C)}set(t,e,n,i,s,r,a,l,c,d,m,g,_,v,M,C){const w=this.elements;return w[0]=t,w[4]=e,w[8]=n,w[12]=i,w[1]=s,w[5]=r,w[9]=a,w[13]=l,w[2]=c,w[6]=d,w[10]=m,w[14]=g,w[3]=_,w[7]=v,w[11]=M,w[15]=C,this}identity(){return this.set(1,0,0,0,0,1,0,0,0,0,1,0,0,0,0,1),this}clone(){return new Me().fromArray(this.elements)}copy(t){const e=this.elements,n=t.elements;return e[0]=n[0],e[1]=n[1],e[2]=n[2],e[3]=n[3],e[4]=n[4],e[5]=n[5],e[6]=n[6],e[7]=n[7],e[8]=n[8],e[9]=n[9],e[10]=n[10],e[11]=n[11],e[12]=n[12],e[13]=n[13],e[14]=n[14],e[15]=n[15],this}copyPosition(t){const e=this.elements,n=t.elements;return e[12]=n[12],e[13]=n[13],e[14]=n[14],this}setFromMatrix3(t){const e=t.elements;return this.set(e[0],e[3],e[6],0,e[1],e[4],e[7],0,e[2],e[5],e[8],0,0,0,0,1),this}extractBasis(t,e,n){return this.determinant()===0?(t.set(1,0,0),e.set(0,1,0),n.set(0,0,1),this):(t.setFromMatrixColumn(this,0),e.setFromMatrixColumn(this,1),n.setFromMatrixColumn(this,2),this)}makeBasis(t,e,n){return this.set(t.x,e.x,n.x,0,t.y,e.y,n.y,0,t.z,e.z,n.z,0,0,0,0,1),this}extractRotation(t){if(t.determinant()===0)return this.identity();const e=this.elements,n=t.elements,i=1/Fi.setFromMatrixColumn(t,0).length(),s=1/Fi.setFromMatrixColumn(t,1).length(),r=1/Fi.setFromMatrixColumn(t,2).length();return e[0]=n[0]*i,e[1]=n[1]*i,e[2]=n[2]*i,e[3]=0,e[4]=n[4]*s,e[5]=n[5]*s,e[6]=n[6]*s,e[7]=0,e[8]=n[8]*r,e[9]=n[9]*r,e[10]=n[10]*r,e[11]=0,e[12]=0,e[13]=0,e[14]=0,e[15]=1,this}makeRotationFromEuler(t){const e=this.elements,n=t.x,i=t.y,s=t.z,r=Math.cos(n),a=Math.sin(n),l=Math.cos(i),c=Math.sin(i),d=Math.cos(s),m=Math.sin(s);if(t.order==="XYZ"){const g=r*d,_=r*m,v=a*d,M=a*m;e[0]=l*d,e[4]=-l*m,e[8]=c,e[1]=_+v*c,e[5]=g-M*c,e[9]=-a*l,e[2]=M-g*c,e[6]=v+_*c,e[10]=r*l}else if(t.order==="YXZ"){const g=l*d,_=l*m,v=c*d,M=c*m;e[0]=g+M*a,e[4]=v*a-_,e[8]=r*c,e[1]=r*m,e[5]=r*d,e[9]=-a,e[2]=_*a-v,e[6]=M+g*a,e[10]=r*l}else if(t.order==="ZXY"){const g=l*d,_=l*m,v=c*d,M=c*m;e[0]=g-M*a,e[4]=-r*m,e[8]=v+_*a,e[1]=_+v*a,e[5]=r*d,e[9]=M-g*a,e[2]=-r*c,e[6]=a,e[10]=r*l}else if(t.order==="ZYX"){const g=r*d,_=r*m,v=a*d,M=a*m;e[0]=l*d,e[4]=v*c-_,e[8]=g*c+M,e[1]=l*m,e[5]=M*c+g,e[9]=_*c-v,e[2]=-c,e[6]=a*l,e[10]=r*l}else if(t.order==="YZX"){const g=r*l,_=r*c,v=a*l,M=a*c;e[0]=l*d,e[4]=M-g*m,e[8]=v*m+_,e[1]=m,e[5]=r*d,e[9]=-a*d,e[2]=-c*d,e[6]=_*m+v,e[10]=g-M*m}else if(t.order==="XZY"){const g=r*l,_=r*c,v=a*l,M=a*c;e[0]=l*d,e[4]=-m,e[8]=c*d,e[1]=g*m+M,e[5]=r*d,e[9]=_*m-v,e[2]=v*m-_,e[6]=a*d,e[10]=M*m+g}return e[3]=0,e[7]=0,e[11]=0,e[12]=0,e[13]=0,e[14]=0,e[15]=1,this}makeRotationFromQuaternion(t){return this.compose(Cc,t,Rc)}lookAt(t,e,n){const i=this.elements;return wn.subVectors(t,e),wn.lengthSq()===0&&(wn.z=1),wn.normalize(),jn.crossVectors(n,wn),jn.lengthSq()===0&&(Math.abs(n.z)===1?wn.x+=1e-4:wn.z+=1e-4,wn.normalize(),jn.crossVectors(n,wn)),jn.normalize(),is.crossVectors(wn,jn),i[0]=jn.x,i[4]=is.x,i[8]=wn.x,i[1]=jn.y,i[5]=is.y,i[9]=wn.y,i[2]=jn.z,i[6]=is.z,i[10]=wn.z,this}multiply(t){return this.multiplyMatrices(this,t)}premultiply(t){return this.multiplyMatrices(t,this)}multiplyMatrices(t,e){const n=t.elements,i=e.elements,s=this.elements,r=n[0],a=n[4],l=n[8],c=n[12],d=n[1],m=n[5],g=n[9],_=n[13],v=n[2],M=n[6],C=n[10],w=n[14],L=n[3],D=n[7],N=n[11],et=n[15],rt=i[0],vt=i[4],at=i[8],Mt=i[12],St=i[1],Gt=i[5],he=i[9],ge=i[13],Ne=i[2],xe=i[6],nn=i[10],cn=i[14],ii=i[3],sn=i[7],hn=i[11],rn=i[15];return s[0]=r*rt+a*St+l*Ne+c*ii,s[4]=r*vt+a*Gt+l*xe+c*sn,s[8]=r*at+a*he+l*nn+c*hn,s[12]=r*Mt+a*ge+l*cn+c*rn,s[1]=d*rt+m*St+g*Ne+_*ii,s[5]=d*vt+m*Gt+g*xe+_*sn,s[9]=d*at+m*he+g*nn+_*hn,s[13]=d*Mt+m*ge+g*cn+_*rn,s[2]=v*rt+M*St+C*Ne+w*ii,s[6]=v*vt+M*Gt+C*xe+w*sn,s[10]=v*at+M*he+C*nn+w*hn,s[14]=v*Mt+M*ge+C*cn+w*rn,s[3]=L*rt+D*St+N*Ne+et*ii,s[7]=L*vt+D*Gt+N*xe+et*sn,s[11]=L*at+D*he+N*nn+et*hn,s[15]=L*Mt+D*ge+N*cn+et*rn,this}multiplyScalar(t){const e=this.elements;return e[0]*=t,e[4]*=t,e[8]*=t,e[12]*=t,e[1]*=t,e[5]*=t,e[9]*=t,e[13]*=t,e[2]*=t,e[6]*=t,e[10]*=t,e[14]*=t,e[3]*=t,e[7]*=t,e[11]*=t,e[15]*=t,this}determinant(){const t=this.elements,e=t[0],n=t[4],i=t[8],s=t[12],r=t[1],a=t[5],l=t[9],c=t[13],d=t[2],m=t[6],g=t[10],_=t[14],v=t[3],M=t[7],C=t[11],w=t[15],L=l*_-c*g,D=a*_-c*m,N=a*g-l*m,et=r*_-c*d,rt=r*g-l*d,vt=r*m-a*d;return e*(M*L-C*D+w*N)-n*(v*L-C*et+w*rt)+i*(v*D-M*et+w*vt)-s*(v*N-M*rt+C*vt)}transpose(){const t=this.elements;let e;return e=t[1],t[1]=t[4],t[4]=e,e=t[2],t[2]=t[8],t[8]=e,e=t[6],t[6]=t[9],t[9]=e,e=t[3],t[3]=t[12],t[12]=e,e=t[7],t[7]=t[13],t[13]=e,e=t[11],t[11]=t[14],t[14]=e,this}setPosition(t,e,n){const i=this.elements;return t.isVector3?(i[12]=t.x,i[13]=t.y,i[14]=t.z):(i[12]=t,i[13]=e,i[14]=n),this}invert(){const t=this.elements,e=t[0],n=t[1],i=t[2],s=t[3],r=t[4],a=t[5],l=t[6],c=t[7],d=t[8],m=t[9],g=t[10],_=t[11],v=t[12],M=t[13],C=t[14],w=t[15],L=e*a-n*r,D=e*l-i*r,N=e*c-s*r,et=n*l-i*a,rt=n*c-s*a,vt=i*c-s*l,at=d*M-m*v,Mt=d*C-g*v,St=d*w-_*v,Gt=m*C-g*M,he=m*w-_*M,ge=g*w-_*C,Ne=L*ge-D*he+N*Gt+et*St-rt*Mt+vt*at;if(Ne===0)return this.set(0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0);const xe=1/Ne;return t[0]=(a*ge-l*he+c*Gt)*xe,t[1]=(i*he-n*ge-s*Gt)*xe,t[2]=(M*vt-C*rt+w*et)*xe,t[3]=(g*rt-m*vt-_*et)*xe,t[4]=(l*St-r*ge-c*Mt)*xe,t[5]=(e*ge-i*St+s*Mt)*xe,t[6]=(C*N-v*vt-w*D)*xe,t[7]=(d*vt-g*N+_*D)*xe,t[8]=(r*he-a*St+c*at)*xe,t[9]=(n*St-e*he-s*at)*xe,t[10]=(v*rt-M*N+w*L)*xe,t[11]=(m*N-d*rt-_*L)*xe,t[12]=(a*Mt-r*Gt-l*at)*xe,t[13]=(e*Gt-n*Mt+i*at)*xe,t[14]=(M*D-v*et-C*L)*xe,t[15]=(d*et-m*D+g*L)*xe,this}scale(t){const e=this.elements,n=t.x,i=t.y,s=t.z;return e[0]*=n,e[4]*=i,e[8]*=s,e[1]*=n,e[5]*=i,e[9]*=s,e[2]*=n,e[6]*=i,e[10]*=s,e[3]*=n,e[7]*=i,e[11]*=s,this}getMaxScaleOnAxis(){const t=this.elements,e=t[0]*t[0]+t[1]*t[1]+t[2]*t[2],n=t[4]*t[4]+t[5]*t[5]+t[6]*t[6],i=t[8]*t[8]+t[9]*t[9]+t[10]*t[10];return Math.sqrt(Math.max(e,n,i))}makeTranslation(t,e,n){return t.isVector3?this.set(1,0,0,t.x,0,1,0,t.y,0,0,1,t.z,0,0,0,1):this.set(1,0,0,t,0,1,0,e,0,0,1,n,0,0,0,1),this}makeRotationX(t){const e=Math.cos(t),n=Math.sin(t);return this.set(1,0,0,0,0,e,-n,0,0,n,e,0,0,0,0,1),this}makeRotationY(t){const e=Math.cos(t),n=Math.sin(t);return this.set(e,0,n,0,0,1,0,0,-n,0,e,0,0,0,0,1),this}makeRotationZ(t){const e=Math.cos(t),n=Math.sin(t);return this.set(e,-n,0,0,n,e,0,0,0,0,1,0,0,0,0,1),this}makeRotationAxis(t,e){const n=Math.cos(e),i=Math.sin(e),s=1-n,r=t.x,a=t.y,l=t.z,c=s*r,d=s*a;return this.set(c*r+n,c*a-i*l,c*l+i*a,0,c*a+i*l,d*a+n,d*l-i*r,0,c*l-i*a,d*l+i*r,s*l*l+n,0,0,0,0,1),this}makeScale(t,e,n){return this.set(t,0,0,0,0,e,0,0,0,0,n,0,0,0,0,1),this}makeShear(t,e,n,i,s,r){return this.set(1,n,s,0,t,1,r,0,e,i,1,0,0,0,0,1),this}compose(t,e,n){const i=this.elements,s=e._x,r=e._y,a=e._z,l=e._w,c=s+s,d=r+r,m=a+a,g=s*c,_=s*d,v=s*m,M=r*d,C=r*m,w=a*m,L=l*c,D=l*d,N=l*m,et=n.x,rt=n.y,vt=n.z;return i[0]=(1-(M+w))*et,i[1]=(_+N)*et,i[2]=(v-D)*et,i[3]=0,i[4]=(_-N)*rt,i[5]=(1-(g+w))*rt,i[6]=(C+L)*rt,i[7]=0,i[8]=(v+D)*vt,i[9]=(C-L)*vt,i[10]=(1-(g+M))*vt,i[11]=0,i[12]=t.x,i[13]=t.y,i[14]=t.z,i[15]=1,this}decompose(t,e,n){const i=this.elements;t.x=i[12],t.y=i[13],t.z=i[14];const s=this.determinant();if(s===0)return n.set(1,1,1),e.identity(),this;let r=Fi.set(i[0],i[1],i[2]).length();const a=Fi.set(i[4],i[5],i[6]).length(),l=Fi.set(i[8],i[9],i[10]).length();s<0&&(r=-r),On.copy(this);const c=1/r,d=1/a,m=1/l;return On.elements[0]*=c,On.elements[1]*=c,On.elements[2]*=c,On.elements[4]*=d,On.elements[5]*=d,On.elements[6]*=d,On.elements[8]*=m,On.elements[9]*=m,On.elements[10]*=m,e.setFromRotationMatrix(On),n.x=r,n.y=a,n.z=l,this}makePerspective(t,e,n,i,s,r,a=Nn,l=!1){const c=this.elements,d=2*s/(e-t),m=2*s/(n-i),g=(e+t)/(e-t),_=(n+i)/(n-i);let v,M;if(l)v=s/(r-s),M=r*s/(r-s);else if(a===Nn)v=-(r+s)/(r-s),M=-2*r*s/(r-s);else if(a===Pi)v=-r/(r-s),M=-r*s/(r-s);else throw new Error("THREE.Matrix4.makePerspective(): Invalid coordinate system: "+a);return c[0]=d,c[4]=0,c[8]=g,c[12]=0,c[1]=0,c[5]=m,c[9]=_,c[13]=0,c[2]=0,c[6]=0,c[10]=v,c[14]=M,c[3]=0,c[7]=0,c[11]=-1,c[15]=0,this}makeOrthographic(t,e,n,i,s,r,a=Nn,l=!1){const c=this.elements,d=2/(e-t),m=2/(n-i),g=-(e+t)/(e-t),_=-(n+i)/(n-i);let v,M;if(l)v=1/(r-s),M=r/(r-s);else if(a===Nn)v=-2/(r-s),M=-(r+s)/(r-s);else if(a===Pi)v=-1/(r-s),M=-s/(r-s);else throw new Error("THREE.Matrix4.makeOrthographic(): Invalid coordinate system: "+a);return c[0]=d,c[4]=0,c[8]=0,c[12]=g,c[1]=0,c[5]=m,c[9]=0,c[13]=_,c[2]=0,c[6]=0,c[10]=v,c[14]=M,c[3]=0,c[7]=0,c[11]=0,c[15]=1,this}equals(t){const e=this.elements,n=t.elements;for(let i=0;i<16;i++)if(e[i]!==n[i])return!1;return!0}fromArray(t,e=0){for(let n=0;n<16;n++)this.elements[n]=t[n+e];return this}toArray(t=[],e=0){const n=this.elements;return t[e]=n[0],t[e+1]=n[1],t[e+2]=n[2],t[e+3]=n[3],t[e+4]=n[4],t[e+5]=n[5],t[e+6]=n[6],t[e+7]=n[7],t[e+8]=n[8],t[e+9]=n[9],t[e+10]=n[10],t[e+11]=n[11],t[e+12]=n[12],t[e+13]=n[13],t[e+14]=n[14],t[e+15]=n[15],t}}const Fi=new O,On=new Me,Cc=new O(0,0,0),Rc=new O(1,1,1),jn=new O,is=new O,wn=new O,Wa=new Me,Xa=new un;class Bn{constructor(t=0,e=0,n=0,i=Bn.DEFAULT_ORDER){this.isEuler=!0,this._x=t,this._y=e,this._z=n,this._order=i}get x(){return this._x}set x(t){this._x=t,this._onChangeCallback()}get y(){return this._y}set y(t){this._y=t,this._onChangeCallback()}get z(){return this._z}set z(t){this._z=t,this._onChangeCallback()}get order(){return this._order}set order(t){this._order=t,this._onChangeCallback()}set(t,e,n,i=this._order){return this._x=t,this._y=e,this._z=n,this._order=i,this._onChangeCallback(),this}clone(){return new this.constructor(this._x,this._y,this._z,this._order)}copy(t){return this._x=t._x,this._y=t._y,this._z=t._z,this._order=t._order,this._onChangeCallback(),this}setFromRotationMatrix(t,e=this._order,n=!0){const i=t.elements,s=i[0],r=i[4],a=i[8],l=i[1],c=i[5],d=i[9],m=i[2],g=i[6],_=i[10];switch(e){case"XYZ":this._y=Math.asin(pe(a,-1,1)),Math.abs(a)<.9999999?(this._x=Math.atan2(-d,_),this._z=Math.atan2(-r,s)):(this._x=Math.atan2(g,c),this._z=0);break;case"YXZ":this._x=Math.asin(-pe(d,-1,1)),Math.abs(d)<.9999999?(this._y=Math.atan2(a,_),this._z=Math.atan2(l,c)):(this._y=Math.atan2(-m,s),this._z=0);break;case"ZXY":this._x=Math.asin(pe(g,-1,1)),Math.abs(g)<.9999999?(this._y=Math.atan2(-m,_),this._z=Math.atan2(-r,c)):(this._y=0,this._z=Math.atan2(l,s));break;case"ZYX":this._y=Math.asin(-pe(m,-1,1)),Math.abs(m)<.9999999?(this._x=Math.atan2(g,_),this._z=Math.atan2(l,s)):(this._x=0,this._z=Math.atan2(-r,c));break;case"YZX":this._z=Math.asin(pe(l,-1,1)),Math.abs(l)<.9999999?(this._x=Math.atan2(-d,c),this._y=Math.atan2(-m,s)):(this._x=0,this._y=Math.atan2(a,_));break;case"XZY":this._z=Math.asin(-pe(r,-1,1)),Math.abs(r)<.9999999?(this._x=Math.atan2(g,c),this._y=Math.atan2(a,s)):(this._x=Math.atan2(-d,_),this._y=0);break;default:le("Euler: .setFromRotationMatrix() encountered an unknown order: "+e)}return this._order=e,n===!0&&this._onChangeCallback(),this}setFromQuaternion(t,e,n){return Wa.makeRotationFromQuaternion(t),this.setFromRotationMatrix(Wa,e,n)}setFromVector3(t,e=this._order){return this.set(t.x,t.y,t.z,e)}reorder(t){return Xa.setFromEuler(this),this.setFromQuaternion(Xa,t)}equals(t){return t._x===this._x&&t._y===this._y&&t._z===this._z&&t._order===this._order}fromArray(t){return this._x=t[0],this._y=t[1],this._z=t[2],t[3]!==void 0&&(this._order=t[3]),this._onChangeCallback(),this}toArray(t=[],e=0){return t[e]=this._x,t[e+1]=this._y,t[e+2]=this._z,t[e+3]=this._order,t}_onChange(t){return this._onChangeCallback=t,this}_onChangeCallback(){}*[Symbol.iterator](){yield this._x,yield this._y,yield this._z,yield this._order}}Bn.DEFAULT_ORDER="XYZ";class wr{constructor(){this.mask=1}set(t){this.mask=(1<<t|0)>>>0}enable(t){this.mask|=1<<t|0}enableAll(){this.mask=-1}toggle(t){this.mask^=1<<t|0}disable(t){this.mask&=~(1<<t|0)}disableAll(){this.mask=0}test(t){return(this.mask&t.mask)!==0}isEnabled(t){return(this.mask&(1<<t|0))!==0}}let Ic=0;const Cr=new O,Oi=new un,ti=new Me,Ns=new O,ss=new O,Pc=new O,qa=new un,Bi=new O(1,0,0),Fs=new O(0,1,0),Ya=new O(0,0,1),Za={type:"added"},Lc={type:"removed"},zi={type:"childadded",child:null},Rr={type:"childremoved",child:null};class De extends Kn{constructor(){super(),this.isObject3D=!0,Object.defineProperty(this,"id",{value:Ic++}),this.uuid=An(),this.name="",this.type="Object3D",this.parent=null,this.children=[],this.up=De.DEFAULT_UP.clone();const t=new O,e=new Bn,n=new un,i=new O(1,1,1);function s(){n.setFromEuler(e,!1)}function r(){e.setFromQuaternion(n,void 0,!1)}e._onChange(s),n._onChange(r),Object.defineProperties(this,{position:{configurable:!0,enumerable:!0,value:t},rotation:{configurable:!0,enumerable:!0,value:e},quaternion:{configurable:!0,enumerable:!0,value:n},scale:{configurable:!0,enumerable:!0,value:i},modelViewMatrix:{value:new Me},normalMatrix:{value:new Fn}}),this.matrix=new Me,this.matrixWorld=new Me,this.matrixAutoUpdate=De.DEFAULT_MATRIX_AUTO_UPDATE,this.matrixWorldAutoUpdate=De.DEFAULT_MATRIX_WORLD_AUTO_UPDATE,this.matrixWorldNeedsUpdate=!1,this.layers=new wr,this.visible=!0,this.castShadow=!1,this.receiveShadow=!1,this.frustumCulled=!0,this.renderOrder=0,this.animations=[],this.customDepthMaterial=void 0,this.customDistanceMaterial=void 0,this.static=!1,this.userData={},this.pivot=null}onBeforeShadow(){}onAfterShadow(){}onBeforeRender(){}onAfterRender(){}applyMatrix4(t){this.matrixAutoUpdate&&this.updateMatrix(),this.matrix.premultiply(t),this.matrix.decompose(this.position,this.quaternion,this.scale)}applyQuaternion(t){return this.quaternion.premultiply(t),this}setRotationFromAxisAngle(t,e){this.quaternion.setFromAxisAngle(t,e)}setRotationFromEuler(t){this.quaternion.setFromEuler(t,!0)}setRotationFromMatrix(t){this.quaternion.setFromRotationMatrix(t)}setRotationFromQuaternion(t){this.quaternion.copy(t)}rotateOnAxis(t,e){return Oi.setFromAxisAngle(t,e),this.quaternion.multiply(Oi),this}rotateOnWorldAxis(t,e){return Oi.setFromAxisAngle(t,e),this.quaternion.premultiply(Oi),this}rotateX(t){return this.rotateOnAxis(Bi,t)}rotateY(t){return this.rotateOnAxis(Fs,t)}rotateZ(t){return this.rotateOnAxis(Ya,t)}translateOnAxis(t,e){return Cr.copy(t).applyQuaternion(this.quaternion),this.position.add(Cr.multiplyScalar(e)),this}translateX(t){return this.translateOnAxis(Bi,t)}translateY(t){return this.translateOnAxis(Fs,t)}translateZ(t){return this.translateOnAxis(Ya,t)}localToWorld(t){return this.updateWorldMatrix(!0,!1),t.applyMatrix4(this.matrixWorld)}worldToLocal(t){return this.updateWorldMatrix(!0,!1),t.applyMatrix4(ti.copy(this.matrixWorld).invert())}lookAt(t,e,n){t.isVector3?Ns.copy(t):Ns.set(t,e,n);const i=this.parent;this.updateWorldMatrix(!0,!1),ss.setFromMatrixPosition(this.matrixWorld),this.isCamera||this.isLight?ti.lookAt(ss,Ns,this.up):ti.lookAt(Ns,ss,this.up),this.quaternion.setFromRotationMatrix(ti),i&&(ti.extractRotation(i.matrixWorld),Oi.setFromRotationMatrix(ti),this.quaternion.premultiply(Oi.invert()))}add(t){if(arguments.length>1){for(let e=0;e<arguments.length;e++)this.add(arguments[e]);return this}return t===this?(Re("Object3D.add: object can't be added as a child of itself.",t),this):(t&&t.isObject3D?(t.removeFromParent(),t.parent=this,this.children.push(t),t.dispatchEvent(Za),zi.child=t,this.dispatchEvent(zi),zi.child=null):Re("Object3D.add: object not an instance of THREE.Object3D.",t),this)}remove(t){if(arguments.length>1){for(let n=0;n<arguments.length;n++)this.remove(arguments[n]);return this}const e=this.children.indexOf(t);return e!==-1&&(t.parent=null,this.children.splice(e,1),t.dispatchEvent(Lc),Rr.child=t,this.dispatchEvent(Rr),Rr.child=null),this}removeFromParent(){const t=this.parent;return t!==null&&t.remove(this),this}clear(){return this.remove(...this.children)}attach(t){return this.updateWorldMatrix(!0,!1),ti.copy(this.matrixWorld).invert(),t.parent!==null&&(t.parent.updateWorldMatrix(!0,!1),ti.multiply(t.parent.matrixWorld)),t.applyMatrix4(ti),t.removeFromParent(),t.parent=this,this.children.push(t),t.updateWorldMatrix(!1,!0),t.dispatchEvent(Za),zi.child=t,this.dispatchEvent(zi),zi.child=null,this}getObjectById(t){return this.getObjectByProperty("id",t)}getObjectByName(t){return this.getObjectByProperty("name",t)}getObjectByProperty(t,e){if(this[t]===e)return this;for(let n=0,i=this.children.length;n<i;n++){const r=this.children[n].getObjectByProperty(t,e);if(r!==void 0)return r}}getObjectsByProperty(t,e,n=[]){this[t]===e&&n.push(this);const i=this.children;for(let s=0,r=i.length;s<r;s++)i[s].getObjectsByProperty(t,e,n);return n}getWorldPosition(t){return this.updateWorldMatrix(!0,!1),t.setFromMatrixPosition(this.matrixWorld)}getWorldQuaternion(t){return this.updateWorldMatrix(!0,!1),this.matrixWorld.decompose(ss,t,Pc),t}getWorldScale(t){return this.updateWorldMatrix(!0,!1),this.matrixWorld.decompose(ss,qa,t),t}getWorldDirection(t){this.updateWorldMatrix(!0,!1);const e=this.matrixWorld.elements;return t.set(e[8],e[9],e[10]).normalize()}raycast(){}traverse(t){t(this);const e=this.children;for(let n=0,i=e.length;n<i;n++)e[n].traverse(t)}traverseVisible(t){if(this.visible===!1)return;t(this);const e=this.children;for(let n=0,i=e.length;n<i;n++)e[n].traverseVisible(t)}traverseAncestors(t){const e=this.parent;e!==null&&(t(e),e.traverseAncestors(t))}updateMatrix(){this.matrix.compose(this.position,this.quaternion,this.scale);const t=this.pivot;if(t!==null){const e=t.x,n=t.y,i=t.z,s=this.matrix.elements;s[12]+=e-s[0]*e-s[4]*n-s[8]*i,s[13]+=n-s[1]*e-s[5]*n-s[9]*i,s[14]+=i-s[2]*e-s[6]*n-s[10]*i}this.matrixWorldNeedsUpdate=!0}updateMatrixWorld(t){this.matrixAutoUpdate&&this.updateMatrix(),(this.matrixWorldNeedsUpdate||t)&&(this.matrixWorldAutoUpdate===!0&&(this.parent===null?this.matrixWorld.copy(this.matrix):this.matrixWorld.multiplyMatrices(this.parent.matrixWorld,this.matrix)),this.matrixWorldNeedsUpdate=!1,t=!0);const e=this.children;for(let n=0,i=e.length;n<i;n++)e[n].updateMatrixWorld(t)}updateWorldMatrix(t,e){const n=this.parent;if(t===!0&&n!==null&&n.updateWorldMatrix(!0,!1),this.matrixAutoUpdate&&this.updateMatrix(),this.matrixWorldAutoUpdate===!0&&(this.parent===null?this.matrixWorld.copy(this.matrix):this.matrixWorld.multiplyMatrices(this.parent.matrixWorld,this.matrix)),e===!0){const i=this.children;for(let s=0,r=i.length;s<r;s++)i[s].updateWorldMatrix(!1,!0)}}toJSON(t){const e=t===void 0||typeof t=="string",n={};e&&(t={geometries:{},materials:{},textures:{},images:{},shapes:{},skeletons:{},animations:{},nodes:{}},n.metadata={version:4.7,type:"Object",generator:"Object3D.toJSON"});const i={};i.uuid=this.uuid,i.type=this.type,this.name!==""&&(i.name=this.name),this.castShadow===!0&&(i.castShadow=!0),this.receiveShadow===!0&&(i.receiveShadow=!0),this.visible===!1&&(i.visible=!1),this.frustumCulled===!1&&(i.frustumCulled=!1),this.renderOrder!==0&&(i.renderOrder=this.renderOrder),this.static!==!1&&(i.static=this.static),Object.keys(this.userData).length>0&&(i.userData=this.userData),i.layers=this.layers.mask,i.matrix=this.matrix.toArray(),i.up=this.up.toArray(),this.pivot!==null&&(i.pivot=this.pivot.toArray()),this.matrixAutoUpdate===!1&&(i.matrixAutoUpdate=!1),this.morphTargetDictionary!==void 0&&(i.morphTargetDictionary=Object.assign({},this.morphTargetDictionary)),this.morphTargetInfluences!==void 0&&(i.morphTargetInfluences=this.morphTargetInfluences.slice()),this.isInstancedMesh&&(i.type="InstancedMesh",i.count=this.count,i.instanceMatrix=this.instanceMatrix.toJSON(),this.instanceColor!==null&&(i.instanceColor=this.instanceColor.toJSON())),this.isBatchedMesh&&(i.type="BatchedMesh",i.perObjectFrustumCulled=this.perObjectFrustumCulled,i.sortObjects=this.sortObjects,i.drawRanges=this._drawRanges,i.reservedRanges=this._reservedRanges,i.geometryInfo=this._geometryInfo.map(a=>({...a,boundingBox:a.boundingBox?a.boundingBox.toJSON():void 0,boundingSphere:a.boundingSphere?a.boundingSphere.toJSON():void 0})),i.instanceInfo=this._instanceInfo.map(a=>({...a})),i.availableInstanceIds=this._availableInstanceIds.slice(),i.availableGeometryIds=this._availableGeometryIds.slice(),i.nextIndexStart=this._nextIndexStart,i.nextVertexStart=this._nextVertexStart,i.geometryCount=this._geometryCount,i.maxInstanceCount=this._maxInstanceCount,i.maxVertexCount=this._maxVertexCount,i.maxIndexCount=this._maxIndexCount,i.geometryInitialized=this._geometryInitialized,i.matricesTexture=this._matricesTexture.toJSON(t),i.indirectTexture=this._indirectTexture.toJSON(t),this._colorsTexture!==null&&(i.colorsTexture=this._colorsTexture.toJSON(t)),this.boundingSphere!==null&&(i.boundingSphere=this.boundingSphere.toJSON()),this.boundingBox!==null&&(i.boundingBox=this.boundingBox.toJSON()));function s(a,l){return a[l.uuid]===void 0&&(a[l.uuid]=l.toJSON(t)),l.uuid}if(this.isScene)this.background&&(this.background.isColor?i.background=this.background.toJSON():this.background.isTexture&&(i.background=this.background.toJSON(t).uuid)),this.environment&&this.environment.isTexture&&this.environment.isRenderTargetTexture!==!0&&(i.environment=this.environment.toJSON(t).uuid);else if(this.isMesh||this.isLine||this.isPoints){i.geometry=s(t.geometries,this.geometry);const a=this.geometry.parameters;if(a!==void 0&&a.shapes!==void 0){const l=a.shapes;if(Array.isArray(l))for(let c=0,d=l.length;c<d;c++){const m=l[c];s(t.shapes,m)}else s(t.shapes,l)}}if(this.isSkinnedMesh&&(i.bindMode=this.bindMode,i.bindMatrix=this.bindMatrix.toArray(),this.skeleton!==void 0&&(s(t.skeletons,this.skeleton),i.skeleton=this.skeleton.uuid)),this.material!==void 0)if(Array.isArray(this.material)){const a=[];for(let l=0,c=this.material.length;l<c;l++)a.push(s(t.materials,this.material[l]));i.material=a}else i.material=s(t.materials,this.material);if(this.children.length>0){i.children=[];for(let a=0;a<this.children.length;a++)i.children.push(this.children[a].toJSON(t).object)}if(this.animations.length>0){i.animations=[];for(let a=0;a<this.animations.length;a++){const l=this.animations[a];i.animations.push(s(t.animations,l))}}if(e){const a=r(t.geometries),l=r(t.materials),c=r(t.textures),d=r(t.images),m=r(t.shapes),g=r(t.skeletons),_=r(t.animations),v=r(t.nodes);a.length>0&&(n.geometries=a),l.length>0&&(n.materials=l),c.length>0&&(n.textures=c),d.length>0&&(n.images=d),m.length>0&&(n.shapes=m),g.length>0&&(n.skeletons=g),_.length>0&&(n.animations=_),v.length>0&&(n.nodes=v)}return n.object=i,n;function r(a){const l=[];for(const c in a){const d=a[c];delete d.metadata,l.push(d)}return l}}clone(t){return new this.constructor().copy(this,t)}copy(t,e=!0){if(this.name=t.name,this.up.copy(t.up),this.position.copy(t.position),this.rotation.order=t.rotation.order,this.quaternion.copy(t.quaternion),this.scale.copy(t.scale),t.pivot!==null&&(this.pivot=t.pivot.clone()),this.matrix.copy(t.matrix),this.matrixWorld.copy(t.matrixWorld),this.matrixAutoUpdate=t.matrixAutoUpdate,this.matrixWorldAutoUpdate=t.matrixWorldAutoUpdate,this.matrixWorldNeedsUpdate=t.matrixWorldNeedsUpdate,this.layers.mask=t.layers.mask,this.visible=t.visible,this.castShadow=t.castShadow,this.receiveShadow=t.receiveShadow,this.frustumCulled=t.frustumCulled,this.renderOrder=t.renderOrder,this.static=t.static,this.animations=t.animations.slice(),this.userData=JSON.parse(JSON.stringify(t.userData)),e===!0)for(let n=0;n<t.children.length;n++){const i=t.children[n];this.add(i.clone())}return this}}De.DEFAULT_UP=new O(0,1,0),De.DEFAULT_MATRIX_AUTO_UPDATE=!0,De.DEFAULT_MATRIX_WORLD_AUTO_UPDATE=!0;class rs extends De{constructor(){super(),this.isGroup=!0,this.type="Group"}}const Dc={type:"move"};class vi{constructor(){this._targetRay=null,this._grip=null,this._hand=null}getHandSpace(){return this._hand===null&&(this._hand=new rs,this._hand.matrixAutoUpdate=!1,this._hand.visible=!1,this._hand.joints={},this._hand.inputState={pinching:!1}),this._hand}getTargetRaySpace(){return this._targetRay===null&&(this._targetRay=new rs,this._targetRay.matrixAutoUpdate=!1,this._targetRay.visible=!1,this._targetRay.hasLinearVelocity=!1,this._targetRay.linearVelocity=new O,this._targetRay.hasAngularVelocity=!1,this._targetRay.angularVelocity=new O),this._targetRay}getGripSpace(){return this._grip===null&&(this._grip=new rs,this._grip.matrixAutoUpdate=!1,this._grip.visible=!1,this._grip.hasLinearVelocity=!1,this._grip.linearVelocity=new O,this._grip.hasAngularVelocity=!1,this._grip.angularVelocity=new O),this._grip}dispatchEvent(t){return this._targetRay!==null&&this._targetRay.dispatchEvent(t),this._grip!==null&&this._grip.dispatchEvent(t),this._hand!==null&&this._hand.dispatchEvent(t),this}connect(t){if(t&&t.hand){const e=this._hand;if(e)for(const n of t.hand.values())this._getHandJoint(e,n)}return this.dispatchEvent({type:"connected",data:t}),this}disconnect(t){return this.dispatchEvent({type:"disconnected",data:t}),this._targetRay!==null&&(this._targetRay.visible=!1),this._grip!==null&&(this._grip.visible=!1),this._hand!==null&&(this._hand.visible=!1),this}update(t,e,n){let i=null,s=null,r=null;const a=this._targetRay,l=this._grip,c=this._hand;if(t&&e.session.visibilityState!=="visible-blurred"){if(c&&t.hand){r=!0;for(const M of t.hand.values()){const C=e.getJointPose(M,n),w=this._getHandJoint(c,M);C!==null&&(w.matrix.fromArray(C.transform.matrix),w.matrix.decompose(w.position,w.rotation,w.scale),w.matrixWorldNeedsUpdate=!0,w.jointRadius=C.radius),w.visible=C!==null}const d=c.joints["index-finger-tip"],m=c.joints["thumb-tip"],g=d.position.distanceTo(m.position),_=.02,v=.005;c.inputState.pinching&&g>_+v?(c.inputState.pinching=!1,this.dispatchEvent({type:"pinchend",handedness:t.handedness,target:this})):!c.inputState.pinching&&g<=_-v&&(c.inputState.pinching=!0,this.dispatchEvent({type:"pinchstart",handedness:t.handedness,target:this}))}else l!==null&&t.gripSpace&&(s=e.getPose(t.gripSpace,n),s!==null&&(l.matrix.fromArray(s.transform.matrix),l.matrix.decompose(l.position,l.rotation,l.scale),l.matrixWorldNeedsUpdate=!0,s.linearVelocity?(l.hasLinearVelocity=!0,l.linearVelocity.copy(s.linearVelocity)):l.hasLinearVelocity=!1,s.angularVelocity?(l.hasAngularVelocity=!0,l.angularVelocity.copy(s.angularVelocity)):l.hasAngularVelocity=!1));a!==null&&(i=e.getPose(t.targetRaySpace,n),i===null&&s!==null&&(i=s),i!==null&&(a.matrix.fromArray(i.transform.matrix),a.matrix.decompose(a.position,a.rotation,a.scale),a.matrixWorldNeedsUpdate=!0,i.linearVelocity?(a.hasLinearVelocity=!0,a.linearVelocity.copy(i.linearVelocity)):a.hasLinearVelocity=!1,i.angularVelocity?(a.hasAngularVelocity=!0,a.angularVelocity.copy(i.angularVelocity)):a.hasAngularVelocity=!1,this.dispatchEvent(Dc)))}return a!==null&&(a.visible=i!==null),l!==null&&(l.visible=s!==null),c!==null&&(c.visible=r!==null),this}_getHandJoint(t,e){if(t.joints[e.jointName]===void 0){const n=new rs;n.matrixAutoUpdate=!1,n.visible=!1,t.joints[e.jointName]=n,t.add(n)}return t.joints[e.jointName]}}const $a={aliceblue:15792383,antiquewhite:16444375,aqua:65535,aquamarine:8388564,azure:15794175,beige:16119260,bisque:16770244,black:0,blanchedalmond:16772045,blue:255,blueviolet:9055202,brown:10824234,burlywood:14596231,cadetblue:6266528,chartreuse:8388352,chocolate:13789470,coral:16744272,cornflowerblue:6591981,cornsilk:16775388,crimson:14423100,cyan:65535,darkblue:139,darkcyan:35723,darkgoldenrod:12092939,darkgray:11119017,darkgreen:25600,darkgrey:11119017,darkkhaki:12433259,darkmagenta:9109643,darkolivegreen:5597999,darkorange:16747520,darkorchid:10040012,darkred:9109504,darksalmon:15308410,darkseagreen:9419919,darkslateblue:4734347,darkslategray:3100495,darkslategrey:3100495,darkturquoise:52945,darkviolet:9699539,deeppink:16716947,deepskyblue:49151,dimgray:6908265,dimgrey:6908265,dodgerblue:2003199,firebrick:11674146,floralwhite:16775920,forestgreen:2263842,fuchsia:16711935,gainsboro:14474460,ghostwhite:16316671,gold:16766720,goldenrod:14329120,gray:8421504,green:32768,greenyellow:11403055,grey:8421504,honeydew:15794160,hotpink:16738740,indianred:13458524,indigo:4915330,ivory:16777200,khaki:15787660,lavender:15132410,lavenderblush:16773365,lawngreen:8190976,lemonchiffon:16775885,lightblue:11393254,lightcoral:15761536,lightcyan:14745599,lightgoldenrodyellow:16448210,lightgray:13882323,lightgreen:9498256,lightgrey:13882323,lightpink:16758465,lightsalmon:16752762,lightseagreen:2142890,lightskyblue:8900346,lightslategray:7833753,lightslategrey:7833753,lightsteelblue:11584734,lightyellow:16777184,lime:65280,limegreen:3329330,linen:16445670,magenta:16711935,maroon:8388608,mediumaquamarine:6737322,mediumblue:205,mediumorchid:12211667,mediumpurple:9662683,mediumseagreen:3978097,mediumslateblue:8087790,mediumspringgreen:64154,mediumturquoise:4772300,mediumvioletred:13047173,midnightblue:1644912,mintcream:16121850,mistyrose:16770273,moccasin:16770229,navajowhite:16768685,navy:128,oldlace:16643558,olive:8421376,olivedrab:7048739,orange:16753920,orangered:16729344,orchid:14315734,palegoldenrod:15657130,palegreen:10025880,paleturquoise:11529966,palevioletred:14381203,papayawhip:16773077,peachpuff:16767673,peru:13468991,pink:16761035,plum:14524637,powderblue:11591910,purple:8388736,rebeccapurple:6697881,red:16711680,rosybrown:12357519,royalblue:4286945,saddlebrown:9127187,salmon:16416882,sandybrown:16032864,seagreen:3050327,seashell:16774638,sienna:10506797,silver:12632256,skyblue:8900331,slateblue:6970061,slategray:7372944,slategrey:7372944,snow:16775930,springgreen:65407,steelblue:4620980,tan:13808780,teal:32896,thistle:14204888,tomato:16737095,turquoise:4251856,violet:15631086,wheat:16113331,white:16777215,whitesmoke:16119285,yellow:16776960,yellowgreen:10145074},li={h:0,s:0,l:0},Os={h:0,s:0,l:0};function Ir(u,t,e){return e<0&&(e+=1),e>1&&(e-=1),e<1/6?u+(t-u)*6*e:e<1/2?t:e<2/3?u+(t-u)*6*(2/3-e):u}class re{constructor(t,e,n){return this.isColor=!0,this.r=1,this.g=1,this.b=1,this.set(t,e,n)}set(t,e,n){if(e===void 0&&n===void 0){const i=t;i&&i.isColor?this.copy(i):typeof i=="number"?this.setHex(i):typeof i=="string"&&this.setStyle(i)}else this.setRGB(t,e,n);return this}setScalar(t){return this.r=t,this.g=t,this.b=t,this}setHex(t,e=yn){return t=Math.floor(t),this.r=(t>>16&255)/255,this.g=(t>>8&255)/255,this.b=(t&255)/255,_n.colorSpaceToWorking(this,e),this}setRGB(t,e,n,i=_n.workingColorSpace){return this.r=t,this.g=e,this.b=n,_n.colorSpaceToWorking(this,i),this}setHSL(t,e,n,i=_n.workingColorSpace){if(t=xr(t,1),e=pe(e,0,1),n=pe(n,0,1),e===0)this.r=this.g=this.b=n;else{const s=n<=.5?n*(1+e):n+e-n*e,r=2*n-s;this.r=Ir(r,s,t+1/3),this.g=Ir(r,s,t),this.b=Ir(r,s,t-1/3)}return _n.colorSpaceToWorking(this,i),this}setStyle(t,e=yn){function n(s){s!==void 0&&parseFloat(s)<1&&le("Color: Alpha component of "+t+" will be ignored.")}let i;if(i=/^(\w+)\(([^\)]*)\)/.exec(t)){let s;const r=i[1],a=i[2];switch(r){case"rgb":case"rgba":if(s=/^\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*(?:,\s*(\d*\.?\d+)\s*)?$/.exec(a))return n(s[4]),this.setRGB(Math.min(255,parseInt(s[1],10))/255,Math.min(255,parseInt(s[2],10))/255,Math.min(255,parseInt(s[3],10))/255,e);if(s=/^\s*(\d+)\%\s*,\s*(\d+)\%\s*,\s*(\d+)\%\s*(?:,\s*(\d*\.?\d+)\s*)?$/.exec(a))return n(s[4]),this.setRGB(Math.min(100,parseInt(s[1],10))/100,Math.min(100,parseInt(s[2],10))/100,Math.min(100,parseInt(s[3],10))/100,e);break;case"hsl":case"hsla":if(s=/^\s*(\d*\.?\d+)\s*,\s*(\d*\.?\d+)\%\s*,\s*(\d*\.?\d+)\%\s*(?:,\s*(\d*\.?\d+)\s*)?$/.exec(a))return n(s[4]),this.setHSL(parseFloat(s[1])/360,parseFloat(s[2])/100,parseFloat(s[3])/100,e);break;default:le("Color: Unknown color model "+t)}}else if(i=/^\#([A-Fa-f\d]+)$/.exec(t)){const s=i[1],r=s.length;if(r===3)return this.setRGB(parseInt(s.charAt(0),16)/15,parseInt(s.charAt(1),16)/15,parseInt(s.charAt(2),16)/15,e);if(r===6)return this.setHex(parseInt(s,16),e);le("Color: Invalid hex color "+t)}else if(t&&t.length>0)return this.setColorName(t,e);return this}setColorName(t,e=yn){const n=$a[t.toLowerCase()];return n!==void 0?this.setHex(n,e):le("Color: Unknown color "+t),this}clone(){return new this.constructor(this.r,this.g,this.b)}copy(t){return this.r=t.r,this.g=t.g,this.b=t.b,this}copySRGBToLinear(t){return this.r=Qn(t.r),this.g=Qn(t.g),this.b=Qn(t.b),this}copyLinearToSRGB(t){return this.r=Ui(t.r),this.g=Ui(t.g),this.b=Ui(t.b),this}convertSRGBToLinear(){return this.copySRGBToLinear(this),this}convertLinearToSRGB(){return this.copyLinearToSRGB(this),this}getHex(t=yn){return _n.workingToColorSpace(xn.copy(this),t),Math.round(pe(xn.r*255,0,255))*65536+Math.round(pe(xn.g*255,0,255))*256+Math.round(pe(xn.b*255,0,255))}getHexString(t=yn){return("000000"+this.getHex(t).toString(16)).slice(-6)}getHSL(t,e=_n.workingColorSpace){_n.workingToColorSpace(xn.copy(this),e);const n=xn.r,i=xn.g,s=xn.b,r=Math.max(n,i,s),a=Math.min(n,i,s);let l,c;const d=(a+r)/2;if(a===r)l=0,c=0;else{const m=r-a;switch(c=d<=.5?m/(r+a):m/(2-r-a),r){case n:l=(i-s)/m+(i<s?6:0);break;case i:l=(s-n)/m+2;break;case s:l=(n-i)/m+4;break}l/=6}return t.h=l,t.s=c,t.l=d,t}getRGB(t,e=_n.workingColorSpace){return _n.workingToColorSpace(xn.copy(this),e),t.r=xn.r,t.g=xn.g,t.b=xn.b,t}getStyle(t=yn){_n.workingToColorSpace(xn.copy(this),t);const e=xn.r,n=xn.g,i=xn.b;return t!==yn?`color(${t} ${e.toFixed(3)} ${n.toFixed(3)} ${i.toFixed(3)})`:`rgb(${Math.round(e*255)},${Math.round(n*255)},${Math.round(i*255)})`}offsetHSL(t,e,n){return this.getHSL(li),this.setHSL(li.h+t,li.s+e,li.l+n)}add(t){return this.r+=t.r,this.g+=t.g,this.b+=t.b,this}addColors(t,e){return this.r=t.r+e.r,this.g=t.g+e.g,this.b=t.b+e.b,this}addScalar(t){return this.r+=t,this.g+=t,this.b+=t,this}sub(t){return this.r=Math.max(0,this.r-t.r),this.g=Math.max(0,this.g-t.g),this.b=Math.max(0,this.b-t.b),this}multiply(t){return this.r*=t.r,this.g*=t.g,this.b*=t.b,this}multiplyScalar(t){return this.r*=t,this.g*=t,this.b*=t,this}lerp(t,e){return this.r+=(t.r-this.r)*e,this.g+=(t.g-this.g)*e,this.b+=(t.b-this.b)*e,this}lerpColors(t,e,n){return this.r=t.r+(e.r-t.r)*n,this.g=t.g+(e.g-t.g)*n,this.b=t.b+(e.b-t.b)*n,this}lerpHSL(t,e){this.getHSL(li),t.getHSL(Os);const n=ns(li.h,Os.h,e),i=ns(li.s,Os.s,e),s=ns(li.l,Os.l,e);return this.setHSL(n,i,s),this}setFromVector3(t){return this.r=t.x,this.g=t.y,this.b=t.z,this}applyMatrix3(t){const e=this.r,n=this.g,i=this.b,s=t.elements;return this.r=s[0]*e+s[3]*n+s[6]*i,this.g=s[1]*e+s[4]*n+s[7]*i,this.b=s[2]*e+s[5]*n+s[8]*i,this}equals(t){return t.r===this.r&&t.g===this.g&&t.b===this.b}fromArray(t,e=0){return this.r=t[e],this.g=t[e+1],this.b=t[e+2],this}toArray(t=[],e=0){return t[e]=this.r,t[e+1]=this.g,t[e+2]=this.b,t}fromBufferAttribute(t,e){return this.r=t.getX(e),this.g=t.getY(e),this.b=t.getZ(e),this}toJSON(){return this.getHex()}*[Symbol.iterator](){yield this.r,yield this.g,yield this.b}}const xn=new re;re.NAMES=$a;class Pr{constructor(t,e=25e-5){this.isFogExp2=!0,this.name="",this.color=new re(t),this.density=e}clone(){return new Pr(this.color,this.density)}toJSON(){return{type:"FogExp2",name:this.name,color:this.color.getHex(),density:this.density}}}class o{constructor(t,e=1,n=1e3){this.isFog=!0,this.name="",this.color=new re(t),this.near=e,this.far=n}clone(){return new o(this.color,this.near,this.far)}toJSON(){return{type:"Fog",name:this.name,color:this.color.getHex(),near:this.near,far:this.far}}}class p extends De{constructor(){super(),this.isScene=!0,this.type="Scene",this.background=null,this.environment=null,this.fog=null,this.backgroundBlurriness=0,this.backgroundIntensity=1,this.backgroundRotation=new Bn,this.environmentIntensity=1,this.environmentRotation=new Bn,this.overrideMaterial=null,typeof __THREE_DEVTOOLS__<"u"&&__THREE_DEVTOOLS__.dispatchEvent(new CustomEvent("observe",{detail:this}))}copy(t,e){return super.copy(t,e),t.background!==null&&(this.background=t.background.clone()),t.environment!==null&&(this.environment=t.environment.clone()),t.fog!==null&&(this.fog=t.fog.clone()),this.backgroundBlurriness=t.backgroundBlurriness,this.backgroundIntensity=t.backgroundIntensity,this.backgroundRotation.copy(t.backgroundRotation),this.environmentIntensity=t.environmentIntensity,this.environmentRotation.copy(t.environmentRotation),t.overrideMaterial!==null&&(this.overrideMaterial=t.overrideMaterial.clone()),this.matrixAutoUpdate=t.matrixAutoUpdate,this}toJSON(t){const e=super.toJSON(t);return this.fog!==null&&(e.object.fog=this.fog.toJSON()),this.backgroundBlurriness>0&&(e.object.backgroundBlurriness=this.backgroundBlurriness),this.backgroundIntensity!==1&&(e.object.backgroundIntensity=this.backgroundIntensity),e.object.backgroundRotation=this.backgroundRotation.toArray(),this.environmentIntensity!==1&&(e.object.environmentIntensity=this.environmentIntensity),e.object.environmentRotation=this.environmentRotation.toArray(),e}}const h=new O,x=new O,E=new O,S=new O,I=new O,V=new O,j=new O,$=new O,ht=new O,K=new O,F=new En,G=new En,X=new En;class nt{constructor(t=new O,e=new O,n=new O){this.a=t,this.b=e,this.c=n}static getNormal(t,e,n,i){i.subVectors(n,e),h.subVectors(t,e),i.cross(h);const s=i.lengthSq();return s>0?i.multiplyScalar(1/Math.sqrt(s)):i.set(0,0,0)}static getBarycoord(t,e,n,i,s){h.subVectors(i,e),x.subVectors(n,e),E.subVectors(t,e);const r=h.dot(h),a=h.dot(x),l=h.dot(E),c=x.dot(x),d=x.dot(E),m=r*c-a*a;if(m===0)return s.set(0,0,0),null;const g=1/m,_=(c*l-a*d)*g,v=(r*d-a*l)*g;return s.set(1-_-v,v,_)}static containsPoint(t,e,n,i){return this.getBarycoord(t,e,n,i,S)===null?!1:S.x>=0&&S.y>=0&&S.x+S.y<=1}static getInterpolation(t,e,n,i,s,r,a,l){return this.getBarycoord(t,e,n,i,S)===null?(l.x=0,l.y=0,"z"in l&&(l.z=0),"w"in l&&(l.w=0),null):(l.setScalar(0),l.addScaledVector(s,S.x),l.addScaledVector(r,S.y),l.addScaledVector(a,S.z),l)}static getInterpolatedAttribute(t,e,n,i,s,r){return F.setScalar(0),G.setScalar(0),X.setScalar(0),F.fromBufferAttribute(t,e),G.fromBufferAttribute(t,n),X.fromBufferAttribute(t,i),r.setScalar(0),r.addScaledVector(F,s.x),r.addScaledVector(G,s.y),r.addScaledVector(X,s.z),r}static isFrontFacing(t,e,n,i){return h.subVectors(n,e),x.subVectors(t,e),h.cross(x).dot(i)<0}set(t,e,n){return this.a.copy(t),this.b.copy(e),this.c.copy(n),this}setFromPointsAndIndices(t,e,n,i){return this.a.copy(t[e]),this.b.copy(t[n]),this.c.copy(t[i]),this}setFromAttributeAndIndices(t,e,n,i){return this.a.fromBufferAttribute(t,e),this.b.fromBufferAttribute(t,n),this.c.fromBufferAttribute(t,i),this}clone(){return new this.constructor().copy(this)}copy(t){return this.a.copy(t.a),this.b.copy(t.b),this.c.copy(t.c),this}getArea(){return h.subVectors(this.c,this.b),x.subVectors(this.a,this.b),h.cross(x).length()*.5}getMidpoint(t){return t.addVectors(this.a,this.b).add(this.c).multiplyScalar(1/3)}getNormal(t){return nt.getNormal(this.a,this.b,this.c,t)}getPlane(t){return t.setFromCoplanarPoints(this.a,this.b,this.c)}getBarycoord(t,e){return nt.getBarycoord(t,this.a,this.b,this.c,e)}getInterpolation(t,e,n,i,s){return nt.getInterpolation(t,this.a,this.b,this.c,e,n,i,s)}containsPoint(t){return nt.containsPoint(t,this.a,this.b,this.c)}isFrontFacing(t){return nt.isFrontFacing(this.a,this.b,this.c,t)}intersectsBox(t){return t.intersectsTriangle(this)}closestPointToPoint(t,e){const n=this.a,i=this.b,s=this.c;let r,a;I.subVectors(i,n),V.subVectors(s,n),$.subVectors(t,n);const l=I.dot($),c=V.dot($);if(l<=0&&c<=0)return e.copy(n);ht.subVectors(t,i);const d=I.dot(ht),m=V.dot(ht);if(d>=0&&m<=d)return e.copy(i);const g=l*m-d*c;if(g<=0&&l>=0&&d<=0)return r=l/(l-d),e.copy(n).addScaledVector(I,r);K.subVectors(t,s);const _=I.dot(K),v=V.dot(K);if(v>=0&&_<=v)return e.copy(s);const M=_*c-l*v;if(M<=0&&c>=0&&v<=0)return a=c/(c-v),e.copy(n).addScaledVector(V,a);const C=d*v-_*m;if(C<=0&&m-d>=0&&_-v>=0)return j.subVectors(s,i),a=(m-d)/(m-d+(_-v)),e.copy(i).addScaledVector(j,a);const w=1/(C+M+g);return r=M*w,a=g*w,e.copy(n).addScaledVector(I,r).addScaledVector(V,a)}equals(t){return t.a.equals(this.a)&&t.b.equals(this.b)&&t.c.equals(this.c)}}class A{constructor(t=new O(1/0,1/0,1/0),e=new O(-1/0,-1/0,-1/0)){this.isBox3=!0,this.min=t,this.max=e}set(t,e){return this.min.copy(t),this.max.copy(e),this}setFromArray(t){this.makeEmpty();for(let e=0,n=t.length;e<n;e+=3)this.expandByPoint(z.fromArray(t,e));return this}setFromBufferAttribute(t){this.makeEmpty();for(let e=0,n=t.count;e<n;e++)this.expandByPoint(z.fromBufferAttribute(t,e));return this}setFromPoints(t){this.makeEmpty();for(let e=0,n=t.length;e<n;e++)this.expandByPoint(t[e]);return this}setFromCenterAndSize(t,e){const n=z.copy(e).multiplyScalar(.5);return this.min.copy(t).sub(n),this.max.copy(t).add(n),this}setFromObject(t,e=!1){return this.makeEmpty(),this.expandByObject(t,e)}clone(){return new this.constructor().copy(this)}copy(t){return this.min.copy(t.min),this.max.copy(t.max),this}makeEmpty(){return this.min.x=this.min.y=this.min.z=1/0,this.max.x=this.max.y=this.max.z=-1/0,this}isEmpty(){return this.max.x<this.min.x||this.max.y<this.min.y||this.max.z<this.min.z}getCenter(t){return this.isEmpty()?t.set(0,0,0):t.addVectors(this.min,this.max).multiplyScalar(.5)}getSize(t){return this.isEmpty()?t.set(0,0,0):t.subVectors(this.max,this.min)}expandByPoint(t){return this.min.min(t),this.max.max(t),this}expandByVector(t){return this.min.sub(t),this.max.add(t),this}expandByScalar(t){return this.min.addScalar(-t),this.max.addScalar(t),this}expandByObject(t,e=!1){t.updateWorldMatrix(!1,!1);const n=t.geometry;if(n!==void 0){const s=n.getAttribute("position");if(e===!0&&s!==void 0&&t.isInstancedMesh!==!0)for(let r=0,a=s.count;r<a;r++)t.isMesh===!0?t.getVertexPosition(r,z):z.fromBufferAttribute(s,r),z.applyMatrix4(t.matrixWorld),this.expandByPoint(z);else t.boundingBox!==void 0?(t.boundingBox===null&&t.computeBoundingBox(),q.copy(t.boundingBox)):(n.boundingBox===null&&n.computeBoundingBox(),q.copy(n.boundingBox)),q.applyMatrix4(t.matrixWorld),this.union(q)}const i=t.children;for(let s=0,r=i.length;s<r;s++)this.expandByObject(i[s],e);return this}containsPoint(t){return t.x>=this.min.x&&t.x<=this.max.x&&t.y>=this.min.y&&t.y<=this.max.y&&t.z>=this.min.z&&t.z<=this.max.z}containsBox(t){return this.min.x<=t.min.x&&t.max.x<=this.max.x&&this.min.y<=t.min.y&&t.max.y<=this.max.y&&this.min.z<=t.min.z&&t.max.z<=this.max.z}getParameter(t,e){return e.set((t.x-this.min.x)/(this.max.x-this.min.x),(t.y-this.min.y)/(this.max.y-this.min.y),(t.z-this.min.z)/(this.max.z-this.min.z))}intersectsBox(t){return t.max.x>=this.min.x&&t.min.x<=this.max.x&&t.max.y>=this.min.y&&t.min.y<=this.max.y&&t.max.z>=this.min.z&&t.min.z<=this.max.z}intersectsSphere(t){return this.clampPoint(t.center,z),z.distanceToSquared(t.center)<=t.radius*t.radius}intersectsPlane(t){let e,n;return t.normal.x>0?(e=t.normal.x*this.min.x,n=t.normal.x*this.max.x):(e=t.normal.x*this.max.x,n=t.normal.x*this.min.x),t.normal.y>0?(e+=t.normal.y*this.min.y,n+=t.normal.y*this.max.y):(e+=t.normal.y*this.max.y,n+=t.normal.y*this.min.y),t.normal.z>0?(e+=t.normal.z*this.min.z,n+=t.normal.z*this.max.z):(e+=t.normal.z*this.max.z,n+=t.normal.z*this.min.z),e<=-t.constant&&n>=-t.constant}intersectsTriangle(t){if(this.isEmpty())return!1;this.getCenter(Tt),Z.subVectors(this.max,Tt),H.subVectors(t.a,Tt),ot.subVectors(t.b,Tt),J.subVectors(t.c,Tt),st.subVectors(ot,H),R.subVectors(J,ot),U.subVectors(H,J);let e=[0,-st.z,st.y,0,-R.z,R.y,0,-U.z,U.y,st.z,0,-st.x,R.z,0,-R.x,U.z,0,-U.x,-st.y,st.x,0,-R.y,R.x,0,-U.y,U.x,0];return!gt(e,H,ot,J,Z)||(e=[1,0,0,0,1,0,0,0,1],!gt(e,H,ot,J,Z))?!1:(dt.crossVectors(st,R),e=[dt.x,dt.y,dt.z],gt(e,H,ot,J,Z))}clampPoint(t,e){return e.copy(t).clamp(this.min,this.max)}distanceToPoint(t){return this.clampPoint(t,z).distanceTo(t)}getBoundingSphere(t){return this.isEmpty()?t.makeEmpty():(this.getCenter(t.center),t.radius=this.getSize(z).length()*.5),t}intersect(t){return this.min.max(t.min),this.max.min(t.max),this.isEmpty()&&this.makeEmpty(),this}union(t){return this.min.min(t.min),this.max.max(t.max),this}applyMatrix4(t){return this.isEmpty()?this:(b[0].set(this.min.x,this.min.y,this.min.z).applyMatrix4(t),b[1].set(this.min.x,this.min.y,this.max.z).applyMatrix4(t),b[2].set(this.min.x,this.max.y,this.min.z).applyMatrix4(t),b[3].set(this.min.x,this.max.y,this.max.z).applyMatrix4(t),b[4].set(this.max.x,this.min.y,this.min.z).applyMatrix4(t),b[5].set(this.max.x,this.min.y,this.max.z).applyMatrix4(t),b[6].set(this.max.x,this.max.y,this.min.z).applyMatrix4(t),b[7].set(this.max.x,this.max.y,this.max.z).applyMatrix4(t),this.setFromPoints(b),this)}translate(t){return this.min.add(t),this.max.add(t),this}equals(t){return t.min.equals(this.min)&&t.max.equals(this.max)}toJSON(){return{min:this.min.toArray(),max:this.max.toArray()}}fromJSON(t){return this.min.fromArray(t.min),this.max.fromArray(t.max),this}}const b=[new O,new O,new O,new O,new O,new O,new O,new O],z=new O,q=new A,H=new O,ot=new O,J=new O,st=new O,R=new O,U=new O,Tt=new O,Z=new O,dt=new O,mt=new O;function gt(u,t,e,n,i){for(let s=0,r=u.length-3;s<=r;s+=3){mt.fromArray(u,s);const a=i.x*Math.abs(mt.x)+i.y*Math.abs(mt.y)+i.z*Math.abs(mt.z),l=t.dot(mt),c=e.dot(mt),d=n.dot(mt);if(Math.max(-Math.max(l,c,d),Math.min(l,c,d))>a)return!1}return!0}const _t=ct();function ct(){const u=new ArrayBuffer(4),t=new Float32Array(u),e=new Uint32Array(u),n=new Uint32Array(512),i=new Uint32Array(512);for(let l=0;l<256;++l){const c=l-127;c<-27?(n[l]=0,n[l|256]=32768,i[l]=24,i[l|256]=24):c<-14?(n[l]=1024>>-c-14,n[l|256]=1024>>-c-14|32768,i[l]=-c-1,i[l|256]=-c-1):c<=15?(n[l]=c+15<<10,n[l|256]=c+15<<10|32768,i[l]=13,i[l|256]=13):c<128?(n[l]=31744,n[l|256]=64512,i[l]=24,i[l|256]=24):(n[l]=31744,n[l|256]=64512,i[l]=13,i[l|256]=13)}const s=new Uint32Array(2048),r=new Uint32Array(64),a=new Uint32Array(64);for(let l=1;l<1024;++l){let c=l<<13,d=0;for(;(c&8388608)===0;)c<<=1,d-=8388608;c&=-8388609,d+=947912704,s[l]=c|d}for(let l=1024;l<2048;++l)s[l]=939524096+(l-1024<<13);for(let l=1;l<31;++l)r[l]=l<<23;r[31]=1199570944,r[32]=2147483648;for(let l=33;l<63;++l)r[l]=2147483648+(l-32<<23);r[63]=3347054592;for(let l=1;l<64;++l)l!==32&&(a[l]=1024);return{floatView:t,uint32View:e,baseTable:n,shiftTable:i,mantissaTable:s,exponentTable:r,offsetTable:a}}function it(u){Math.abs(u)>65504&&le("DataUtils.toHalfFloat(): Value out of range."),u=pe(u,-65504,65504),_t.floatView[0]=u;const t=_t.uint32View[0],e=t>>23&511;return _t.baseTable[e]+((t&8388607)>>_t.shiftTable[e])}function It(u){const t=u>>10;return _t.uint32View[0]=_t.mantissaTable[_t.offsetTable[t]+(u&1023)]+_t.exponentTable[t],_t.floatView[0]}class zt{static toHalfFloat(t){return it(t)}static fromHalfFloat(t){return It(t)}}const kt=new O,ue=new At;let te=0;class se{constructor(t,e,n=!1){if(Array.isArray(t))throw new TypeError("THREE.BufferAttribute: array should be a Typed Array.");this.isBufferAttribute=!0,Object.defineProperty(this,"id",{value:te++}),this.name="",this.array=t,this.itemSize=e,this.count=t!==void 0?t.length/e:0,this.normalized=n,this.usage=ji,this.updateRanges=[],this.gpuType=gi,this.version=0}onUploadCallback(){}set needsUpdate(t){t===!0&&this.version++}setUsage(t){return this.usage=t,this}addUpdateRange(t,e){this.updateRanges.push({start:t,count:e})}clearUpdateRanges(){this.updateRanges.length=0}copy(t){return this.name=t.name,this.array=new t.array.constructor(t.array),this.itemSize=t.itemSize,this.count=t.count,this.normalized=t.normalized,this.usage=t.usage,this.gpuType=t.gpuType,this}copyAt(t,e,n){t*=this.itemSize,n*=e.itemSize;for(let i=0,s=this.itemSize;i<s;i++)this.array[t+i]=e.array[n+i];return this}copyArray(t){return this.array.set(t),this}applyMatrix3(t){if(this.itemSize===2)for(let e=0,n=this.count;e<n;e++)ue.fromBufferAttribute(this,e),ue.applyMatrix3(t),this.setXY(e,ue.x,ue.y);else if(this.itemSize===3)for(let e=0,n=this.count;e<n;e++)kt.fromBufferAttribute(this,e),kt.applyMatrix3(t),this.setXYZ(e,kt.x,kt.y,kt.z);return this}applyMatrix4(t){for(let e=0,n=this.count;e<n;e++)kt.fromBufferAttribute(this,e),kt.applyMatrix4(t),this.setXYZ(e,kt.x,kt.y,kt.z);return this}applyNormalMatrix(t){for(let e=0,n=this.count;e<n;e++)kt.fromBufferAttribute(this,e),kt.applyNormalMatrix(t),this.setXYZ(e,kt.x,kt.y,kt.z);return this}transformDirection(t){for(let e=0,n=this.count;e<n;e++)kt.fromBufferAttribute(this,e),kt.transformDirection(t),this.setXYZ(e,kt.x,kt.y,kt.z);return this}set(t,e=0){return this.array.set(t,e),this}getComponent(t,e){let n=this.array[t*this.itemSize+e];return this.normalized&&(n=gn(n,this.array)),n}setComponent(t,e,n){return this.normalized&&(n=ye(n,this.array)),this.array[t*this.itemSize+e]=n,this}getX(t){let e=this.array[t*this.itemSize];return this.normalized&&(e=gn(e,this.array)),e}setX(t,e){return this.normalized&&(e=ye(e,this.array)),this.array[t*this.itemSize]=e,this}getY(t){let e=this.array[t*this.itemSize+1];return this.normalized&&(e=gn(e,this.array)),e}setY(t,e){return this.normalized&&(e=ye(e,this.array)),this.array[t*this.itemSize+1]=e,this}getZ(t){let e=this.array[t*this.itemSize+2];return this.normalized&&(e=gn(e,this.array)),e}setZ(t,e){return this.normalized&&(e=ye(e,this.array)),this.array[t*this.itemSize+2]=e,this}getW(t){let e=this.array[t*this.itemSize+3];return this.normalized&&(e=gn(e,this.array)),e}setW(t,e){return this.normalized&&(e=ye(e,this.array)),this.array[t*this.itemSize+3]=e,this}setXY(t,e,n){return t*=this.itemSize,this.normalized&&(e=ye(e,this.array),n=ye(n,this.array)),this.array[t+0]=e,this.array[t+1]=n,this}setXYZ(t,e,n,i){return t*=this.itemSize,this.normalized&&(e=ye(e,this.array),n=ye(n,this.array),i=ye(i,this.array)),this.array[t+0]=e,this.array[t+1]=n,this.array[t+2]=i,this}setXYZW(t,e,n,i,s){return t*=this.itemSize,this.normalized&&(e=ye(e,this.array),n=ye(n,this.array),i=ye(i,this.array),s=ye(s,this.array)),this.array[t+0]=e,this.array[t+1]=n,this.array[t+2]=i,this.array[t+3]=s,this}onUpload(t){return this.onUploadCallback=t,this}clone(){return new this.constructor(this.array,this.itemSize).copy(this)}toJSON(){const t={itemSize:this.itemSize,type:this.array.constructor.name,array:Array.from(this.array),normalized:this.normalized};return this.name!==""&&(t.name=this.name),this.usage!==ji&&(t.usage=this.usage),t}}class Ze extends se{constructor(t,e,n){super(new Int8Array(t),e,n)}}class $e extends se{constructor(t,e,n){super(new Uint8Array(t),e,n)}}class xt extends se{constructor(t,e,n){super(new Uint8ClampedArray(t),e,n)}}class Lt extends se{constructor(t,e,n){super(new Int16Array(t),e,n)}}class Dt extends se{constructor(t,e,n){super(new Uint16Array(t),e,n)}}class Te extends se{constructor(t,e,n){super(new Int32Array(t),e,n)}}class de extends se{constructor(t,e,n){super(new Uint32Array(t),e,n)}}class _e extends se{constructor(t,e,n){super(new Uint16Array(t),e,n),this.isFloat16BufferAttribute=!0}getX(t){let e=It(this.array[t*this.itemSize]);return this.normalized&&(e=gn(e,this.array)),e}setX(t,e){return this.normalized&&(e=ye(e,this.array)),this.array[t*this.itemSize]=it(e),this}getY(t){let e=It(this.array[t*this.itemSize+1]);return this.normalized&&(e=gn(e,this.array)),e}setY(t,e){return this.normalized&&(e=ye(e,this.array)),this.array[t*this.itemSize+1]=it(e),this}getZ(t){let e=It(this.array[t*this.itemSize+2]);return this.normalized&&(e=gn(e,this.array)),e}setZ(t,e){return this.normalized&&(e=ye(e,this.array)),this.array[t*this.itemSize+2]=it(e),this}getW(t){let e=It(this.array[t*this.itemSize+3]);return this.normalized&&(e=gn(e,this.array)),e}setW(t,e){return this.normalized&&(e=ye(e,this.array)),this.array[t*this.itemSize+3]=it(e),this}setXY(t,e,n){return t*=this.itemSize,this.normalized&&(e=ye(e,this.array),n=ye(n,this.array)),this.array[t+0]=it(e),this.array[t+1]=it(n),this}setXYZ(t,e,n,i){return t*=this.itemSize,this.normalized&&(e=ye(e,this.array),n=ye(n,this.array),i=ye(i,this.array)),this.array[t+0]=it(e),this.array[t+1]=it(n),this.array[t+2]=it(i),this}setXYZW(t,e,n,i,s){return t*=this.itemSize,this.normalized&&(e=ye(e,this.array),n=ye(n,this.array),i=ye(i,this.array),s=ye(s,this.array)),this.array[t+0]=it(e),this.array[t+1]=it(n),this.array[t+2]=it(i),this.array[t+3]=it(s),this}}class Ut extends se{constructor(t,e,n){super(new Float32Array(t),e,n)}}const Ie=new A,Pe=new O,Be=new O;class ae{constructor(t=new O,e=-1){this.isSphere=!0,this.center=t,this.radius=e}set(t,e){return this.center.copy(t),this.radius=e,this}setFromPoints(t,e){const n=this.center;e!==void 0?n.copy(e):Ie.setFromPoints(t).getCenter(n);let i=0;for(let s=0,r=t.length;s<r;s++)i=Math.max(i,n.distanceToSquared(t[s]));return this.radius=Math.sqrt(i),this}copy(t){return this.center.copy(t.center),this.radius=t.radius,this}isEmpty(){return this.radius<0}makeEmpty(){return this.center.set(0,0,0),this.radius=-1,this}containsPoint(t){return t.distanceToSquared(this.center)<=this.radius*this.radius}distanceToPoint(t){return t.distanceTo(this.center)-this.radius}intersectsSphere(t){const e=this.radius+t.radius;return t.center.distanceToSquared(this.center)<=e*e}intersectsBox(t){return t.intersectsSphere(this)}intersectsPlane(t){return Math.abs(t.distanceToPoint(this.center))<=this.radius}clampPoint(t,e){const n=this.center.distanceToSquared(t);return e.copy(t),n>this.radius*this.radius&&(e.sub(this.center).normalize(),e.multiplyScalar(this.radius).add(this.center)),e}getBoundingBox(t){return this.isEmpty()?(t.makeEmpty(),t):(t.set(this.center,this.center),t.expandByScalar(this.radius),t)}applyMatrix4(t){return this.center.applyMatrix4(t),this.radius=this.radius*t.getMaxScaleOnAxis(),this}translate(t){return this.center.add(t),this}expandByPoint(t){if(this.isEmpty())return this.center.copy(t),this.radius=0,this;Pe.subVectors(t,this.center);const e=Pe.lengthSq();if(e>this.radius*this.radius){const n=Math.sqrt(e),i=(n-this.radius)*.5;this.center.addScaledVector(Pe,i/n),this.radius+=i}return this}union(t){return t.isEmpty()?this:this.isEmpty()?(this.copy(t),this):(this.center.equals(t.center)===!0?this.radius=Math.max(this.radius,t.radius):(Be.subVectors(t.center,this.center).setLength(t.radius),this.expandByPoint(Pe.copy(t.center).add(Be)),this.expandByPoint(Pe.copy(t.center).sub(Be))),this)}equals(t){return t.center.equals(this.center)&&t.radius===this.radius}clone(){return new this.constructor().copy(this)}toJSON(){return{radius:this.radius,center:this.center.toArray()}}fromJSON(t){return this.radius=t.radius,this.center.fromArray(t.center),this}}let Je=0;const B=new Me,Xe=new De,Ce=new O,Ee=new A,Jt=new A,P=new O;class y extends Kn{constructor(){super(),this.isBufferGeometry=!0,Object.defineProperty(this,"id",{value:Je++}),this.uuid=An(),this.name="",this.type="BufferGeometry",this.index=null,this.indirect=null,this.indirectOffset=0,this.attributes={},this.morphAttributes={},this.morphTargetsRelative=!1,this.groups=[],this.boundingBox=null,this.boundingSphere=null,this.drawRange={start:0,count:1/0},this.userData={}}getIndex(){return this.index}setIndex(t){return Array.isArray(t)?this.index=new(cc(t)?de:Dt)(t,1):this.index=t,this}setIndirect(t,e=0){return this.indirect=t,this.indirectOffset=e,this}getIndirect(){return this.indirect}getAttribute(t){return this.attributes[t]}setAttribute(t,e){return this.attributes[t]=e,this}deleteAttribute(t){return delete this.attributes[t],this}hasAttribute(t){return this.attributes[t]!==void 0}addGroup(t,e,n=0){this.groups.push({start:t,count:e,materialIndex:n})}clearGroups(){this.groups=[]}setDrawRange(t,e){this.drawRange.start=t,this.drawRange.count=e}applyMatrix4(t){const e=this.attributes.position;e!==void 0&&(e.applyMatrix4(t),e.needsUpdate=!0);const n=this.attributes.normal;if(n!==void 0){const s=new Fn().getNormalMatrix(t);n.applyNormalMatrix(s),n.needsUpdate=!0}const i=this.attributes.tangent;return i!==void 0&&(i.transformDirection(t),i.needsUpdate=!0),this.boundingBox!==null&&this.computeBoundingBox(),this.boundingSphere!==null&&this.computeBoundingSphere(),this}applyQuaternion(t){return B.makeRotationFromQuaternion(t),this.applyMatrix4(B),this}rotateX(t){return B.makeRotationX(t),this.applyMatrix4(B),this}rotateY(t){return B.makeRotationY(t),this.applyMatrix4(B),this}rotateZ(t){return B.makeRotationZ(t),this.applyMatrix4(B),this}translate(t,e,n){return B.makeTranslation(t,e,n),this.applyMatrix4(B),this}scale(t,e,n){return B.makeScale(t,e,n),this.applyMatrix4(B),this}lookAt(t){return Xe.lookAt(t),Xe.updateMatrix(),this.applyMatrix4(Xe.matrix),this}center(){return this.computeBoundingBox(),this.boundingBox.getCenter(Ce).negate(),this.translate(Ce.x,Ce.y,Ce.z),this}setFromPoints(t){const e=this.getAttribute("position");if(e===void 0){const n=[];for(let i=0,s=t.length;i<s;i++){const r=t[i];n.push(r.x,r.y,r.z||0)}this.setAttribute("position",new Ut(n,3))}else{const n=Math.min(t.length,e.count);for(let i=0;i<n;i++){const s=t[i];e.setXYZ(i,s.x,s.y,s.z||0)}t.length>e.count&&le("BufferGeometry: Buffer size too small for points data. Use .dispose() and create a new geometry."),e.needsUpdate=!0}return this}computeBoundingBox(){this.boundingBox===null&&(this.boundingBox=new A);const t=this.attributes.position,e=this.morphAttributes.position;if(t&&t.isGLBufferAttribute){Re("BufferGeometry.computeBoundingBox(): GLBufferAttribute requires a manual bounding box.",this),this.boundingBox.set(new O(-1/0,-1/0,-1/0),new O(1/0,1/0,1/0));return}if(t!==void 0){if(this.boundingBox.setFromBufferAttribute(t),e)for(let n=0,i=e.length;n<i;n++){const s=e[n];Ee.setFromBufferAttribute(s),this.morphTargetsRelative?(P.addVectors(this.boundingBox.min,Ee.min),this.boundingBox.expandByPoint(P),P.addVectors(this.boundingBox.max,Ee.max),this.boundingBox.expandByPoint(P)):(this.boundingBox.expandByPoint(Ee.min),this.boundingBox.expandByPoint(Ee.max))}}else this.boundingBox.makeEmpty();(isNaN(this.boundingBox.min.x)||isNaN(this.boundingBox.min.y)||isNaN(this.boundingBox.min.z))&&Re('BufferGeometry.computeBoundingBox(): Computed min/max have NaN values. The "position" attribute is likely to have NaN values.',this)}computeBoundingSphere(){this.boundingSphere===null&&(this.boundingSphere=new ae);const t=this.attributes.position,e=this.morphAttributes.position;if(t&&t.isGLBufferAttribute){Re("BufferGeometry.computeBoundingSphere(): GLBufferAttribute requires a manual bounding sphere.",this),this.boundingSphere.set(new O,1/0);return}if(t){const n=this.boundingSphere.center;if(Ee.setFromBufferAttribute(t),e)for(let s=0,r=e.length;s<r;s++){const a=e[s];Jt.setFromBufferAttribute(a),this.morphTargetsRelative?(P.addVectors(Ee.min,Jt.min),Ee.expandByPoint(P),P.addVectors(Ee.max,Jt.max),Ee.expandByPoint(P)):(Ee.expandByPoint(Jt.min),Ee.expandByPoint(Jt.max))}Ee.getCenter(n);let i=0;for(let s=0,r=t.count;s<r;s++)P.fromBufferAttribute(t,s),i=Math.max(i,n.distanceToSquared(P));if(e)for(let s=0,r=e.length;s<r;s++){const a=e[s],l=this.morphTargetsRelative;for(let c=0,d=a.count;c<d;c++)P.fromBufferAttribute(a,c),l&&(Ce.fromBufferAttribute(t,c),P.add(Ce)),i=Math.max(i,n.distanceToSquared(P))}this.boundingSphere.radius=Math.sqrt(i),isNaN(this.boundingSphere.radius)&&Re('BufferGeometry.computeBoundingSphere(): Computed radius is NaN. The "position" attribute is likely to have NaN values.',this)}}computeTangents(){const t=this.index,e=this.attributes;if(t===null||e.position===void 0||e.normal===void 0||e.uv===void 0){Re("BufferGeometry: .computeTangents() failed. Missing required attributes (index, position, normal or uv)");return}const n=e.position,i=e.normal,s=e.uv;this.hasAttribute("tangent")===!1&&this.setAttribute("tangent",new se(new Float32Array(4*n.count),4));const r=this.getAttribute("tangent"),a=[],l=[];for(let at=0;at<n.count;at++)a[at]=new O,l[at]=new O;const c=new O,d=new O,m=new O,g=new At,_=new At,v=new At,M=new O,C=new O;function w(at,Mt,St){c.fromBufferAttribute(n,at),d.fromBufferAttribute(n,Mt),m.fromBufferAttribute(n,St),g.fromBufferAttribute(s,at),_.fromBufferAttribute(s,Mt),v.fromBufferAttribute(s,St),d.sub(c),m.sub(c),_.sub(g),v.sub(g);const Gt=1/(_.x*v.y-v.x*_.y);isFinite(Gt)&&(M.copy(d).multiplyScalar(v.y).addScaledVector(m,-_.y).multiplyScalar(Gt),C.copy(m).multiplyScalar(_.x).addScaledVector(d,-v.x).multiplyScalar(Gt),a[at].add(M),a[Mt].add(M),a[St].add(M),l[at].add(C),l[Mt].add(C),l[St].add(C))}let L=this.groups;L.length===0&&(L=[{start:0,count:t.count}]);for(let at=0,Mt=L.length;at<Mt;++at){const St=L[at],Gt=St.start,he=St.count;for(let ge=Gt,Ne=Gt+he;ge<Ne;ge+=3)w(t.getX(ge+0),t.getX(ge+1),t.getX(ge+2))}const D=new O,N=new O,et=new O,rt=new O;function vt(at){et.fromBufferAttribute(i,at),rt.copy(et);const Mt=a[at];D.copy(Mt),D.sub(et.multiplyScalar(et.dot(Mt))).normalize(),N.crossVectors(rt,Mt);const Gt=N.dot(l[at])<0?-1:1;r.setXYZW(at,D.x,D.y,D.z,Gt)}for(let at=0,Mt=L.length;at<Mt;++at){const St=L[at],Gt=St.start,he=St.count;for(let ge=Gt,Ne=Gt+he;ge<Ne;ge+=3)vt(t.getX(ge+0)),vt(t.getX(ge+1)),vt(t.getX(ge+2))}}computeVertexNormals(){const t=this.index,e=this.getAttribute("position");if(e!==void 0){let n=this.getAttribute("normal");if(n===void 0)n=new se(new Float32Array(e.count*3),3),this.setAttribute("normal",n);else for(let g=0,_=n.count;g<_;g++)n.setXYZ(g,0,0,0);const i=new O,s=new O,r=new O,a=new O,l=new O,c=new O,d=new O,m=new O;if(t)for(let g=0,_=t.count;g<_;g+=3){const v=t.getX(g+0),M=t.getX(g+1),C=t.getX(g+2);i.fromBufferAttribute(e,v),s.fromBufferAttribute(e,M),r.fromBufferAttribute(e,C),d.subVectors(r,s),m.subVectors(i,s),d.cross(m),a.fromBufferAttribute(n,v),l.fromBufferAttribute(n,M),c.fromBufferAttribute(n,C),a.add(d),l.add(d),c.add(d),n.setXYZ(v,a.x,a.y,a.z),n.setXYZ(M,l.x,l.y,l.z),n.setXYZ(C,c.x,c.y,c.z)}else for(let g=0,_=e.count;g<_;g+=3)i.fromBufferAttribute(e,g+0),s.fromBufferAttribute(e,g+1),r.fromBufferAttribute(e,g+2),d.subVectors(r,s),m.subVectors(i,s),d.cross(m),n.setXYZ(g+0,d.x,d.y,d.z),n.setXYZ(g+1,d.x,d.y,d.z),n.setXYZ(g+2,d.x,d.y,d.z);this.normalizeNormals(),n.needsUpdate=!0}}normalizeNormals(){const t=this.attributes.normal;for(let e=0,n=t.count;e<n;e++)P.fromBufferAttribute(t,e),P.normalize(),t.setXYZ(e,P.x,P.y,P.z)}toNonIndexed(){function t(a,l){const c=a.array,d=a.itemSize,m=a.normalized,g=new c.constructor(l.length*d);let _=0,v=0;for(let M=0,C=l.length;M<C;M++){a.isInterleavedBufferAttribute?_=l[M]*a.data.stride+a.offset:_=l[M]*d;for(let w=0;w<d;w++)g[v++]=c[_++]}return new se(g,d,m)}if(this.index===null)return le("BufferGeometry.toNonIndexed(): BufferGeometry is already non-indexed."),this;const e=new y,n=this.index.array,i=this.attributes;for(const a in i){const l=i[a],c=t(l,n);e.setAttribute(a,c)}const s=this.morphAttributes;for(const a in s){const l=[],c=s[a];for(let d=0,m=c.length;d<m;d++){const g=c[d],_=t(g,n);l.push(_)}e.morphAttributes[a]=l}e.morphTargetsRelative=this.morphTargetsRelative;const r=this.groups;for(let a=0,l=r.length;a<l;a++){const c=r[a];e.addGroup(c.start,c.count,c.materialIndex)}return e}toJSON(){const t={metadata:{version:4.7,type:"BufferGeometry",generator:"BufferGeometry.toJSON"}};if(t.uuid=this.uuid,t.type=this.type,this.name!==""&&(t.name=this.name),Object.keys(this.userData).length>0&&(t.userData=this.userData),this.parameters!==void 0){const l=this.parameters;for(const c in l)l[c]!==void 0&&(t[c]=l[c]);return t}t.data={attributes:{}};const e=this.index;e!==null&&(t.data.index={type:e.array.constructor.name,array:Array.prototype.slice.call(e.array)});const n=this.attributes;for(const l in n){const c=n[l];t.data.attributes[l]=c.toJSON(t.data)}const i={};let s=!1;for(const l in this.morphAttributes){const c=this.morphAttributes[l],d=[];for(let m=0,g=c.length;m<g;m++){const _=c[m];d.push(_.toJSON(t.data))}d.length>0&&(i[l]=d,s=!0)}s&&(t.data.morphAttributes=i,t.data.morphTargetsRelative=this.morphTargetsRelative);const r=this.groups;r.length>0&&(t.data.groups=JSON.parse(JSON.stringify(r)));const a=this.boundingSphere;return a!==null&&(t.data.boundingSphere=a.toJSON()),t}clone(){return new this.constructor().copy(this)}copy(t){this.index=null,this.attributes={},this.morphAttributes={},this.groups=[],this.boundingBox=null,this.boundingSphere=null;const e={};this.name=t.name;const n=t.index;n!==null&&this.setIndex(n.clone());const i=t.attributes;for(const c in i){const d=i[c];this.setAttribute(c,d.clone(e))}const s=t.morphAttributes;for(const c in s){const d=[],m=s[c];for(let g=0,_=m.length;g<_;g++)d.push(m[g].clone(e));this.morphAttributes[c]=d}this.morphTargetsRelative=t.morphTargetsRelative;const r=t.groups;for(let c=0,d=r.length;c<d;c++){const m=r[c];this.addGroup(m.start,m.count,m.materialIndex)}const a=t.boundingBox;a!==null&&(this.boundingBox=a.clone());const l=t.boundingSphere;return l!==null&&(this.boundingSphere=l.clone()),this.drawRange.start=t.drawRange.start,this.drawRange.count=t.drawRange.count,this.userData=t.userData,this}dispose(){this.dispatchEvent({type:"dispose"})}}class W{constructor(t,e){this.isInterleavedBuffer=!0,this.array=t,this.stride=e,this.count=t!==void 0?t.length/e:0,this.usage=ji,this.updateRanges=[],this.version=0,this.uuid=An()}onUploadCallback(){}set needsUpdate(t){t===!0&&this.version++}setUsage(t){return this.usage=t,this}addUpdateRange(t,e){this.updateRanges.push({start:t,count:e})}clearUpdateRanges(){this.updateRanges.length=0}copy(t){return this.array=new t.array.constructor(t.array),this.count=t.count,this.stride=t.stride,this.usage=t.usage,this}copyAt(t,e,n){t*=this.stride,n*=e.stride;for(let i=0,s=this.stride;i<s;i++)this.array[t+i]=e.array[n+i];return this}set(t,e=0){return this.array.set(t,e),this}clone(t){t.arrayBuffers===void 0&&(t.arrayBuffers={}),this.array.buffer._uuid===void 0&&(this.array.buffer._uuid=An()),t.arrayBuffers[this.array.buffer._uuid]===void 0&&(t.arrayBuffers[this.array.buffer._uuid]=this.array.slice(0).buffer);const e=new this.array.constructor(t.arrayBuffers[this.array.buffer._uuid]),n=new this.constructor(e,this.stride);return n.setUsage(this.usage),n}onUpload(t){return this.onUploadCallback=t,this}toJSON(t){return t.arrayBuffers===void 0&&(t.arrayBuffers={}),this.array.buffer._uuid===void 0&&(this.array.buffer._uuid=An()),t.arrayBuffers[this.array.buffer._uuid]===void 0&&(t.arrayBuffers[this.array.buffer._uuid]=Array.from(new Uint32Array(this.array.buffer))),{uuid:this.uuid,buffer:this.array.buffer._uuid,type:this.array.constructor.name,stride:this.stride}}}const ut=new O;class yt{constructor(t,e,n,i=!1){this.isInterleavedBufferAttribute=!0,this.name="",this.data=t,this.itemSize=e,this.offset=n,this.normalized=i}get count(){return this.data.count}get array(){return this.data.array}set needsUpdate(t){this.data.needsUpdate=t}applyMatrix4(t){for(let e=0,n=this.data.count;e<n;e++)ut.fromBufferAttribute(this,e),ut.applyMatrix4(t),this.setXYZ(e,ut.x,ut.y,ut.z);return this}applyNormalMatrix(t){for(let e=0,n=this.count;e<n;e++)ut.fromBufferAttribute(this,e),ut.applyNormalMatrix(t),this.setXYZ(e,ut.x,ut.y,ut.z);return this}transformDirection(t){for(let e=0,n=this.count;e<n;e++)ut.fromBufferAttribute(this,e),ut.transformDirection(t),this.setXYZ(e,ut.x,ut.y,ut.z);return this}getComponent(t,e){let n=this.array[t*this.data.stride+this.offset+e];return this.normalized&&(n=gn(n,this.array)),n}setComponent(t,e,n){return this.normalized&&(n=ye(n,this.array)),this.data.array[t*this.data.stride+this.offset+e]=n,this}setX(t,e){return this.normalized&&(e=ye(e,this.array)),this.data.array[t*this.data.stride+this.offset]=e,this}setY(t,e){return this.normalized&&(e=ye(e,this.array)),this.data.array[t*this.data.stride+this.offset+1]=e,this}setZ(t,e){return this.normalized&&(e=ye(e,this.array)),this.data.array[t*this.data.stride+this.offset+2]=e,this}setW(t,e){return this.normalized&&(e=ye(e,this.array)),this.data.array[t*this.data.stride+this.offset+3]=e,this}getX(t){let e=this.data.array[t*this.data.stride+this.offset];return this.normalized&&(e=gn(e,this.array)),e}getY(t){let e=this.data.array[t*this.data.stride+this.offset+1];return this.normalized&&(e=gn(e,this.array)),e}getZ(t){let e=this.data.array[t*this.data.stride+this.offset+2];return this.normalized&&(e=gn(e,this.array)),e}getW(t){let e=this.data.array[t*this.data.stride+this.offset+3];return this.normalized&&(e=gn(e,this.array)),e}setXY(t,e,n){return t=t*this.data.stride+this.offset,this.normalized&&(e=ye(e,this.array),n=ye(n,this.array)),this.data.array[t+0]=e,this.data.array[t+1]=n,this}setXYZ(t,e,n,i){return t=t*this.data.stride+this.offset,this.normalized&&(e=ye(e,this.array),n=ye(n,this.array),i=ye(i,this.array)),this.data.array[t+0]=e,this.data.array[t+1]=n,this.data.array[t+2]=i,this}setXYZW(t,e,n,i,s){return t=t*this.data.stride+this.offset,this.normalized&&(e=ye(e,this.array),n=ye(n,this.array),i=ye(i,this.array),s=ye(s,this.array)),this.data.array[t+0]=e,this.data.array[t+1]=n,this.data.array[t+2]=i,this.data.array[t+3]=s,this}clone(t){if(t===void 0){gr("InterleavedBufferAttribute.clone(): Cloning an interleaved buffer attribute will de-interleave buffer data.");const e=[];for(let n=0;n<this.count;n++){const i=n*this.data.stride+this.offset;for(let s=0;s<this.itemSize;s++)e.push(this.data.array[i+s])}return new se(new this.array.constructor(e),this.itemSize,this.normalized)}else return t.interleavedBuffers===void 0&&(t.interleavedBuffers={}),t.interleavedBuffers[this.data.uuid]===void 0&&(t.interleavedBuffers[this.data.uuid]=this.data.clone(t)),new yt(t.interleavedBuffers[this.data.uuid],this.itemSize,this.offset,this.normalized)}toJSON(t){if(t===void 0){gr("InterleavedBufferAttribute.toJSON(): Serializing an interleaved buffer attribute will de-interleave buffer data.");const e=[];for(let n=0;n<this.count;n++){const i=n*this.data.stride+this.offset;for(let s=0;s<this.itemSize;s++)e.push(this.data.array[i+s])}return{itemSize:this.itemSize,type:this.array.constructor.name,array:e,normalized:this.normalized}}else return t.interleavedBuffers===void 0&&(t.interleavedBuffers={}),t.interleavedBuffers[this.data.uuid]===void 0&&(t.interleavedBuffers[this.data.uuid]=this.data.toJSON(t)),{isInterleavedBufferAttribute:!0,itemSize:this.itemSize,data:this.data.uuid,offset:this.offset,normalized:this.normalized}}}let pt=0;class Ft extends Kn{constructor(){super(),this.isMaterial=!0,Object.defineProperty(this,"id",{value:pt++}),this.uuid=An(),this.name="",this.type="Material",this.blending=Ks,this.side=As,this.vertexColors=!1,this.opacity=1,this.transparent=!1,this.alphaHash=!1,this.blendSrc=js,this.blendDst=tr,this.blendEquation=Qs,this.blendSrcAlpha=null,this.blendDstAlpha=null,this.blendEquationAlpha=null,this.blendColor=new re(0,0,0),this.blendAlpha=0,this.depthFunc=qi,this.depthTest=!0,this.depthWrite=!0,this.stencilWriteMask=255,this.stencilFunc=Aa,this.stencilRef=0,this.stencilFuncMask=255,this.stencilFail=Ii,this.stencilZFail=Ii,this.stencilZPass=Ii,this.stencilWrite=!1,this.clippingPlanes=null,this.clipIntersection=!1,this.clipShadows=!1,this.shadowSide=null,this.colorWrite=!0,this.precision=null,this.polygonOffset=!1,this.polygonOffsetFactor=0,this.polygonOffsetUnits=0,this.dithering=!1,this.alphaToCoverage=!1,this.premultipliedAlpha=!1,this.forceSinglePass=!1,this.allowOverride=!0,this.visible=!0,this.toneMapped=!0,this.userData={},this.version=0,this._alphaTest=0}get alphaTest(){return this._alphaTest}set alphaTest(t){this._alphaTest>0!=t>0&&this.version++,this._alphaTest=t}onBeforeRender(){}onBeforeCompile(){}customProgramCacheKey(){return this.onBeforeCompile.toString()}setValues(t){if(t!==void 0)for(const e in t){const n=t[e];if(n===void 0){le(`Material: parameter '${e}' has value of undefined.`);continue}const i=this[e];if(i===void 0){le(`Material: '${e}' is not a property of THREE.${this.type}.`);continue}i&&i.isColor?i.set(n):i&&i.isVector3&&n&&n.isVector3?i.copy(n):this[e]=n}}toJSON(t){const e=t===void 0||typeof t=="string";e&&(t={textures:{},images:{}});const n={metadata:{version:4.7,type:"Material",generator:"Material.toJSON"}};n.uuid=this.uuid,n.type=this.type,this.name!==""&&(n.name=this.name),this.color&&this.color.isColor&&(n.color=this.color.getHex()),this.roughness!==void 0&&(n.roughness=this.roughness),this.metalness!==void 0&&(n.metalness=this.metalness),this.sheen!==void 0&&(n.sheen=this.sheen),this.sheenColor&&this.sheenColor.isColor&&(n.sheenColor=this.sheenColor.getHex()),this.sheenRoughness!==void 0&&(n.sheenRoughness=this.sheenRoughness),this.emissive&&this.emissive.isColor&&(n.emissive=this.emissive.getHex()),this.emissiveIntensity!==void 0&&this.emissiveIntensity!==1&&(n.emissiveIntensity=this.emissiveIntensity),this.specular&&this.specular.isColor&&(n.specular=this.specular.getHex()),this.specularIntensity!==void 0&&(n.specularIntensity=this.specularIntensity),this.specularColor&&this.specularColor.isColor&&(n.specularColor=this.specularColor.getHex()),this.shininess!==void 0&&(n.shininess=this.shininess),this.clearcoat!==void 0&&(n.clearcoat=this.clearcoat),this.clearcoatRoughness!==void 0&&(n.clearcoatRoughness=this.clearcoatRoughness),this.clearcoatMap&&this.clearcoatMap.isTexture&&(n.clearcoatMap=this.clearcoatMap.toJSON(t).uuid),this.clearcoatRoughnessMap&&this.clearcoatRoughnessMap.isTexture&&(n.clearcoatRoughnessMap=this.clearcoatRoughnessMap.toJSON(t).uuid),this.clearcoatNormalMap&&this.clearcoatNormalMap.isTexture&&(n.clearcoatNormalMap=this.clearcoatNormalMap.toJSON(t).uuid,n.clearcoatNormalScale=this.clearcoatNormalScale.toArray()),this.sheenColorMap&&this.sheenColorMap.isTexture&&(n.sheenColorMap=this.sheenColorMap.toJSON(t).uuid),this.sheenRoughnessMap&&this.sheenRoughnessMap.isTexture&&(n.sheenRoughnessMap=this.sheenRoughnessMap.toJSON(t).uuid),this.dispersion!==void 0&&(n.dispersion=this.dispersion),this.iridescence!==void 0&&(n.iridescence=this.iridescence),this.iridescenceIOR!==void 0&&(n.iridescenceIOR=this.iridescenceIOR),this.iridescenceThicknessRange!==void 0&&(n.iridescenceThicknessRange=this.iridescenceThicknessRange),this.iridescenceMap&&this.iridescenceMap.isTexture&&(n.iridescenceMap=this.iridescenceMap.toJSON(t).uuid),this.iridescenceThicknessMap&&this.iridescenceThicknessMap.isTexture&&(n.iridescenceThicknessMap=this.iridescenceThicknessMap.toJSON(t).uuid),this.anisotropy!==void 0&&(n.anisotropy=this.anisotropy),this.anisotropyRotation!==void 0&&(n.anisotropyRotation=this.anisotropyRotation),this.anisotropyMap&&this.anisotropyMap.isTexture&&(n.anisotropyMap=this.anisotropyMap.toJSON(t).uuid),this.map&&this.map.isTexture&&(n.map=this.map.toJSON(t).uuid),this.matcap&&this.matcap.isTexture&&(n.matcap=this.matcap.toJSON(t).uuid),this.alphaMap&&this.alphaMap.isTexture&&(n.alphaMap=this.alphaMap.toJSON(t).uuid),this.lightMap&&this.lightMap.isTexture&&(n.lightMap=this.lightMap.toJSON(t).uuid,n.lightMapIntensity=this.lightMapIntensity),this.aoMap&&this.aoMap.isTexture&&(n.aoMap=this.aoMap.toJSON(t).uuid,n.aoMapIntensity=this.aoMapIntensity),this.bumpMap&&this.bumpMap.isTexture&&(n.bumpMap=this.bumpMap.toJSON(t).uuid,n.bumpScale=this.bumpScale),this.normalMap&&this.normalMap.isTexture&&(n.normalMap=this.normalMap.toJSON(t).uuid,n.normalMapType=this.normalMapType,n.normalScale=this.normalScale.toArray()),this.displacementMap&&this.displacementMap.isTexture&&(n.displacementMap=this.displacementMap.toJSON(t).uuid,n.displacementScale=this.displacementScale,n.displacementBias=this.displacementBias),this.roughnessMap&&this.roughnessMap.isTexture&&(n.roughnessMap=this.roughnessMap.toJSON(t).uuid),this.metalnessMap&&this.metalnessMap.isTexture&&(n.metalnessMap=this.metalnessMap.toJSON(t).uuid),this.emissiveMap&&this.emissiveMap.isTexture&&(n.emissiveMap=this.emissiveMap.toJSON(t).uuid),this.specularMap&&this.specularMap.isTexture&&(n.specularMap=this.specularMap.toJSON(t).uuid),this.specularIntensityMap&&this.specularIntensityMap.isTexture&&(n.specularIntensityMap=this.specularIntensityMap.toJSON(t).uuid),this.specularColorMap&&this.specularColorMap.isTexture&&(n.specularColorMap=this.specularColorMap.toJSON(t).uuid),this.envMap&&this.envMap.isTexture&&(n.envMap=this.envMap.toJSON(t).uuid,this.combine!==void 0&&(n.combine=this.combine)),this.envMapRotation!==void 0&&(n.envMapRotation=this.envMapRotation.toArray()),this.envMapIntensity!==void 0&&(n.envMapIntensity=this.envMapIntensity),this.reflectivity!==void 0&&(n.reflectivity=this.reflectivity),this.refractionRatio!==void 0&&(n.refractionRatio=this.refractionRatio),this.gradientMap&&this.gradientMap.isTexture&&(n.gradientMap=this.gradientMap.toJSON(t).uuid),this.transmission!==void 0&&(n.transmission=this.transmission),this.transmissionMap&&this.transmissionMap.isTexture&&(n.transmissionMap=this.transmissionMap.toJSON(t).uuid),this.thickness!==void 0&&(n.thickness=this.thickness),this.thicknessMap&&this.thicknessMap.isTexture&&(n.thicknessMap=this.thicknessMap.toJSON(t).uuid),this.attenuationDistance!==void 0&&this.attenuationDistance!==1/0&&(n.attenuationDistance=this.attenuationDistance),this.attenuationColor!==void 0&&(n.attenuationColor=this.attenuationColor.getHex()),this.size!==void 0&&(n.size=this.size),this.shadowSide!==null&&(n.shadowSide=this.shadowSide),this.sizeAttenuation!==void 0&&(n.sizeAttenuation=this.sizeAttenuation),this.blending!==Ks&&(n.blending=this.blending),this.side!==As&&(n.side=this.side),this.vertexColors===!0&&(n.vertexColors=!0),this.opacity<1&&(n.opacity=this.opacity),this.transparent===!0&&(n.transparent=!0),this.blendSrc!==js&&(n.blendSrc=this.blendSrc),this.blendDst!==tr&&(n.blendDst=this.blendDst),this.blendEquation!==Qs&&(n.blendEquation=this.blendEquation),this.blendSrcAlpha!==null&&(n.blendSrcAlpha=this.blendSrcAlpha),this.blendDstAlpha!==null&&(n.blendDstAlpha=this.blendDstAlpha),this.blendEquationAlpha!==null&&(n.blendEquationAlpha=this.blendEquationAlpha),this.blendColor&&this.blendColor.isColor&&(n.blendColor=this.blendColor.getHex()),this.blendAlpha!==0&&(n.blendAlpha=this.blendAlpha),this.depthFunc!==qi&&(n.depthFunc=this.depthFunc),this.depthTest===!1&&(n.depthTest=this.depthTest),this.depthWrite===!1&&(n.depthWrite=this.depthWrite),this.colorWrite===!1&&(n.colorWrite=this.colorWrite),this.stencilWriteMask!==255&&(n.stencilWriteMask=this.stencilWriteMask),this.stencilFunc!==Aa&&(n.stencilFunc=this.stencilFunc),this.stencilRef!==0&&(n.stencilRef=this.stencilRef),this.stencilFuncMask!==255&&(n.stencilFuncMask=this.stencilFuncMask),this.stencilFail!==Ii&&(n.stencilFail=this.stencilFail),this.stencilZFail!==Ii&&(n.stencilZFail=this.stencilZFail),this.stencilZPass!==Ii&&(n.stencilZPass=this.stencilZPass),this.stencilWrite===!0&&(n.stencilWrite=this.stencilWrite),this.rotation!==void 0&&this.rotation!==0&&(n.rotation=this.rotation),this.polygonOffset===!0&&(n.polygonOffset=!0),this.polygonOffsetFactor!==0&&(n.polygonOffsetFactor=this.polygonOffsetFactor),this.polygonOffsetUnits!==0&&(n.polygonOffsetUnits=this.polygonOffsetUnits),this.linewidth!==void 0&&this.linewidth!==1&&(n.linewidth=this.linewidth),this.dashSize!==void 0&&(n.dashSize=this.dashSize),this.gapSize!==void 0&&(n.gapSize=this.gapSize),this.scale!==void 0&&(n.scale=this.scale),this.dithering===!0&&(n.dithering=!0),this.alphaTest>0&&(n.alphaTest=this.alphaTest),this.alphaHash===!0&&(n.alphaHash=!0),this.alphaToCoverage===!0&&(n.alphaToCoverage=!0),this.premultipliedAlpha===!0&&(n.premultipliedAlpha=!0),this.forceSinglePass===!0&&(n.forceSinglePass=!0),this.allowOverride===!1&&(n.allowOverride=!1),this.wireframe===!0&&(n.wireframe=!0),this.wireframeLinewidth>1&&(n.wireframeLinewidth=this.wireframeLinewidth),this.wireframeLinecap!=="round"&&(n.wireframeLinecap=this.wireframeLinecap),this.wireframeLinejoin!=="round"&&(n.wireframeLinejoin=this.wireframeLinejoin),this.flatShading===!0&&(n.flatShading=!0),this.visible===!1&&(n.visible=!1),this.toneMapped===!1&&(n.toneMapped=!1),this.fog===!1&&(n.fog=!1),Object.keys(this.userData).length>0&&(n.userData=this.userData);function i(s){const r=[];for(const a in s){const l=s[a];delete l.metadata,r.push(l)}return r}if(e){const s=i(t.textures),r=i(t.images);s.length>0&&(n.textures=s),r.length>0&&(n.images=r)}return n}clone(){return new this.constructor().copy(this)}copy(t){this.name=t.name,this.blending=t.blending,this.side=t.side,this.vertexColors=t.vertexColors,this.opacity=t.opacity,this.transparent=t.transparent,this.blendSrc=t.blendSrc,this.blendDst=t.blendDst,this.blendEquation=t.blendEquation,this.blendSrcAlpha=t.blendSrcAlpha,this.blendDstAlpha=t.blendDstAlpha,this.blendEquationAlpha=t.blendEquationAlpha,this.blendColor.copy(t.blendColor),this.blendAlpha=t.blendAlpha,this.depthFunc=t.depthFunc,this.depthTest=t.depthTest,this.depthWrite=t.depthWrite,this.stencilWriteMask=t.stencilWriteMask,this.stencilFunc=t.stencilFunc,this.stencilRef=t.stencilRef,this.stencilFuncMask=t.stencilFuncMask,this.stencilFail=t.stencilFail,this.stencilZFail=t.stencilZFail,this.stencilZPass=t.stencilZPass,this.stencilWrite=t.stencilWrite;const e=t.clippingPlanes;let n=null;if(e!==null){const i=e.length;n=new Array(i);for(let s=0;s!==i;++s)n[s]=e[s].clone()}return this.clippingPlanes=n,this.clipIntersection=t.clipIntersection,this.clipShadows=t.clipShadows,this.shadowSide=t.shadowSide,this.colorWrite=t.colorWrite,this.precision=t.precision,this.polygonOffset=t.polygonOffset,this.polygonOffsetFactor=t.polygonOffsetFactor,this.polygonOffsetUnits=t.polygonOffsetUnits,this.dithering=t.dithering,this.alphaTest=t.alphaTest,this.alphaHash=t.alphaHash,this.alphaToCoverage=t.alphaToCoverage,this.premultipliedAlpha=t.premultipliedAlpha,this.forceSinglePass=t.forceSinglePass,this.allowOverride=t.allowOverride,this.visible=t.visible,this.toneMapped=t.toneMapped,this.userData=JSON.parse(JSON.stringify(t.userData)),this}dispose(){this.dispatchEvent({type:"dispose"})}set needsUpdate(t){t===!0&&this.version++}}class Pt extends Ft{constructor(t){super(),this.isSpriteMaterial=!0,this.type="SpriteMaterial",this.color=new re(16777215),this.map=null,this.alphaMap=null,this.rotation=0,this.sizeAttenuation=!0,this.transparent=!0,this.fog=!0,this.setValues(t)}copy(t){return super.copy(t),this.color.copy(t.color),this.map=t.map,this.alphaMap=t.alphaMap,this.rotation=t.rotation,this.sizeAttenuation=t.sizeAttenuation,this.fog=t.fog,this}}let ne;const ce=new O,bt=new O,wt=new O,qt=new At,Zt=new At,Ht=new Me,Se=new O,k=new O,Rt=new O,Ct=new At,Xt=new At,Et=new At;class ft extends De{constructor(t=new Pt){if(super(),this.isSprite=!0,this.type="Sprite",ne===void 0){ne=new y;const e=new Float32Array([-.5,-.5,0,0,0,.5,-.5,0,1,0,.5,.5,0,1,1,-.5,.5,0,0,1]),n=new W(e,5);ne.setIndex([0,1,2,0,2,3]),ne.setAttribute("position",new yt(n,3,0,!1)),ne.setAttribute("uv",new yt(n,2,3,!1))}this.geometry=ne,this.material=t,this.center=new At(.5,.5),this.count=1}raycast(t,e){t.camera===null&&Re('Sprite: "Raycaster.camera" needs to be set in order to raycast against sprites.'),bt.setFromMatrixScale(this.matrixWorld),Ht.copy(t.camera.matrixWorld),this.modelViewMatrix.multiplyMatrices(t.camera.matrixWorldInverse,this.matrixWorld),wt.setFromMatrixPosition(this.modelViewMatrix),t.camera.isPerspectiveCamera&&this.material.sizeAttenuation===!1&&bt.multiplyScalar(-wt.z);const n=this.material.rotation;let i,s;n!==0&&(s=Math.cos(n),i=Math.sin(n));const r=this.center;$t(Se.set(-.5,-.5,0),wt,r,bt,i,s),$t(k.set(.5,-.5,0),wt,r,bt,i,s),$t(Rt.set(.5,.5,0),wt,r,bt,i,s),Ct.set(0,0),Xt.set(1,0),Et.set(1,1);let a=t.ray.intersectTriangle(Se,k,Rt,!1,ce);if(a===null&&($t(k.set(-.5,.5,0),wt,r,bt,i,s),Xt.set(0,1),a=t.ray.intersectTriangle(Se,Rt,k,!1,ce),a===null))return;const l=t.ray.origin.distanceTo(ce);l<t.near||l>t.far||e.push({distance:l,point:ce.clone(),uv:nt.getInterpolation(ce,Se,k,Rt,Ct,Xt,Et,new At),face:null,object:this})}copy(t,e){return super.copy(t,e),t.center!==void 0&&this.center.copy(t.center),this.material=t.material,this}}function $t(u,t,e,n,i,s){qt.subVectors(u,e).addScalar(.5).multiply(n),i!==void 0?(Zt.x=s*qt.x-i*qt.y,Zt.y=i*qt.x+s*qt.y):Zt.copy(qt),u.copy(t),u.x+=Zt.x,u.y+=Zt.y,u.applyMatrix4(Ht)}const me=new O,Ge=new O;class ze extends De{constructor(){super(),this.isLOD=!0,this._currentLevel=0,this.type="LOD",Object.defineProperties(this,{levels:{enumerable:!0,value:[]}}),this.autoUpdate=!0}copy(t){super.copy(t,!1);const e=t.levels;for(let n=0,i=e.length;n<i;n++){const s=e[n];this.addLevel(s.object.clone(),s.distance,s.hysteresis)}return this.autoUpdate=t.autoUpdate,this}addLevel(t,e=0,n=0){e=Math.abs(e);const i=this.levels;let s;for(s=0;s<i.length&&!(e<i[s].distance);s++);return i.splice(s,0,{distance:e,hysteresis:n,object:t}),this.add(t),this}removeLevel(t){const e=this.levels;for(let n=0;n<e.length;n++)if(e[n].distance===t){const i=e.splice(n,1);return this.remove(i[0].object),!0}return!1}getCurrentLevel(){return this._currentLevel}getObjectForDistance(t){const e=this.levels;if(e.length>0){let n,i;for(n=1,i=e.length;n<i;n++){let s=e[n].distance;if(e[n].object.visible&&(s-=s*e[n].hysteresis),t<s)break}return e[n-1].object}return null}raycast(t,e){if(this.levels.length>0){me.setFromMatrixPosition(this.matrixWorld);const i=t.ray.origin.distanceTo(me);this.getObjectForDistance(i).raycast(t,e)}}update(t){const e=this.levels;if(e.length>1){me.setFromMatrixPosition(t.matrixWorld),Ge.setFromMatrixPosition(this.matrixWorld);const n=me.distanceTo(Ge)/t.zoom;e[0].object.visible=!0;let i,s;for(i=1,s=e.length;i<s;i++){let r=e[i].distance;if(e[i].object.visible&&(r-=r*e[i].hysteresis),n>=r)e[i-1].object.visible=!1,e[i].object.visible=!0;else break}for(this._currentLevel=i-1;i<s;i++)e[i].object.visible=!1}}toJSON(t){const e=super.toJSON(t);this.autoUpdate===!1&&(e.object.autoUpdate=!1),e.object.levels=[];const n=this.levels;for(let i=0,s=n.length;i<s;i++){const r=n[i];e.object.levels.push({object:r.object.uuid,distance:r.distance,hysteresis:r.hysteresis})}return e}}const fn=new O,zn=new O,as=new O,Zn=new O,Lr=new O,os=new O,Bs=new O;class Vn{constructor(t=new O,e=new O(0,0,-1)){this.origin=t,this.direction=e}set(t,e){return this.origin.copy(t),this.direction.copy(e),this}copy(t){return this.origin.copy(t.origin),this.direction.copy(t.direction),this}at(t,e){return e.copy(this.origin).addScaledVector(this.direction,t)}lookAt(t){return this.direction.copy(t).sub(this.origin).normalize(),this}recast(t){return this.origin.copy(this.at(t,fn)),this}closestPointToPoint(t,e){e.subVectors(t,this.origin);const n=e.dot(this.direction);return n<0?e.copy(this.origin):e.copy(this.origin).addScaledVector(this.direction,n)}distanceToPoint(t){return Math.sqrt(this.distanceSqToPoint(t))}distanceSqToPoint(t){const e=fn.subVectors(t,this.origin).dot(this.direction);return e<0?this.origin.distanceToSquared(t):(fn.copy(this.origin).addScaledVector(this.direction,e),fn.distanceToSquared(t))}distanceSqToSegment(t,e,n,i){zn.copy(t).add(e).multiplyScalar(.5),as.copy(e).sub(t).normalize(),Zn.copy(this.origin).sub(zn);const s=t.distanceTo(e)*.5,r=-this.direction.dot(as),a=Zn.dot(this.direction),l=-Zn.dot(as),c=Zn.lengthSq(),d=Math.abs(1-r*r);let m,g,_,v;if(d>0)if(m=r*l-a,g=r*a-l,v=s*d,m>=0)if(g>=-v)if(g<=v){const M=1/d;m*=M,g*=M,_=m*(m+r*g+2*a)+g*(r*m+g+2*l)+c}else g=s,m=Math.max(0,-(r*g+a)),_=-m*m+g*(g+2*l)+c;else g=-s,m=Math.max(0,-(r*g+a)),_=-m*m+g*(g+2*l)+c;else g<=-v?(m=Math.max(0,-(-r*s+a)),g=m>0?-s:Math.min(Math.max(-s,-l),s),_=-m*m+g*(g+2*l)+c):g<=v?(m=0,g=Math.min(Math.max(-s,-l),s),_=g*(g+2*l)+c):(m=Math.max(0,-(r*s+a)),g=m>0?s:Math.min(Math.max(-s,-l),s),_=-m*m+g*(g+2*l)+c);else g=r>0?-s:s,m=Math.max(0,-(r*g+a)),_=-m*m+g*(g+2*l)+c;return n&&n.copy(this.origin).addScaledVector(this.direction,m),i&&i.copy(zn).addScaledVector(as,g),_}intersectSphere(t,e){fn.subVectors(t.center,this.origin);const n=fn.dot(this.direction),i=fn.dot(fn)-n*n,s=t.radius*t.radius;if(i>s)return null;const r=Math.sqrt(s-i),a=n-r,l=n+r;return l<0?null:a<0?this.at(l,e):this.at(a,e)}intersectsSphere(t){return t.radius<0?!1:this.distanceSqToPoint(t.center)<=t.radius*t.radius}distanceToPlane(t){const e=t.normal.dot(this.direction);if(e===0)return t.distanceToPoint(this.origin)===0?0:null;const n=-(this.origin.dot(t.normal)+t.constant)/e;return n>=0?n:null}intersectPlane(t,e){const n=this.distanceToPlane(t);return n===null?null:this.at(n,e)}intersectsPlane(t){const e=t.distanceToPoint(this.origin);return e===0||t.normal.dot(this.direction)*e<0}intersectBox(t,e){let n,i,s,r,a,l;const c=1/this.direction.x,d=1/this.direction.y,m=1/this.direction.z,g=this.origin;return c>=0?(n=(t.min.x-g.x)*c,i=(t.max.x-g.x)*c):(n=(t.max.x-g.x)*c,i=(t.min.x-g.x)*c),d>=0?(s=(t.min.y-g.y)*d,r=(t.max.y-g.y)*d):(s=(t.max.y-g.y)*d,r=(t.min.y-g.y)*d),n>r||s>i||((s>n||isNaN(n))&&(n=s),(r<i||isNaN(i))&&(i=r),m>=0?(a=(t.min.z-g.z)*m,l=(t.max.z-g.z)*m):(a=(t.max.z-g.z)*m,l=(t.min.z-g.z)*m),n>l||a>i)||((a>n||n!==n)&&(n=a),(l<i||i!==i)&&(i=l),i<0)?null:this.at(n>=0?n:i,e)}intersectsBox(t){return this.intersectBox(t,fn)!==null}intersectTriangle(t,e,n,i,s){Lr.subVectors(e,t),os.subVectors(n,t),Bs.crossVectors(Lr,os);let r=this.direction.dot(Bs),a;if(r>0){if(i)return null;a=1}else if(r<0)a=-1,r=-r;else return null;Zn.subVectors(this.origin,t);const l=a*this.direction.dot(os.crossVectors(Zn,os));if(l<0)return null;const c=a*this.direction.dot(Lr.cross(Zn));if(c<0||l+c>r)return null;const d=-a*Zn.dot(Bs);return d<0?null:this.at(d/r,s)}applyMatrix4(t){return this.origin.applyMatrix4(t),this.direction.transformDirection(t),this}equals(t){return t.origin.equals(this.origin)&&t.direction.equals(this.direction)}clone(){return new this.constructor().copy(this)}}class ei extends Ft{constructor(t){super(),this.isMeshBasicMaterial=!0,this.type="MeshBasicMaterial",this.color=new re(16777215),this.map=null,this.lightMap=null,this.lightMapIntensity=1,this.aoMap=null,this.aoMapIntensity=1,this.specularMap=null,this.alphaMap=null,this.envMap=null,this.envMapRotation=new Bn,this.combine=Es,this.reflectivity=1,this.refractionRatio=.98,this.wireframe=!1,this.wireframeLinewidth=1,this.wireframeLinecap="round",this.wireframeLinejoin="round",this.fog=!0,this.setValues(t)}copy(t){return super.copy(t),this.color.copy(t.color),this.map=t.map,this.lightMap=t.lightMap,this.lightMapIntensity=t.lightMapIntensity,this.aoMap=t.aoMap,this.aoMapIntensity=t.aoMapIntensity,this.specularMap=t.specularMap,this.alphaMap=t.alphaMap,this.envMap=t.envMap,this.envMapRotation.copy(t.envMapRotation),this.combine=t.combine,this.reflectivity=t.reflectivity,this.refractionRatio=t.refractionRatio,this.wireframe=t.wireframe,this.wireframeLinewidth=t.wireframeLinewidth,this.wireframeLinecap=t.wireframeLinecap,this.wireframeLinejoin=t.wireframeLinejoin,this.fog=t.fog,this}}const Dr=new Me,ci=new Vn,yi=new ae,Ur=new O,Mi=new O,ls=new O,cs=new O,Nr=new O,zs=new O,Ja=new O,Vs=new O;class kn extends De{constructor(t=new y,e=new ei){super(),this.isMesh=!0,this.type="Mesh",this.geometry=t,this.material=e,this.morphTargetDictionary=void 0,this.morphTargetInfluences=void 0,this.count=1,this.updateMorphTargets()}copy(t,e){return super.copy(t,e),t.morphTargetInfluences!==void 0&&(this.morphTargetInfluences=t.morphTargetInfluences.slice()),t.morphTargetDictionary!==void 0&&(this.morphTargetDictionary=Object.assign({},t.morphTargetDictionary)),this.material=Array.isArray(t.material)?t.material.slice():t.material,this.geometry=t.geometry,this}updateMorphTargets(){const e=this.geometry.morphAttributes,n=Object.keys(e);if(n.length>0){const i=e[n[0]];if(i!==void 0){this.morphTargetInfluences=[],this.morphTargetDictionary={};for(let s=0,r=i.length;s<r;s++){const a=i[s].name||String(s);this.morphTargetInfluences.push(0),this.morphTargetDictionary[a]=s}}}}getVertexPosition(t,e){const n=this.geometry,i=n.attributes.position,s=n.morphAttributes.position,r=n.morphTargetsRelative;e.fromBufferAttribute(i,t);const a=this.morphTargetInfluences;if(s&&a){zs.set(0,0,0);for(let l=0,c=s.length;l<c;l++){const d=a[l],m=s[l];d!==0&&(Nr.fromBufferAttribute(m,t),r?zs.addScaledVector(Nr,d):zs.addScaledVector(Nr.sub(e),d))}e.add(zs)}return e}raycast(t,e){const n=this.geometry,i=this.material,s=this.matrixWorld;i!==void 0&&(n.boundingSphere===null&&n.computeBoundingSphere(),yi.copy(n.boundingSphere),yi.applyMatrix4(s),ci.copy(t.ray).recast(t.near),!(yi.containsPoint(ci.origin)===!1&&(ci.intersectSphere(yi,Ur)===null||ci.origin.distanceToSquared(Ur)>(t.far-t.near)**2))&&(Dr.copy(s).invert(),ci.copy(t.ray).applyMatrix4(Dr),!(n.boundingBox!==null&&ci.intersectsBox(n.boundingBox)===!1)&&this._computeIntersections(t,e,ci)))}_computeIntersections(t,e,n){let i;const s=this.geometry,r=this.material,a=s.index,l=s.attributes.position,c=s.attributes.uv,d=s.attributes.uv1,m=s.attributes.normal,g=s.groups,_=s.drawRange;if(a!==null)if(Array.isArray(r))for(let v=0,M=g.length;v<M;v++){const C=g[v],w=r[C.materialIndex],L=Math.max(C.start,_.start),D=Math.min(a.count,Math.min(C.start+C.count,_.start+_.count));for(let N=L,et=D;N<et;N+=3){const rt=a.getX(N),vt=a.getX(N+1),at=a.getX(N+2);i=T(this,w,t,n,c,d,m,rt,vt,at),i&&(i.faceIndex=Math.floor(N/3),i.face.materialIndex=C.materialIndex,e.push(i))}}else{const v=Math.max(0,_.start),M=Math.min(a.count,_.start+_.count);for(let C=v,w=M;C<w;C+=3){const L=a.getX(C),D=a.getX(C+1),N=a.getX(C+2);i=T(this,r,t,n,c,d,m,L,D,N),i&&(i.faceIndex=Math.floor(C/3),e.push(i))}}else if(l!==void 0)if(Array.isArray(r))for(let v=0,M=g.length;v<M;v++){const C=g[v],w=r[C.materialIndex],L=Math.max(C.start,_.start),D=Math.min(l.count,Math.min(C.start+C.count,_.start+_.count));for(let N=L,et=D;N<et;N+=3){const rt=N,vt=N+1,at=N+2;i=T(this,w,t,n,c,d,m,rt,vt,at),i&&(i.faceIndex=Math.floor(N/3),i.face.materialIndex=C.materialIndex,e.push(i))}}else{const v=Math.max(0,_.start),M=Math.min(l.count,_.start+_.count);for(let C=v,w=M;C<w;C+=3){const L=C,D=C+1,N=C+2;i=T(this,r,t,n,c,d,m,L,D,N),i&&(i.faceIndex=Math.floor(C/3),e.push(i))}}}}function Uc(u,t,e,n,i,s,r,a){let l;if(t.side===Yr?l=n.intersectTriangle(r,s,i,!0,a):l=n.intersectTriangle(i,s,r,t.side===As,a),l===null)return null;Vs.copy(a),Vs.applyMatrix4(u.matrixWorld);const c=e.ray.origin.distanceTo(Vs);return c<e.near||c>e.far?null:{distance:c,point:Vs.clone(),object:u}}function T(u,t,e,n,i,s,r,a,l,c){u.getVertexPosition(a,Mi),u.getVertexPosition(l,ls),u.getVertexPosition(c,cs);const d=Uc(u,t,e,n,Mi,ls,cs,Ja);if(d){const m=new O;nt.getBarycoord(Ja,Mi,ls,cs,m),i&&(d.uv=nt.getInterpolatedAttribute(i,a,l,c,m,new At)),s&&(d.uv1=nt.getInterpolatedAttribute(s,a,l,c,m,new At)),r&&(d.normal=nt.getInterpolatedAttribute(r,a,l,c,m,new O),d.normal.dot(n.direction)>0&&d.normal.multiplyScalar(-1));const g={a,b:l,c,normal:new O,materialIndex:0};nt.getNormal(Mi,ls,cs,g.normal),d.face=g,d.barycoord=m}return d}const Y=new O,lt=new En,tt=new En,Q=new O,Vt=new Me,Wt=new O,Ot=new ae,jt=new Me,ie=new Vn;class be extends kn{constructor(t,e){super(t,e),this.isSkinnedMesh=!0,this.type="SkinnedMesh",this.bindMode=Zr,this.bindMatrix=new Me,this.bindMatrixInverse=new Me,this.boundingBox=null,this.boundingSphere=null}computeBoundingBox(){const t=this.geometry;this.boundingBox===null&&(this.boundingBox=new A),this.boundingBox.makeEmpty();const e=t.getAttribute("position");for(let n=0;n<e.count;n++)this.getVertexPosition(n,Wt),this.boundingBox.expandByPoint(Wt)}computeBoundingSphere(){const t=this.geometry;this.boundingSphere===null&&(this.boundingSphere=new ae),this.boundingSphere.makeEmpty();const e=t.getAttribute("position");for(let n=0;n<e.count;n++)this.getVertexPosition(n,Wt),this.boundingSphere.expandByPoint(Wt)}copy(t,e){return super.copy(t,e),this.bindMode=t.bindMode,this.bindMatrix.copy(t.bindMatrix),this.bindMatrixInverse.copy(t.bindMatrixInverse),this.skeleton=t.skeleton,t.boundingBox!==null&&(this.boundingBox=t.boundingBox.clone()),t.boundingSphere!==null&&(this.boundingSphere=t.boundingSphere.clone()),this}raycast(t,e){const n=this.material,i=this.matrixWorld;n!==void 0&&(this.boundingSphere===null&&this.computeBoundingSphere(),Ot.copy(this.boundingSphere),Ot.applyMatrix4(i),t.ray.intersectsSphere(Ot)!==!1&&(jt.copy(i).invert(),ie.copy(t.ray).applyMatrix4(jt),!(this.boundingBox!==null&&ie.intersectsBox(this.boundingBox)===!1)&&this._computeIntersections(t,e,ie)))}getVertexPosition(t,e){return super.getVertexPosition(t,e),this.applyBoneTransform(t,e),e}bind(t,e){this.skeleton=t,e===void 0&&(this.updateMatrixWorld(!0),this.skeleton.calculateInverses(),e=this.matrixWorld),this.bindMatrix.copy(e),this.bindMatrixInverse.copy(e).invert()}pose(){this.skeleton.pose()}normalizeSkinWeights(){const t=new En,e=this.geometry.attributes.skinWeight;for(let n=0,i=e.count;n<i;n++){t.fromBufferAttribute(e,n);const s=1/t.manhattanLength();s!==1/0?t.multiplyScalar(s):t.set(1,0,0,0),e.setXYZW(n,t.x,t.y,t.z,t.w)}}updateMatrixWorld(t){super.updateMatrixWorld(t),this.bindMode===Zr?this.bindMatrixInverse.copy(this.matrixWorld).invert():this.bindMode===bl?this.bindMatrixInverse.copy(this.bindMatrix).invert():le("SkinnedMesh: Unrecognized bindMode: "+this.bindMode)}applyBoneTransform(t,e){const n=this.skeleton,i=this.geometry;lt.fromBufferAttribute(i.attributes.skinIndex,t),tt.fromBufferAttribute(i.attributes.skinWeight,t),Y.copy(e).applyMatrix4(this.bindMatrix),e.set(0,0,0);for(let s=0;s<4;s++){const r=tt.getComponent(s);if(r!==0){const a=lt.getComponent(s);Vt.multiplyMatrices(n.bones[a].matrixWorld,n.boneInverses[a]),e.addScaledVector(Q.copy(Y).applyMatrix4(Vt),r)}}return e.applyMatrix4(this.bindMatrixInverse)}}class Ae extends De{constructor(){super(),this.isBone=!0,this.type="Bone"}}class Yt extends Ye{constructor(t=null,e=1,n=1,i,s,r,a,l,c=Mn,d=Mn,m,g){super(null,r,a,l,c,d,i,s,m,g),this.isDataTexture=!0,this.image={data:t,width:e,height:n},this.generateMipmaps=!1,this.flipY=!1,this.unpackAlignment=1}}const Ve=new Me,Ke=new Me;class He{constructor(t=[],e=[]){this.uuid=An(),this.bones=t.slice(0),this.boneInverses=e,this.boneMatrices=null,this.previousBoneMatrices=null,this.boneTexture=null,this.init()}init(){const t=this.bones,e=this.boneInverses;if(this.boneMatrices=new Float32Array(t.length*16),e.length===0)this.calculateInverses();else if(t.length!==e.length){le("Skeleton: Number of inverse bone matrices does not match amount of bones."),this.boneInverses=[];for(let n=0,i=this.bones.length;n<i;n++)this.boneInverses.push(new Me)}}calculateInverses(){this.boneInverses.length=0;for(let t=0,e=this.bones.length;t<e;t++){const n=new Me;this.bones[t]&&n.copy(this.bones[t].matrixWorld).invert(),this.boneInverses.push(n)}}pose(){for(let t=0,e=this.bones.length;t<e;t++){const n=this.bones[t];n&&n.matrixWorld.copy(this.boneInverses[t]).invert()}for(let t=0,e=this.bones.length;t<e;t++){const n=this.bones[t];n&&(n.parent&&n.parent.isBone?(n.matrix.copy(n.parent.matrixWorld).invert(),n.matrix.multiply(n.matrixWorld)):n.matrix.copy(n.matrixWorld),n.matrix.decompose(n.position,n.quaternion,n.scale))}}update(){const t=this.bones,e=this.boneInverses,n=this.boneMatrices,i=this.boneTexture;for(let s=0,r=t.length;s<r;s++){const a=t[s]?t[s].matrixWorld:Ke;Ve.multiplyMatrices(a,e[s]),Ve.toArray(n,s*16)}i!==null&&(i.needsUpdate=!0)}clone(){return new He(this.bones,this.boneInverses)}computeBoneTexture(){let t=Math.sqrt(this.bones.length*4);t=Math.ceil(t/4)*4,t=Math.max(t,4);const e=new Float32Array(t*t*4);e.set(this.boneMatrices);const n=new Yt(e,t,t,Ci,gi);return n.needsUpdate=!0,this.boneMatrices=e,this.boneTexture=n,this}getBoneByName(t){for(let e=0,n=this.bones.length;e<n;e++){const i=this.bones[e];if(i.name===t)return i}}dispose(){this.boneTexture!==null&&(this.boneTexture.dispose(),this.boneTexture=null)}fromJSON(t,e){this.uuid=t.uuid;for(let n=0,i=t.bones.length;n<i;n++){const s=t.bones[n];let r=e[s];r===void 0&&(le("Skeleton: No bone found with UUID:",s),r=new Ae),this.bones.push(r),this.boneInverses.push(new Me().fromArray(t.boneInverses[n]))}return this.init(),this}toJSON(){const t={metadata:{version:4.7,type:"Skeleton",generator:"Skeleton.toJSON"},bones:[],boneInverses:[]};t.uuid=this.uuid;const e=this.bones,n=this.boneInverses;for(let i=0,s=e.length;i<s;i++){const r=e[i];t.bones.push(r.uuid);const a=n[i];t.boneInverses.push(a.toArray())}return t}}class Ue extends se{constructor(t,e,n,i=1){super(t,e,n),this.isInstancedBufferAttribute=!0,this.meshPerAttribute=i}copy(t){return super.copy(t),this.meshPerAttribute=t.meshPerAttribute,this}toJSON(){const t=super.toJSON();return t.meshPerAttribute=this.meshPerAttribute,t.isInstancedBufferAttribute=!0,t}}const Qe=new Me,ee=new Me,vn=[],Le=new A,Gn=new Me,Sn=new kn,Hn=new ae;class Vi extends kn{constructor(t,e,n){super(t,e),this.isInstancedMesh=!0,this.instanceMatrix=new Ue(new Float32Array(n*16),16),this.previousInstanceMatrix=null,this.instanceColor=null,this.morphTexture=null,this.count=n,this.boundingBox=null,this.boundingSphere=null;for(let i=0;i<n;i++)this.setMatrixAt(i,Gn)}computeBoundingBox(){const t=this.geometry,e=this.count;this.boundingBox===null&&(this.boundingBox=new A),t.boundingBox===null&&t.computeBoundingBox(),this.boundingBox.makeEmpty();for(let n=0;n<e;n++)this.getMatrixAt(n,Qe),Le.copy(t.boundingBox).applyMatrix4(Qe),this.boundingBox.union(Le)}computeBoundingSphere(){const t=this.geometry,e=this.count;this.boundingSphere===null&&(this.boundingSphere=new ae),t.boundingSphere===null&&t.computeBoundingSphere(),this.boundingSphere.makeEmpty();for(let n=0;n<e;n++)this.getMatrixAt(n,Qe),Hn.copy(t.boundingSphere).applyMatrix4(Qe),this.boundingSphere.union(Hn)}copy(t,e){return super.copy(t,e),this.instanceMatrix.copy(t.instanceMatrix),t.previousInstanceMatrix!==null&&(this.previousInstanceMatrix=t.previousInstanceMatrix.clone()),t.morphTexture!==null&&(this.morphTexture=t.morphTexture.clone()),t.instanceColor!==null&&(this.instanceColor=t.instanceColor.clone()),this.count=t.count,t.boundingBox!==null&&(this.boundingBox=t.boundingBox.clone()),t.boundingSphere!==null&&(this.boundingSphere=t.boundingSphere.clone()),this}getColorAt(t,e){e.fromArray(this.instanceColor.array,t*3)}getMatrixAt(t,e){e.fromArray(this.instanceMatrix.array,t*16)}getMorphAt(t,e){const n=e.morphTargetInfluences,i=this.morphTexture.source.data.data,s=n.length+1,r=t*s+1;for(let a=0;a<n.length;a++)n[a]=i[r+a]}raycast(t,e){const n=this.matrixWorld,i=this.count;if(Sn.geometry=this.geometry,Sn.material=this.material,Sn.material!==void 0&&(this.boundingSphere===null&&this.computeBoundingSphere(),Hn.copy(this.boundingSphere),Hn.applyMatrix4(n),t.ray.intersectsSphere(Hn)!==!1))for(let s=0;s<i;s++){this.getMatrixAt(s,Qe),ee.multiplyMatrices(n,Qe),Sn.matrixWorld=ee,Sn.raycast(t,vn);for(let r=0,a=vn.length;r<a;r++){const l=vn[r];l.instanceId=s,l.object=this,e.push(l)}vn.length=0}}setColorAt(t,e){this.instanceColor===null&&(this.instanceColor=new Ue(new Float32Array(this.instanceMatrix.count*3).fill(1),3)),e.toArray(this.instanceColor.array,t*3)}setMatrixAt(t,e){e.toArray(this.instanceMatrix.array,t*16)}setMorphAt(t,e){const n=e.morphTargetInfluences,i=n.length+1;this.morphTexture===null&&(this.morphTexture=new Yt(new Float32Array(i*this.count),i,this.count,cr,gi));const s=this.morphTexture.source.data.data;let r=0;for(let c=0;c<n.length;c++)r+=n[c];const a=this.geometry.morphTargetsRelative?1:1-r,l=i*t;s[l]=a,s.set(n,l+1)}updateMorphTargets(){}dispose(){this.dispatchEvent({type:"dispose"}),this.morphTexture!==null&&(this.morphTexture.dispose(),this.morphTexture=null)}}const ke=new O,dn=new O,hi=new Fn;class We{constructor(t=new O(1,0,0),e=0){this.isPlane=!0,this.normal=t,this.constant=e}set(t,e){return this.normal.copy(t),this.constant=e,this}setComponents(t,e,n,i){return this.normal.set(t,e,n),this.constant=i,this}setFromNormalAndCoplanarPoint(t,e){return this.normal.copy(t),this.constant=-e.dot(this.normal),this}setFromCoplanarPoints(t,e,n){const i=ke.subVectors(n,e).cross(dn.subVectors(t,e)).normalize();return this.setFromNormalAndCoplanarPoint(i,t),this}copy(t){return this.normal.copy(t.normal),this.constant=t.constant,this}normalize(){const t=1/this.normal.length();return this.normal.multiplyScalar(t),this.constant*=t,this}negate(){return this.constant*=-1,this.normal.negate(),this}distanceToPoint(t){return this.normal.dot(t)+this.constant}distanceToSphere(t){return this.distanceToPoint(t.center)-t.radius}projectPoint(t,e){return e.copy(t).addScaledVector(this.normal,-this.distanceToPoint(t))}intersectLine(t,e){const n=t.delta(ke),i=this.normal.dot(n);if(i===0)return this.distanceToPoint(t.start)===0?e.copy(t.start):null;const s=-(t.start.dot(this.normal)+this.constant)/i;return s<0||s>1?null:e.copy(t.start).addScaledVector(n,s)}intersectsLine(t){const e=this.distanceToPoint(t.start),n=this.distanceToPoint(t.end);return e<0&&n>0||n<0&&e>0}intersectsBox(t){return t.intersectsPlane(this)}intersectsSphere(t){return t.intersectsPlane(this)}coplanarPoint(t){return t.copy(this.normal).multiplyScalar(-this.constant)}applyMatrix4(t,e){const n=e||hi.getNormalMatrix(t),i=this.coplanarPoint(ke).applyMatrix4(t),s=this.normal.applyMatrix3(n).normalize();return this.constant=-i.dot(s),this}translate(t){return this.constant-=t.dot(this.normal),this}equals(t){return t.normal.equals(this.normal)&&t.constant===this.constant}clone(){return new this.constructor().copy(this)}}const bn=new ae,ki=new At(.5,.5),hs=new O;class Ka{constructor(t=new We,e=new We,n=new We,i=new We,s=new We,r=new We){this.planes=[t,e,n,i,s,r]}set(t,e,n,i,s,r){const a=this.planes;return a[0].copy(t),a[1].copy(e),a[2].copy(n),a[3].copy(i),a[4].copy(s),a[5].copy(r),this}copy(t){const e=this.planes;for(let n=0;n<6;n++)e[n].copy(t.planes[n]);return this}setFromProjectionMatrix(t,e=Nn,n=!1){const i=this.planes,s=t.elements,r=s[0],a=s[1],l=s[2],c=s[3],d=s[4],m=s[5],g=s[6],_=s[7],v=s[8],M=s[9],C=s[10],w=s[11],L=s[12],D=s[13],N=s[14],et=s[15];if(i[0].setComponents(c-r,_-d,w-v,et-L).normalize(),i[1].setComponents(c+r,_+d,w+v,et+L).normalize(),i[2].setComponents(c+a,_+m,w+M,et+D).normalize(),i[3].setComponents(c-a,_-m,w-M,et-D).normalize(),n)i[4].setComponents(l,g,C,N).normalize(),i[5].setComponents(c-l,_-g,w-C,et-N).normalize();else if(i[4].setComponents(c-l,_-g,w-C,et-N).normalize(),e===Nn)i[5].setComponents(c+l,_+g,w+C,et+N).normalize();else if(e===Pi)i[5].setComponents(l,g,C,N).normalize();else throw new Error("THREE.Frustum.setFromProjectionMatrix(): Invalid coordinate system: "+e);return this}intersectsObject(t){if(t.boundingSphere!==void 0)t.boundingSphere===null&&t.computeBoundingSphere(),bn.copy(t.boundingSphere).applyMatrix4(t.matrixWorld);else{const e=t.geometry;e.boundingSphere===null&&e.computeBoundingSphere(),bn.copy(e.boundingSphere).applyMatrix4(t.matrixWorld)}return this.intersectsSphere(bn)}intersectsSprite(t){bn.center.set(0,0,0);const e=ki.distanceTo(t.center);return bn.radius=.7071067811865476+e,bn.applyMatrix4(t.matrixWorld),this.intersectsSphere(bn)}intersectsSphere(t){const e=this.planes,n=t.center,i=-t.radius;for(let s=0;s<6;s++)if(e[s].distanceToPoint(n)<i)return!1;return!0}intersectsBox(t){const e=this.planes;for(let n=0;n<6;n++){const i=e[n];if(hs.x=i.normal.x>0?t.max.x:t.min.x,hs.y=i.normal.y>0?t.max.y:t.min.y,hs.z=i.normal.z>0?t.max.z:t.min.z,i.distanceToPoint(hs)<0)return!1}return!0}containsPoint(t){const e=this.planes;for(let n=0;n<6;n++)if(e[n].distanceToPoint(t)<0)return!1;return!0}clone(){return new this.constructor().copy(this)}}const ui=new Me,fi=new Ka;class Nc{constructor(){this.coordinateSystem=Nn}intersectsObject(t,e){if(!e.isArrayCamera||e.cameras.length===0)return!1;for(let n=0;n<e.cameras.length;n++){const i=e.cameras[n];if(ui.multiplyMatrices(i.projectionMatrix,i.matrixWorldInverse),fi.setFromProjectionMatrix(ui,i.coordinateSystem,i.reversedDepth),fi.intersectsObject(t))return!0}return!1}intersectsSprite(t,e){if(!e||!e.cameras||e.cameras.length===0)return!1;for(let n=0;n<e.cameras.length;n++){const i=e.cameras[n];if(ui.multiplyMatrices(i.projectionMatrix,i.matrixWorldInverse),fi.setFromProjectionMatrix(ui,i.coordinateSystem,i.reversedDepth),fi.intersectsSprite(t))return!0}return!1}intersectsSphere(t,e){if(!e||!e.cameras||e.cameras.length===0)return!1;for(let n=0;n<e.cameras.length;n++){const i=e.cameras[n];if(ui.multiplyMatrices(i.projectionMatrix,i.matrixWorldInverse),fi.setFromProjectionMatrix(ui,i.coordinateSystem,i.reversedDepth),fi.intersectsSphere(t))return!0}return!1}intersectsBox(t,e){if(!e||!e.cameras||e.cameras.length===0)return!1;for(let n=0;n<e.cameras.length;n++){const i=e.cameras[n];if(ui.multiplyMatrices(i.projectionMatrix,i.matrixWorldInverse),fi.setFromProjectionMatrix(ui,i.coordinateSystem,i.reversedDepth),fi.intersectsBox(t))return!0}return!1}containsPoint(t,e){if(!e||!e.cameras||e.cameras.length===0)return!1;for(let n=0;n<e.cameras.length;n++){const i=e.cameras[n];if(ui.multiplyMatrices(i.projectionMatrix,i.matrixWorldInverse),fi.setFromProjectionMatrix(ui,i.coordinateSystem,i.reversedDepth),fi.containsPoint(t))return!0}return!1}clone(){return new Nc}}function Fc(u,t){return u-t}function Af(u,t){return u.z-t.z}function Ef(u,t){return t.z-u.z}class wf{constructor(){this.index=0,this.pool=[],this.list=[]}push(t,e,n,i){const s=this.pool,r=this.list;this.index>=s.length&&s.push({start:-1,count:-1,z:-1,index:-1});const a=s[this.index];r.push(a),this.index++,a.start=t,a.count=e,a.z=n,a.index=i}reset(){this.list.length=0,this.index=0}}const In=new Me,Cf=new re(1,1,1),cu=new Ka,Rf=new Nc,Qa=new A,us=new ae,Fr=new O,hu=new O,If=new O,Oc=new wf,Tn=new kn,ja=[];function Pf(u,t,e=0){const n=t.itemSize;if(u.isInterleavedBufferAttribute||u.array.constructor!==t.array.constructor){const i=u.count;for(let s=0;s<i;s++)for(let r=0;r<n;r++)t.setComponent(s+e,r,u.getComponent(s,r))}else t.array.set(u.array,e*n);t.needsUpdate=!0}function fs(u,t){if(u.constructor!==t.constructor){const e=Math.min(u.length,t.length);for(let n=0;n<e;n++)t[n]=u[n]}else{const e=Math.min(u.length,t.length);t.set(new u.constructor(u.buffer,0,e))}}class Lf extends kn{constructor(t,e,n=e*2,i){super(new y,i),this.isBatchedMesh=!0,this.perObjectFrustumCulled=!0,this.sortObjects=!0,this.boundingBox=null,this.boundingSphere=null,this.customSort=null,this._instanceInfo=[],this._geometryInfo=[],this._availableInstanceIds=[],this._availableGeometryIds=[],this._nextIndexStart=0,this._nextVertexStart=0,this._geometryCount=0,this._visibilityChanged=!0,this._geometryInitialized=!1,this._maxInstanceCount=t,this._maxVertexCount=e,this._maxIndexCount=n,this._multiDrawCounts=new Int32Array(t),this._multiDrawStarts=new Int32Array(t),this._multiDrawCount=0,this._multiDrawInstances=null,this._matricesTexture=null,this._indirectTexture=null,this._colorsTexture=null,this._initMatricesTexture(),this._initIndirectTexture()}get maxInstanceCount(){return this._maxInstanceCount}get instanceCount(){return this._instanceInfo.length-this._availableInstanceIds.length}get unusedVertexCount(){return this._maxVertexCount-this._nextVertexStart}get unusedIndexCount(){return this._maxIndexCount-this._nextIndexStart}_initMatricesTexture(){let t=Math.sqrt(this._maxInstanceCount*4);t=Math.ceil(t/4)*4,t=Math.max(t,4);const e=new Float32Array(t*t*4),n=new Yt(e,t,t,Ci,gi);this._matricesTexture=n}_initIndirectTexture(){let t=Math.sqrt(this._maxInstanceCount);t=Math.ceil(t);const e=new Uint32Array(t*t),n=new Yt(e,t,t,hr,Zi);this._indirectTexture=n}_initColorsTexture(){let t=Math.sqrt(this._maxInstanceCount);t=Math.ceil(t);const e=new Float32Array(t*t*4).fill(1),n=new Yt(e,t,t,Ci,gi);n.colorSpace=_n.workingColorSpace,this._colorsTexture=n}_initializeGeometry(t){const e=this.geometry,n=this._maxVertexCount,i=this._maxIndexCount;if(this._geometryInitialized===!1){for(const s in t.attributes){const r=t.getAttribute(s),{array:a,itemSize:l,normalized:c}=r,d=new a.constructor(n*l),m=new se(d,l,c);e.setAttribute(s,m)}if(t.getIndex()!==null){const s=n>65535?new Uint32Array(i):new Uint16Array(i);e.setIndex(new se(s,1))}this._geometryInitialized=!0}}_validateGeometry(t){const e=this.geometry;if(!!t.getIndex()!=!!e.getIndex())throw new Error('THREE.BatchedMesh: All geometries must consistently have "index".');for(const n in e.attributes){if(!t.hasAttribute(n))throw new Error(`THREE.BatchedMesh: Added geometry missing "${n}". All geometries must have consistent attributes.`);const i=t.getAttribute(n),s=e.getAttribute(n);if(i.itemSize!==s.itemSize||i.normalized!==s.normalized)throw new Error("THREE.BatchedMesh: All attributes must have a consistent itemSize and normalized value.")}}validateInstanceId(t){const e=this._instanceInfo;if(t<0||t>=e.length||e[t].active===!1)throw new Error(`THREE.BatchedMesh: Invalid instanceId ${t}. Instance is either out of range or has been deleted.`)}validateGeometryId(t){const e=this._geometryInfo;if(t<0||t>=e.length||e[t].active===!1)throw new Error(`THREE.BatchedMesh: Invalid geometryId ${t}. Geometry is either out of range or has been deleted.`)}setCustomSort(t){return this.customSort=t,this}computeBoundingBox(){this.boundingBox===null&&(this.boundingBox=new A);const t=this.boundingBox,e=this._instanceInfo;t.makeEmpty();for(let n=0,i=e.length;n<i;n++){if(e[n].active===!1)continue;const s=e[n].geometryIndex;this.getMatrixAt(n,In),this.getBoundingBoxAt(s,Qa).applyMatrix4(In),t.union(Qa)}}computeBoundingSphere(){this.boundingSphere===null&&(this.boundingSphere=new ae);const t=this.boundingSphere,e=this._instanceInfo;t.makeEmpty();for(let n=0,i=e.length;n<i;n++){if(e[n].active===!1)continue;const s=e[n].geometryIndex;this.getMatrixAt(n,In),this.getBoundingSphereAt(s,us).applyMatrix4(In),t.union(us)}}addInstance(t){if(this._instanceInfo.length>=this.maxInstanceCount&&this._availableInstanceIds.length===0)throw new Error("THREE.BatchedMesh: Maximum item count reached.");const n={visible:!0,active:!0,geometryIndex:t};let i=null;this._availableInstanceIds.length>0?(this._availableInstanceIds.sort(Fc),i=this._availableInstanceIds.shift(),this._instanceInfo[i]=n):(i=this._instanceInfo.length,this._instanceInfo.push(n));const s=this._matricesTexture;In.identity().toArray(s.image.data,i*16),s.needsUpdate=!0;const r=this._colorsTexture;return r&&(Cf.toArray(r.image.data,i*4),r.needsUpdate=!0),this._visibilityChanged=!0,i}addGeometry(t,e=-1,n=-1){this._initializeGeometry(t),this._validateGeometry(t);const i={vertexStart:-1,vertexCount:-1,reservedVertexCount:-1,indexStart:-1,indexCount:-1,reservedIndexCount:-1,start:-1,count:-1,boundingBox:null,boundingSphere:null,active:!0},s=this._geometryInfo;i.vertexStart=this._nextVertexStart,i.reservedVertexCount=e===-1?t.getAttribute("position").count:e;const r=t.getIndex();if(r!==null&&(i.indexStart=this._nextIndexStart,i.reservedIndexCount=n===-1?r.count:n),i.indexStart!==-1&&i.indexStart+i.reservedIndexCount>this._maxIndexCount||i.vertexStart+i.reservedVertexCount>this._maxVertexCount)throw new Error("THREE.BatchedMesh: Reserved space request exceeds the maximum buffer size.");let l;return this._availableGeometryIds.length>0?(this._availableGeometryIds.sort(Fc),l=this._availableGeometryIds.shift(),s[l]=i):(l=this._geometryCount,this._geometryCount++,s.push(i)),this.setGeometryAt(l,t),this._nextIndexStart=i.indexStart+i.reservedIndexCount,this._nextVertexStart=i.vertexStart+i.reservedVertexCount,l}setGeometryAt(t,e){if(t>=this._geometryCount)throw new Error("THREE.BatchedMesh: Maximum geometry count reached.");this._validateGeometry(e);const n=this.geometry,i=n.getIndex()!==null,s=n.getIndex(),r=e.getIndex(),a=this._geometryInfo[t];if(i&&r.count>a.reservedIndexCount||e.attributes.position.count>a.reservedVertexCount)throw new Error("THREE.BatchedMesh: Reserved space not large enough for provided geometry.");const l=a.vertexStart,c=a.reservedVertexCount;a.vertexCount=e.getAttribute("position").count;for(const d in n.attributes){const m=e.getAttribute(d),g=n.getAttribute(d);Pf(m,g,l);const _=m.itemSize;for(let v=m.count,M=c;v<M;v++){const C=l+v;for(let w=0;w<_;w++)g.setComponent(C,w,0)}g.needsUpdate=!0,g.addUpdateRange(l*_,c*_)}if(i){const d=a.indexStart,m=a.reservedIndexCount;a.indexCount=e.getIndex().count;for(let g=0;g<r.count;g++)s.setX(d+g,l+r.getX(g));for(let g=r.count,_=m;g<_;g++)s.setX(d+g,l);s.needsUpdate=!0,s.addUpdateRange(d,a.reservedIndexCount)}return a.start=i?a.indexStart:a.vertexStart,a.count=i?a.indexCount:a.vertexCount,a.boundingBox=null,e.boundingBox!==null&&(a.boundingBox=e.boundingBox.clone()),a.boundingSphere=null,e.boundingSphere!==null&&(a.boundingSphere=e.boundingSphere.clone()),this._visibilityChanged=!0,t}deleteGeometry(t){const e=this._geometryInfo;if(t>=e.length||e[t].active===!1)return this;const n=this._instanceInfo;for(let i=0,s=n.length;i<s;i++)n[i].active&&n[i].geometryIndex===t&&this.deleteInstance(i);return e[t].active=!1,this._availableGeometryIds.push(t),this._visibilityChanged=!0,this}deleteInstance(t){return this.validateInstanceId(t),this._instanceInfo[t].active=!1,this._availableInstanceIds.push(t),this._visibilityChanged=!0,this}optimize(){let t=0,e=0;const n=this._geometryInfo,i=n.map((r,a)=>a).sort((r,a)=>n[r].vertexStart-n[a].vertexStart),s=this.geometry;for(let r=0,a=n.length;r<a;r++){const l=i[r],c=n[l];if(c.active!==!1){if(s.index!==null){if(c.indexStart!==e){const{indexStart:d,vertexStart:m,reservedIndexCount:g}=c,_=s.index,v=_.array,M=t-m;for(let C=d;C<d+g;C++)v[C]=v[C]+M;_.array.copyWithin(e,d,d+g),_.addUpdateRange(e,g),_.needsUpdate=!0,c.indexStart=e}e+=c.reservedIndexCount}if(c.vertexStart!==t){const{vertexStart:d,reservedVertexCount:m}=c,g=s.attributes;for(const _ in g){const v=g[_],{array:M,itemSize:C}=v;M.copyWithin(t*C,d*C,(d+m)*C),v.addUpdateRange(t*C,m*C),v.needsUpdate=!0}c.vertexStart=t}t+=c.reservedVertexCount,c.start=s.index?c.indexStart:c.vertexStart}}return this._nextIndexStart=e,this._nextVertexStart=t,this._visibilityChanged=!0,this}getBoundingBoxAt(t,e){if(t>=this._geometryCount)return null;const n=this.geometry,i=this._geometryInfo[t];if(i.boundingBox===null){const s=new A,r=n.index,a=n.attributes.position;for(let l=i.start,c=i.start+i.count;l<c;l++){let d=l;r&&(d=r.getX(d)),s.expandByPoint(Fr.fromBufferAttribute(a,d))}i.boundingBox=s}return e.copy(i.boundingBox),e}getBoundingSphereAt(t,e){if(t>=this._geometryCount)return null;const n=this.geometry,i=this._geometryInfo[t];if(i.boundingSphere===null){const s=new ae;this.getBoundingBoxAt(t,Qa),Qa.getCenter(s.center);const r=n.index,a=n.attributes.position;let l=0;for(let c=i.start,d=i.start+i.count;c<d;c++){let m=c;r&&(m=r.getX(m)),Fr.fromBufferAttribute(a,m),l=Math.max(l,s.center.distanceToSquared(Fr))}s.radius=Math.sqrt(l),i.boundingSphere=s}return e.copy(i.boundingSphere),e}setMatrixAt(t,e){this.validateInstanceId(t);const n=this._matricesTexture,i=this._matricesTexture.image.data;return e.toArray(i,t*16),n.needsUpdate=!0,this}getMatrixAt(t,e){return this.validateInstanceId(t),e.fromArray(this._matricesTexture.image.data,t*16)}setColorAt(t,e){return this.validateInstanceId(t),this._colorsTexture===null&&this._initColorsTexture(),e.toArray(this._colorsTexture.image.data,t*4),this._colorsTexture.needsUpdate=!0,this}getColorAt(t,e){return this.validateInstanceId(t),e.fromArray(this._colorsTexture.image.data,t*4)}setVisibleAt(t,e){return this.validateInstanceId(t),this._instanceInfo[t].visible===e?this:(this._instanceInfo[t].visible=e,this._visibilityChanged=!0,this)}getVisibleAt(t){return this.validateInstanceId(t),this._instanceInfo[t].visible}setGeometryIdAt(t,e){return this.validateInstanceId(t),this.validateGeometryId(e),this._instanceInfo[t].geometryIndex=e,this}getGeometryIdAt(t){return this.validateInstanceId(t),this._instanceInfo[t].geometryIndex}getGeometryRangeAt(t,e={}){this.validateGeometryId(t);const n=this._geometryInfo[t];return e.vertexStart=n.vertexStart,e.vertexCount=n.vertexCount,e.reservedVertexCount=n.reservedVertexCount,e.indexStart=n.indexStart,e.indexCount=n.indexCount,e.reservedIndexCount=n.reservedIndexCount,e.start=n.start,e.count=n.count,e}setInstanceCount(t){const e=this._availableInstanceIds,n=this._instanceInfo;for(e.sort(Fc);e[e.length-1]===n.length-1;)n.pop(),e.pop();if(t<n.length)throw new Error(`BatchedMesh: Instance ids outside the range ${t} are being used. Cannot shrink instance count.`);const i=new Int32Array(t),s=new Int32Array(t);fs(this._multiDrawCounts,i),fs(this._multiDrawStarts,s),this._multiDrawCounts=i,this._multiDrawStarts=s,this._maxInstanceCount=t;const r=this._indirectTexture,a=this._matricesTexture,l=this._colorsTexture;r.dispose(),this._initIndirectTexture(),fs(r.image.data,this._indirectTexture.image.data),a.dispose(),this._initMatricesTexture(),fs(a.image.data,this._matricesTexture.image.data),l&&(l.dispose(),this._initColorsTexture(),fs(l.image.data,this._colorsTexture.image.data))}setGeometrySize(t,e){const n=[...this._geometryInfo].filter(a=>a.active);if(Math.max(...n.map(a=>a.vertexStart+a.reservedVertexCount))>t)throw new Error(`BatchedMesh: Geometry vertex values are being used outside the range ${e}. Cannot shrink further.`);if(this.geometry.index&&Math.max(...n.map(l=>l.indexStart+l.reservedIndexCount))>e)throw new Error(`BatchedMesh: Geometry index values are being used outside the range ${e}. Cannot shrink further.`);const s=this.geometry;s.dispose(),this._maxVertexCount=t,this._maxIndexCount=e,this._geometryInitialized&&(this._geometryInitialized=!1,this.geometry=new y,this._initializeGeometry(s));const r=this.geometry;s.index&&fs(s.index.array,r.index.array);for(const a in s.attributes)fs(s.attributes[a].array,r.attributes[a].array)}raycast(t,e){const n=this._instanceInfo,i=this._geometryInfo,s=this.matrixWorld,r=this.geometry;Tn.material=this.material,Tn.geometry.index=r.index,Tn.geometry.attributes=r.attributes,Tn.geometry.boundingBox===null&&(Tn.geometry.boundingBox=new A),Tn.geometry.boundingSphere===null&&(Tn.geometry.boundingSphere=new ae);for(let a=0,l=n.length;a<l;a++){if(!n[a].visible||!n[a].active)continue;const c=n[a].geometryIndex,d=i[c];Tn.geometry.setDrawRange(d.start,d.count),this.getMatrixAt(a,Tn.matrixWorld).premultiply(s),this.getBoundingBoxAt(c,Tn.geometry.boundingBox),this.getBoundingSphereAt(c,Tn.geometry.boundingSphere),Tn.raycast(t,ja);for(let m=0,g=ja.length;m<g;m++){const _=ja[m];_.object=this,_.batchId=a,e.push(_)}ja.length=0}Tn.material=null,Tn.geometry.index=null,Tn.geometry.attributes={},Tn.geometry.setDrawRange(0,1/0)}copy(t){return super.copy(t),this.geometry=t.geometry.clone(),this.perObjectFrustumCulled=t.perObjectFrustumCulled,this.sortObjects=t.sortObjects,this.boundingBox=t.boundingBox!==null?t.boundingBox.clone():null,this.boundingSphere=t.boundingSphere!==null?t.boundingSphere.clone():null,this._geometryInfo=t._geometryInfo.map(e=>({...e,boundingBox:e.boundingBox!==null?e.boundingBox.clone():null,boundingSphere:e.boundingSphere!==null?e.boundingSphere.clone():null})),this._instanceInfo=t._instanceInfo.map(e=>({...e})),this._availableInstanceIds=t._availableInstanceIds.slice(),this._availableGeometryIds=t._availableGeometryIds.slice(),this._nextIndexStart=t._nextIndexStart,this._nextVertexStart=t._nextVertexStart,this._geometryCount=t._geometryCount,this._maxInstanceCount=t._maxInstanceCount,this._maxVertexCount=t._maxVertexCount,this._maxIndexCount=t._maxIndexCount,this._geometryInitialized=t._geometryInitialized,this._multiDrawCounts=t._multiDrawCounts.slice(),this._multiDrawStarts=t._multiDrawStarts.slice(),this._indirectTexture=t._indirectTexture.clone(),this._indirectTexture.image.data=this._indirectTexture.image.data.slice(),this._matricesTexture=t._matricesTexture.clone(),this._matricesTexture.image.data=this._matricesTexture.image.data.slice(),this._colorsTexture!==null&&(this._colorsTexture=t._colorsTexture.clone(),this._colorsTexture.image.data=this._colorsTexture.image.data.slice()),this}dispose(){this.geometry.dispose(),this._matricesTexture.dispose(),this._matricesTexture=null,this._indirectTexture.dispose(),this._indirectTexture=null,this._colorsTexture!==null&&(this._colorsTexture.dispose(),this._colorsTexture=null)}onBeforeRender(t,e,n,i,s){if(!this._visibilityChanged&&!this.perObjectFrustumCulled&&!this.sortObjects)return;const r=i.getIndex();let a=r===null?1:r.array.BYTES_PER_ELEMENT,l=1;s.wireframe&&(l=2,a=i.attributes.position.count>65535?4:2);const c=this._instanceInfo,d=this._multiDrawStarts,m=this._multiDrawCounts,g=this._geometryInfo,_=this.perObjectFrustumCulled,v=this._indirectTexture,M=v.image.data,C=n.isArrayCamera?Rf:cu;_&&!n.isArrayCamera&&(In.multiplyMatrices(n.projectionMatrix,n.matrixWorldInverse).multiply(this.matrixWorld),cu.setFromProjectionMatrix(In,n.coordinateSystem,n.reversedDepth));let w=0;if(this.sortObjects){In.copy(this.matrixWorld).invert(),Fr.setFromMatrixPosition(n.matrixWorld).applyMatrix4(In),hu.set(0,0,-1).transformDirection(n.matrixWorld).transformDirection(In);for(let N=0,et=c.length;N<et;N++)if(c[N].visible&&c[N].active){const rt=c[N].geometryIndex;this.getMatrixAt(N,In),this.getBoundingSphereAt(rt,us).applyMatrix4(In);let vt=!1;if(_&&(vt=!C.intersectsSphere(us,n)),!vt){const at=g[rt],Mt=If.subVectors(us.center,Fr).dot(hu);Oc.push(at.start,at.count,Mt,N)}}const L=Oc.list,D=this.customSort;D===null?L.sort(s.transparent?Ef:Af):D.call(this,L,n);for(let N=0,et=L.length;N<et;N++){const rt=L[N];d[w]=rt.start*a*l,m[w]=rt.count*l,M[w]=rt.index,w++}Oc.reset()}else for(let L=0,D=c.length;L<D;L++)if(c[L].visible&&c[L].active){const N=c[L].geometryIndex;let et=!1;if(_&&(this.getMatrixAt(L,In),this.getBoundingSphereAt(N,us).applyMatrix4(In),et=!C.intersectsSphere(us,n)),!et){const rt=g[N];d[w]=rt.start*a*l,m[w]=rt.count*l,M[w]=L,w++}}v.needsUpdate=!0,this._multiDrawCount=w,this._visibilityChanged=!1}onBeforeShadow(t,e,n,i,s,r){this.onBeforeRender(t,null,i,s,r)}}class Pn extends Ft{constructor(t){super(),this.isLineBasicMaterial=!0,this.type="LineBasicMaterial",this.color=new re(16777215),this.map=null,this.linewidth=1,this.linecap="round",this.linejoin="round",this.fog=!0,this.setValues(t)}copy(t){return super.copy(t),this.color.copy(t.color),this.map=t.map,this.linewidth=t.linewidth,this.linecap=t.linecap,this.linejoin=t.linejoin,this.fog=t.fog,this}}const to=new O,eo=new O,uu=new Me,Or=new Vn,no=new ae,Bc=new O,fu=new O;class ds extends De{constructor(t=new y,e=new Pn){super(),this.isLine=!0,this.type="Line",this.geometry=t,this.material=e,this.morphTargetDictionary=void 0,this.morphTargetInfluences=void 0,this.updateMorphTargets()}copy(t,e){return super.copy(t,e),this.material=Array.isArray(t.material)?t.material.slice():t.material,this.geometry=t.geometry,this}computeLineDistances(){const t=this.geometry;if(t.index===null){const e=t.attributes.position,n=[0];for(let i=1,s=e.count;i<s;i++)to.fromBufferAttribute(e,i-1),eo.fromBufferAttribute(e,i),n[i]=n[i-1],n[i]+=to.distanceTo(eo);t.setAttribute("lineDistance",new Ut(n,1))}else le("Line.computeLineDistances(): Computation only possible with non-indexed BufferGeometry.");return this}raycast(t,e){const n=this.geometry,i=this.matrixWorld,s=t.params.Line.threshold,r=n.drawRange;if(n.boundingSphere===null&&n.computeBoundingSphere(),no.copy(n.boundingSphere),no.applyMatrix4(i),no.radius+=s,t.ray.intersectsSphere(no)===!1)return;uu.copy(i).invert(),Or.copy(t.ray).applyMatrix4(uu);const a=s/((this.scale.x+this.scale.y+this.scale.z)/3),l=a*a,c=this.isLineSegments?2:1,d=n.index,g=n.attributes.position;if(d!==null){const _=Math.max(0,r.start),v=Math.min(d.count,r.start+r.count);for(let M=_,C=v-1;M<C;M+=c){const w=d.getX(M),L=d.getX(M+1),D=io(this,t,Or,l,w,L,M);D&&e.push(D)}if(this.isLineLoop){const M=d.getX(v-1),C=d.getX(_),w=io(this,t,Or,l,M,C,v-1);w&&e.push(w)}}else{const _=Math.max(0,r.start),v=Math.min(g.count,r.start+r.count);for(let M=_,C=v-1;M<C;M+=c){const w=io(this,t,Or,l,M,M+1,M);w&&e.push(w)}if(this.isLineLoop){const M=io(this,t,Or,l,v-1,_,v-1);M&&e.push(M)}}}updateMorphTargets(){const e=this.geometry.morphAttributes,n=Object.keys(e);if(n.length>0){const i=e[n[0]];if(i!==void 0){this.morphTargetInfluences=[],this.morphTargetDictionary={};for(let s=0,r=i.length;s<r;s++){const a=i[s].name||String(s);this.morphTargetInfluences.push(0),this.morphTargetDictionary[a]=s}}}}}function io(u,t,e,n,i,s,r){const a=u.geometry.attributes.position;if(to.fromBufferAttribute(a,i),eo.fromBufferAttribute(a,s),e.distanceSqToSegment(to,eo,Bc,fu)>n)return;Bc.applyMatrix4(u.matrixWorld);const c=t.ray.origin.distanceTo(Bc);if(!(c<t.near||c>t.far))return{distance:c,point:fu.clone().applyMatrix4(u.matrixWorld),index:r,face:null,faceIndex:null,barycoord:null,object:u}}const du=new O,pu=new O;class Si extends ds{constructor(t,e){super(t,e),this.isLineSegments=!0,this.type="LineSegments"}computeLineDistances(){const t=this.geometry;if(t.index===null){const e=t.attributes.position,n=[];for(let i=0,s=e.count;i<s;i+=2)du.fromBufferAttribute(e,i),pu.fromBufferAttribute(e,i+1),n[i]=i===0?0:n[i-1],n[i+1]=n[i]+du.distanceTo(pu);t.setAttribute("lineDistance",new Ut(n,1))}else le("LineSegments.computeLineDistances(): Computation only possible with non-indexed BufferGeometry.");return this}}class Df extends ds{constructor(t,e){super(t,e),this.isLineLoop=!0,this.type="LineLoop"}}class mu extends Ft{constructor(t){super(),this.isPointsMaterial=!0,this.type="PointsMaterial",this.color=new re(16777215),this.map=null,this.alphaMap=null,this.size=1,this.sizeAttenuation=!0,this.fog=!0,this.setValues(t)}copy(t){return super.copy(t),this.color.copy(t.color),this.map=t.map,this.alphaMap=t.alphaMap,this.size=t.size,this.sizeAttenuation=t.sizeAttenuation,this.fog=t.fog,this}}const gu=new Me,zc=new Vn,so=new ae,ro=new O;class Uf extends De{constructor(t=new y,e=new mu){super(),this.isPoints=!0,this.type="Points",this.geometry=t,this.material=e,this.morphTargetDictionary=void 0,this.morphTargetInfluences=void 0,this.updateMorphTargets()}copy(t,e){return super.copy(t,e),this.material=Array.isArray(t.material)?t.material.slice():t.material,this.geometry=t.geometry,this}raycast(t,e){const n=this.geometry,i=this.matrixWorld,s=t.params.Points.threshold,r=n.drawRange;if(n.boundingSphere===null&&n.computeBoundingSphere(),so.copy(n.boundingSphere),so.applyMatrix4(i),so.radius+=s,t.ray.intersectsSphere(so)===!1)return;gu.copy(i).invert(),zc.copy(t.ray).applyMatrix4(gu);const a=s/((this.scale.x+this.scale.y+this.scale.z)/3),l=a*a,c=n.index,m=n.attributes.position;if(c!==null){const g=Math.max(0,r.start),_=Math.min(c.count,r.start+r.count);for(let v=g,M=_;v<M;v++){const C=c.getX(v);ro.fromBufferAttribute(m,C),_u(ro,C,l,i,t,e,this)}}else{const g=Math.max(0,r.start),_=Math.min(m.count,r.start+r.count);for(let v=g,M=_;v<M;v++)ro.fromBufferAttribute(m,v),_u(ro,v,l,i,t,e,this)}}updateMorphTargets(){const e=this.geometry.morphAttributes,n=Object.keys(e);if(n.length>0){const i=e[n[0]];if(i!==void 0){this.morphTargetInfluences=[],this.morphTargetDictionary={};for(let s=0,r=i.length;s<r;s++){const a=i[s].name||String(s);this.morphTargetInfluences.push(0),this.morphTargetDictionary[a]=s}}}}}function _u(u,t,e,n,i,s,r){const a=zc.distanceSqToPoint(u);if(a<e){const l=new O;zc.closestPointToPoint(u,l),l.applyMatrix4(n);const c=i.ray.origin.distanceTo(l);if(c<i.near||c>i.far)return;s.push({distance:c,distanceToRay:Math.sqrt(a),point:l,index:t,face:null,faceIndex:null,barycoord:null,object:r})}}class Nf extends Ye{constructor(t,e,n,i,s=Cn,r=Cn,a,l,c){super(t,e,n,i,s,r,a,l,c),this.isVideoTexture=!0,this.generateMipmaps=!1,this._requestVideoFrameCallbackId=0;const d=this;function m(){d.needsUpdate=!0,d._requestVideoFrameCallbackId=t.requestVideoFrameCallback(m)}"requestVideoFrameCallback"in t&&(this._requestVideoFrameCallbackId=t.requestVideoFrameCallback(m))}clone(){return new this.constructor(this.image).copy(this)}update(){const t=this.image;"requestVideoFrameCallback"in t===!1&&t.readyState>=t.HAVE_CURRENT_DATA&&(this.needsUpdate=!0)}dispose(){this._requestVideoFrameCallbackId!==0&&(this.source.data.cancelVideoFrameCallback(this._requestVideoFrameCallbackId),this._requestVideoFrameCallbackId=0),super.dispose()}}class Dp extends Nf{constructor(t,e,n,i,s,r,a,l){super({},t,e,n,i,s,r,a,l),this.isVideoFrameTexture=!0}update(){}clone(){return new this.constructor().copy(this)}setFrame(t){this.image=t,this.needsUpdate=!0}}class Up extends Ye{constructor(t,e){super({width:t,height:e}),this.isFramebufferTexture=!0,this.magFilter=Mn,this.minFilter=Mn,this.generateMipmaps=!1,this.needsUpdate=!0}}class Vc extends Ye{constructor(t,e,n,i,s,r,a,l,c,d,m,g){super(null,r,a,l,c,d,i,s,m,g),this.isCompressedTexture=!0,this.image={width:e,height:n},this.mipmaps=t,this.flipY=!1,this.generateMipmaps=!1}}class Np extends Vc{constructor(t,e,n,i,s,r){super(t,e,n,s,r),this.isCompressedArrayTexture=!0,this.image.depth=i,this.wrapR=Ln,this.layerUpdates=new Set}addLayerUpdate(t){this.layerUpdates.add(t)}clearLayerUpdates(){this.layerUpdates.clear()}}class Fp extends Vc{constructor(t,e,n){super(void 0,t[0].width,t[0].height,e,n,Yi),this.isCompressedCubeTexture=!0,this.isCubeTexture=!0,this.image=t}}class kc extends Ye{constructor(t=[],e=Yi,n,i,s,r,a,l,c,d){super(t,e,n,i,s,r,a,l,c,d),this.isCubeTexture=!0,this.flipY=!1}get images(){return this.image}set images(t){this.image=t}}class Op extends Ye{constructor(t,e,n,i,s,r,a,l,c){super(t,e,n,i,s,r,a,l,c),this.isCanvasTexture=!0,this.needsUpdate=!0}}class xu extends Ye{constructor(t,e,n=Zi,i,s,r,a=Mn,l=Mn,c,d=Ps,m=1){if(d!==Ps&&d!==da)throw new Error("DepthTexture format must be either THREE.DepthFormat or THREE.DepthStencilFormat");const g={width:t,height:e,depth:m};super(g,i,s,r,a,l,d,n,c),this.isDepthTexture=!0,this.flipY=!1,this.generateMipmaps=!1,this.compareFunction=null}copy(t){return super.copy(t),this.source=new oi(Object.assign({},t.image)),this.compareFunction=t.compareFunction,this}toJSON(t){const e=super.toJSON(t);return this.compareFunction!==null&&(e.compareFunction=this.compareFunction),e}}class Ff extends xu{constructor(t,e=Zi,n=Yi,i,s,r=Mn,a=Mn,l,c=Ps){const d={width:t,height:t,depth:1},m=[d,d,d,d,d,d];super(t,t,e,n,i,s,r,a,l,c),this.image=m,this.isCubeDepthTexture=!0,this.isCubeTexture=!0}get images(){return this.image}set images(t){this.image=t}}class Of extends Ye{constructor(t=null){super(),this.sourceTexture=t,this.isExternalTexture=!0}copy(t){return super.copy(t),this.sourceTexture=t.sourceTexture,this}}class ao extends y{constructor(t=1,e=1,n=1,i=1,s=1,r=1){super(),this.type="BoxGeometry",this.parameters={width:t,height:e,depth:n,widthSegments:i,heightSegments:s,depthSegments:r};const a=this;i=Math.floor(i),s=Math.floor(s),r=Math.floor(r);const l=[],c=[],d=[],m=[];let g=0,_=0;v("z","y","x",-1,-1,n,e,t,r,s,0),v("z","y","x",1,-1,n,e,-t,r,s,1),v("x","z","y",1,1,t,n,e,i,r,2),v("x","z","y",1,-1,t,n,-e,i,r,3),v("x","y","z",1,-1,t,e,n,i,s,4),v("x","y","z",-1,-1,t,e,-n,i,s,5),this.setIndex(l),this.setAttribute("position",new Ut(c,3)),this.setAttribute("normal",new Ut(d,3)),this.setAttribute("uv",new Ut(m,2));function v(M,C,w,L,D,N,et,rt,vt,at,Mt){const St=N/vt,Gt=et/at,he=N/2,ge=et/2,Ne=rt/2,xe=vt+1,nn=at+1;let cn=0,ii=0;const sn=new O;for(let hn=0;hn<nn;hn++){const rn=hn*Gt-ge;for(let si=0;si<xe;si++){const $n=si*St-he;sn[M]=$n*L,sn[C]=rn*D,sn[w]=Ne,c.push(sn.x,sn.y,sn.z),sn[M]=0,sn[C]=0,sn[w]=rt>0?1:-1,d.push(sn.x,sn.y,sn.z),m.push(si/vt),m.push(1-hn/at),cn+=1}}for(let hn=0;hn<at;hn++)for(let rn=0;rn<vt;rn++){const si=g+rn+xe*hn,$n=g+rn+xe*(hn+1),Js=g+(rn+1)+xe*(hn+1),Xr=g+(rn+1)+xe*hn;l.push(si,$n,Xr),l.push($n,Js,Xr),ii+=6}a.addGroup(_,ii,Mt),_+=ii,g+=cn}}copy(t){return super.copy(t),this.parameters=Object.assign({},t.parameters),this}static fromJSON(t){return new ao(t.width,t.height,t.depth,t.widthSegments,t.heightSegments,t.depthSegments)}}class Gc extends y{constructor(t=1,e=1,n=4,i=8,s=1){super(),this.type="CapsuleGeometry",this.parameters={radius:t,height:e,capSegments:n,radialSegments:i,heightSegments:s},e=Math.max(0,e),n=Math.max(1,Math.floor(n)),i=Math.max(3,Math.floor(i)),s=Math.max(1,Math.floor(s));const r=[],a=[],l=[],c=[],d=e/2,m=Math.PI/2*t,g=e,_=2*m+g,v=n*2+s,M=i+1,C=new O,w=new O;for(let L=0;L<=v;L++){let D=0,N=0,et=0,rt=0;if(L<=n){const Mt=L/n,St=Mt*Math.PI/2;N=-d-t*Math.cos(St),et=t*Math.sin(St),rt=-t*Math.cos(St),D=Mt*m}else if(L<=n+s){const Mt=(L-n)/s;N=-d+Mt*e,et=t,rt=0,D=m+Mt*g}else{const Mt=(L-n-s)/n,St=Mt*Math.PI/2;N=d+t*Math.sin(St),et=t*Math.cos(St),rt=t*Math.sin(St),D=m+g+Mt*m}const vt=Math.max(0,Math.min(1,D/_));let at=0;L===0?at=.5/i:L===v&&(at=-.5/i);for(let Mt=0;Mt<=i;Mt++){const St=Mt/i,Gt=St*Math.PI*2,he=Math.sin(Gt),ge=Math.cos(Gt);w.x=-et*ge,w.y=N,w.z=et*he,a.push(w.x,w.y,w.z),C.set(-et*ge,rt,et*he),C.normalize(),l.push(C.x,C.y,C.z),c.push(St+at,vt)}if(L>0){const Mt=(L-1)*M;for(let St=0;St<i;St++){const Gt=Mt+St,he=Mt+St+1,ge=L*M+St,Ne=L*M+St+1;r.push(Gt,he,ge),r.push(he,Ne,ge)}}}this.setIndex(r),this.setAttribute("position",new Ut(a,3)),this.setAttribute("normal",new Ut(l,3)),this.setAttribute("uv",new Ut(c,2))}copy(t){return super.copy(t),this.parameters=Object.assign({},t.parameters),this}static fromJSON(t){return new Gc(t.radius,t.height,t.capSegments,t.radialSegments,t.heightSegments)}}class Hc extends y{constructor(t=1,e=32,n=0,i=Math.PI*2){super(),this.type="CircleGeometry",this.parameters={radius:t,segments:e,thetaStart:n,thetaLength:i},e=Math.max(3,e);const s=[],r=[],a=[],l=[],c=new O,d=new At;r.push(0,0,0),a.push(0,0,1),l.push(.5,.5);for(let m=0,g=3;m<=e;m++,g+=3){const _=n+m/e*i;c.x=t*Math.cos(_),c.y=t*Math.sin(_),r.push(c.x,c.y,c.z),a.push(0,0,1),d.x=(r[g]/t+1)/2,d.y=(r[g+1]/t+1)/2,l.push(d.x,d.y)}for(let m=1;m<=e;m++)s.push(m,m+1,0);this.setIndex(s),this.setAttribute("position",new Ut(r,3)),this.setAttribute("normal",new Ut(a,3)),this.setAttribute("uv",new Ut(l,2))}copy(t){return super.copy(t),this.parameters=Object.assign({},t.parameters),this}static fromJSON(t){return new Hc(t.radius,t.segments,t.thetaStart,t.thetaLength)}}class oo extends y{constructor(t=1,e=1,n=1,i=32,s=1,r=!1,a=0,l=Math.PI*2){super(),this.type="CylinderGeometry",this.parameters={radiusTop:t,radiusBottom:e,height:n,radialSegments:i,heightSegments:s,openEnded:r,thetaStart:a,thetaLength:l};const c=this;i=Math.floor(i),s=Math.floor(s);const d=[],m=[],g=[],_=[];let v=0;const M=[],C=n/2;let w=0;L(),r===!1&&(t>0&&D(!0),e>0&&D(!1)),this.setIndex(d),this.setAttribute("position",new Ut(m,3)),this.setAttribute("normal",new Ut(g,3)),this.setAttribute("uv",new Ut(_,2));function L(){const N=new O,et=new O;let rt=0;const vt=(e-t)/n;for(let at=0;at<=s;at++){const Mt=[],St=at/s,Gt=St*(e-t)+t;for(let he=0;he<=i;he++){const ge=he/i,Ne=ge*l+a,xe=Math.sin(Ne),nn=Math.cos(Ne);et.x=Gt*xe,et.y=-St*n+C,et.z=Gt*nn,m.push(et.x,et.y,et.z),N.set(xe,vt,nn).normalize(),g.push(N.x,N.y,N.z),_.push(ge,1-St),Mt.push(v++)}M.push(Mt)}for(let at=0;at<i;at++)for(let Mt=0;Mt<s;Mt++){const St=M[Mt][at],Gt=M[Mt+1][at],he=M[Mt+1][at+1],ge=M[Mt][at+1];(t>0||Mt!==0)&&(d.push(St,Gt,ge),rt+=3),(e>0||Mt!==s-1)&&(d.push(Gt,he,ge),rt+=3)}c.addGroup(w,rt,0),w+=rt}function D(N){const et=v,rt=new At,vt=new O;let at=0;const Mt=N===!0?t:e,St=N===!0?1:-1;for(let he=1;he<=i;he++)m.push(0,C*St,0),g.push(0,St,0),_.push(.5,.5),v++;const Gt=v;for(let he=0;he<=i;he++){const Ne=he/i*l+a,xe=Math.cos(Ne),nn=Math.sin(Ne);vt.x=Mt*nn,vt.y=C*St,vt.z=Mt*xe,m.push(vt.x,vt.y,vt.z),g.push(0,St,0),rt.x=xe*.5+.5,rt.y=nn*.5*St+.5,_.push(rt.x,rt.y),v++}for(let he=0;he<i;he++){const ge=et+he,Ne=Gt+he;N===!0?d.push(Ne,Ne+1,ge):d.push(Ne+1,Ne,ge),at+=3}c.addGroup(w,at,N===!0?1:2),w+=at}}copy(t){return super.copy(t),this.parameters=Object.assign({},t.parameters),this}static fromJSON(t){return new oo(t.radiusTop,t.radiusBottom,t.height,t.radialSegments,t.heightSegments,t.openEnded,t.thetaStart,t.thetaLength)}}class lo extends oo{constructor(t=1,e=1,n=32,i=1,s=!1,r=0,a=Math.PI*2){super(0,t,e,n,i,s,r,a),this.type="ConeGeometry",this.parameters={radius:t,height:e,radialSegments:n,heightSegments:i,openEnded:s,thetaStart:r,thetaLength:a}}static fromJSON(t){return new lo(t.radius,t.height,t.radialSegments,t.heightSegments,t.openEnded,t.thetaStart,t.thetaLength)}}class ps extends y{constructor(t=[],e=[],n=1,i=0){super(),this.type="PolyhedronGeometry",this.parameters={vertices:t,indices:e,radius:n,detail:i};const s=[],r=[];a(i),c(n),d(),this.setAttribute("position",new Ut(s,3)),this.setAttribute("normal",new Ut(s.slice(),3)),this.setAttribute("uv",new Ut(r,2)),i===0?this.computeVertexNormals():this.normalizeNormals();function a(L){const D=new O,N=new O,et=new O;for(let rt=0;rt<e.length;rt+=3)_(e[rt+0],D),_(e[rt+1],N),_(e[rt+2],et),l(D,N,et,L)}function l(L,D,N,et){const rt=et+1,vt=[];for(let at=0;at<=rt;at++){vt[at]=[];const Mt=L.clone().lerp(N,at/rt),St=D.clone().lerp(N,at/rt),Gt=rt-at;for(let he=0;he<=Gt;he++)he===0&&at===rt?vt[at][he]=Mt:vt[at][he]=Mt.clone().lerp(St,he/Gt)}for(let at=0;at<rt;at++)for(let Mt=0;Mt<2*(rt-at)-1;Mt++){const St=Math.floor(Mt/2);Mt%2===0?(g(vt[at][St+1]),g(vt[at+1][St]),g(vt[at][St])):(g(vt[at][St+1]),g(vt[at+1][St+1]),g(vt[at+1][St]))}}function c(L){const D=new O;for(let N=0;N<s.length;N+=3)D.x=s[N+0],D.y=s[N+1],D.z=s[N+2],D.normalize().multiplyScalar(L),s[N+0]=D.x,s[N+1]=D.y,s[N+2]=D.z}function d(){const L=new O;for(let D=0;D<s.length;D+=3){L.x=s[D+0],L.y=s[D+1],L.z=s[D+2];const N=C(L)/2/Math.PI+.5,et=w(L)/Math.PI+.5;r.push(N,1-et)}v(),m()}function m(){for(let L=0;L<r.length;L+=6){const D=r[L+0],N=r[L+2],et=r[L+4],rt=Math.max(D,N,et),vt=Math.min(D,N,et);rt>.9&&vt<.1&&(D<.2&&(r[L+0]+=1),N<.2&&(r[L+2]+=1),et<.2&&(r[L+4]+=1))}}function g(L){s.push(L.x,L.y,L.z)}function _(L,D){const N=L*3;D.x=t[N+0],D.y=t[N+1],D.z=t[N+2]}function v(){const L=new O,D=new O,N=new O,et=new O,rt=new At,vt=new At,at=new At;for(let Mt=0,St=0;Mt<s.length;Mt+=9,St+=6){L.set(s[Mt+0],s[Mt+1],s[Mt+2]),D.set(s[Mt+3],s[Mt+4],s[Mt+5]),N.set(s[Mt+6],s[Mt+7],s[Mt+8]),rt.set(r[St+0],r[St+1]),vt.set(r[St+2],r[St+3]),at.set(r[St+4],r[St+5]),et.copy(L).add(D).add(N).divideScalar(3);const Gt=C(et);M(rt,St+0,L,Gt),M(vt,St+2,D,Gt),M(at,St+4,N,Gt)}}function M(L,D,N,et){et<0&&L.x===1&&(r[D]=L.x-1),N.x===0&&N.z===0&&(r[D]=et/2/Math.PI+.5)}function C(L){return Math.atan2(L.z,-L.x)}function w(L){return Math.atan2(-L.y,Math.sqrt(L.x*L.x+L.z*L.z))}}copy(t){return super.copy(t),this.parameters=Object.assign({},t.parameters),this}static fromJSON(t){return new ps(t.vertices,t.indices,t.radius,t.detail)}}class Wc extends ps{constructor(t=1,e=0){const n=(1+Math.sqrt(5))/2,i=1/n,s=[-1,-1,-1,-1,-1,1,-1,1,-1,-1,1,1,1,-1,-1,1,-1,1,1,1,-1,1,1,1,0,-i,-n,0,-i,n,0,i,-n,0,i,n,-i,-n,0,-i,n,0,i,-n,0,i,n,0,-n,0,-i,n,0,-i,-n,0,i,n,0,i],r=[3,11,7,3,7,15,3,15,13,7,19,17,7,17,6,7,6,15,17,4,8,17,8,10,17,10,6,8,0,16,8,16,2,8,2,10,0,12,1,0,1,18,0,18,16,6,10,2,6,2,13,6,13,15,2,16,18,2,18,3,2,3,13,18,1,9,18,9,11,18,11,3,4,14,12,4,12,0,4,0,8,11,9,5,11,5,19,11,19,7,19,5,14,19,14,4,19,4,17,1,12,14,1,14,5,1,5,9];super(s,r,t,e),this.type="DodecahedronGeometry",this.parameters={radius:t,detail:e}}static fromJSON(t){return new Wc(t.radius,t.detail)}}const co=new O,ho=new O,Xc=new O,uo=new nt;class Bf extends y{constructor(t=null,e=1){if(super(),this.type="EdgesGeometry",this.parameters={geometry:t,thresholdAngle:e},t!==null){const i=Math.pow(10,4),s=Math.cos(_i*e),r=t.getIndex(),a=t.getAttribute("position"),l=r?r.count:a.count,c=[0,0,0],d=["a","b","c"],m=new Array(3),g={},_=[];for(let v=0;v<l;v+=3){r?(c[0]=r.getX(v),c[1]=r.getX(v+1),c[2]=r.getX(v+2)):(c[0]=v,c[1]=v+1,c[2]=v+2);const{a:M,b:C,c:w}=uo;if(M.fromBufferAttribute(a,c[0]),C.fromBufferAttribute(a,c[1]),w.fromBufferAttribute(a,c[2]),uo.getNormal(Xc),m[0]=`${Math.round(M.x*i)},${Math.round(M.y*i)},${Math.round(M.z*i)}`,m[1]=`${Math.round(C.x*i)},${Math.round(C.y*i)},${Math.round(C.z*i)}`,m[2]=`${Math.round(w.x*i)},${Math.round(w.y*i)},${Math.round(w.z*i)}`,!(m[0]===m[1]||m[1]===m[2]||m[2]===m[0]))for(let L=0;L<3;L++){const D=(L+1)%3,N=m[L],et=m[D],rt=uo[d[L]],vt=uo[d[D]],at=`${N}_${et}`,Mt=`${et}_${N}`;Mt in g&&g[Mt]?(Xc.dot(g[Mt].normal)<=s&&(_.push(rt.x,rt.y,rt.z),_.push(vt.x,vt.y,vt.z)),g[Mt]=null):at in g||(g[at]={index0:c[L],index1:c[D],normal:Xc.clone()})}}for(const v in g)if(g[v]){const{index0:M,index1:C}=g[v];co.fromBufferAttribute(a,M),ho.fromBufferAttribute(a,C),_.push(co.x,co.y,co.z),_.push(ho.x,ho.y,ho.z)}this.setAttribute("position",new Ut(_,3))}}copy(t){return super.copy(t),this.parameters=Object.assign({},t.parameters),this}}class di{constructor(){this.type="Curve",this.arcLengthDivisions=200,this.needsUpdate=!1,this.cacheArcLengths=null}getPoint(){le("Curve: .getPoint() not implemented.")}getPointAt(t,e){const n=this.getUtoTmapping(t);return this.getPoint(n,e)}getPoints(t=5){const e=[];for(let n=0;n<=t;n++)e.push(this.getPoint(n/t));return e}getSpacedPoints(t=5){const e=[];for(let n=0;n<=t;n++)e.push(this.getPointAt(n/t));return e}getLength(){const t=this.getLengths();return t[t.length-1]}getLengths(t=this.arcLengthDivisions){if(this.cacheArcLengths&&this.cacheArcLengths.length===t+1&&!this.needsUpdate)return this.cacheArcLengths;this.needsUpdate=!1;const e=[];let n,i=this.getPoint(0),s=0;e.push(0);for(let r=1;r<=t;r++)n=this.getPoint(r/t),s+=n.distanceTo(i),e.push(s),i=n;return this.cacheArcLengths=e,e}updateArcLengths(){this.needsUpdate=!0,this.getLengths()}getUtoTmapping(t,e=null){const n=this.getLengths();let i=0;const s=n.length;let r;e?r=e:r=t*n[s-1];let a=0,l=s-1,c;for(;a<=l;)if(i=Math.floor(a+(l-a)/2),c=n[i]-r,c<0)a=i+1;else if(c>0)l=i-1;else{l=i;break}if(i=l,n[i]===r)return i/(s-1);const d=n[i],g=n[i+1]-d,_=(r-d)/g;return(i+_)/(s-1)}getTangent(t,e){let i=t-1e-4,s=t+1e-4;i<0&&(i=0),s>1&&(s=1);const r=this.getPoint(i),a=this.getPoint(s),l=e||(r.isVector2?new At:new O);return l.copy(a).sub(r).normalize(),l}getTangentAt(t,e){const n=this.getUtoTmapping(t);return this.getTangent(n,e)}computeFrenetFrames(t,e=!1){const n=new O,i=[],s=[],r=[],a=new O,l=new Me;for(let _=0;_<=t;_++){const v=_/t;i[_]=this.getTangentAt(v,new O)}s[0]=new O,r[0]=new O;let c=Number.MAX_VALUE;const d=Math.abs(i[0].x),m=Math.abs(i[0].y),g=Math.abs(i[0].z);d<=c&&(c=d,n.set(1,0,0)),m<=c&&(c=m,n.set(0,1,0)),g<=c&&n.set(0,0,1),a.crossVectors(i[0],n).normalize(),s[0].crossVectors(i[0],a),r[0].crossVectors(i[0],s[0]);for(let _=1;_<=t;_++){if(s[_]=s[_-1].clone(),r[_]=r[_-1].clone(),a.crossVectors(i[_-1],i[_]),a.length()>Number.EPSILON){a.normalize();const v=Math.acos(pe(i[_-1].dot(i[_]),-1,1));s[_].applyMatrix4(l.makeRotationAxis(a,v))}r[_].crossVectors(i[_],s[_])}if(e===!0){let _=Math.acos(pe(s[0].dot(s[t]),-1,1));_/=t,i[0].dot(a.crossVectors(s[0],s[t]))>0&&(_=-_);for(let v=1;v<=t;v++)s[v].applyMatrix4(l.makeRotationAxis(i[v],_*v)),r[v].crossVectors(i[v],s[v])}return{tangents:i,normals:s,binormals:r}}clone(){return new this.constructor().copy(this)}copy(t){return this.arcLengthDivisions=t.arcLengthDivisions,this}toJSON(){const t={metadata:{version:4.7,type:"Curve",generator:"Curve.toJSON"}};return t.arcLengthDivisions=this.arcLengthDivisions,t.type=this.type,t}fromJSON(t){return this.arcLengthDivisions=t.arcLengthDivisions,this}}class qc extends di{constructor(t=0,e=0,n=1,i=1,s=0,r=Math.PI*2,a=!1,l=0){super(),this.isEllipseCurve=!0,this.type="EllipseCurve",this.aX=t,this.aY=e,this.xRadius=n,this.yRadius=i,this.aStartAngle=s,this.aEndAngle=r,this.aClockwise=a,this.aRotation=l}getPoint(t,e=new At){const n=e,i=Math.PI*2;let s=this.aEndAngle-this.aStartAngle;const r=Math.abs(s)<Number.EPSILON;for(;s<0;)s+=i;for(;s>i;)s-=i;s<Number.EPSILON&&(r?s=0:s=i),this.aClockwise===!0&&!r&&(s===i?s=-i:s=s-i);const a=this.aStartAngle+t*s;let l=this.aX+this.xRadius*Math.cos(a),c=this.aY+this.yRadius*Math.sin(a);if(this.aRotation!==0){const d=Math.cos(this.aRotation),m=Math.sin(this.aRotation),g=l-this.aX,_=c-this.aY;l=g*d-_*m+this.aX,c=g*m+_*d+this.aY}return n.set(l,c)}copy(t){return super.copy(t),this.aX=t.aX,this.aY=t.aY,this.xRadius=t.xRadius,this.yRadius=t.yRadius,this.aStartAngle=t.aStartAngle,this.aEndAngle=t.aEndAngle,this.aClockwise=t.aClockwise,this.aRotation=t.aRotation,this}toJSON(){const t=super.toJSON();return t.aX=this.aX,t.aY=this.aY,t.xRadius=this.xRadius,t.yRadius=this.yRadius,t.aStartAngle=this.aStartAngle,t.aEndAngle=this.aEndAngle,t.aClockwise=this.aClockwise,t.aRotation=this.aRotation,t}fromJSON(t){return super.fromJSON(t),this.aX=t.aX,this.aY=t.aY,this.xRadius=t.xRadius,this.yRadius=t.yRadius,this.aStartAngle=t.aStartAngle,this.aEndAngle=t.aEndAngle,this.aClockwise=t.aClockwise,this.aRotation=t.aRotation,this}}class zf extends qc{constructor(t,e,n,i,s,r){super(t,e,n,n,i,s,r),this.isArcCurve=!0,this.type="ArcCurve"}}function Yc(){let u=0,t=0,e=0,n=0;function i(s,r,a,l){u=s,t=a,e=-3*s+3*r-2*a-l,n=2*s-2*r+a+l}return{initCatmullRom:function(s,r,a,l,c){i(r,a,c*(a-s),c*(l-r))},initNonuniformCatmullRom:function(s,r,a,l,c,d,m){let g=(r-s)/c-(a-s)/(c+d)+(a-r)/d,_=(a-r)/d-(l-r)/(d+m)+(l-a)/m;g*=d,_*=d,i(r,a,g,_)},calc:function(s){const r=s*s,a=r*s;return u+t*s+e*r+n*a}}}const fo=new O,Zc=new Yc,$c=new Yc,Jc=new Yc;class Vf extends di{constructor(t=[],e=!1,n="centripetal",i=.5){super(),this.isCatmullRomCurve3=!0,this.type="CatmullRomCurve3",this.points=t,this.closed=e,this.curveType=n,this.tension=i}getPoint(t,e=new O){const n=e,i=this.points,s=i.length,r=(s-(this.closed?0:1))*t;let a=Math.floor(r),l=r-a;this.closed?a+=a>0?0:(Math.floor(Math.abs(a)/s)+1)*s:l===0&&a===s-1&&(a=s-2,l=1);let c,d;this.closed||a>0?c=i[(a-1)%s]:(fo.subVectors(i[0],i[1]).add(i[0]),c=fo);const m=i[a%s],g=i[(a+1)%s];if(this.closed||a+2<s?d=i[(a+2)%s]:(fo.subVectors(i[s-1],i[s-2]).add(i[s-1]),d=fo),this.curveType==="centripetal"||this.curveType==="chordal"){const _=this.curveType==="chordal"?.5:.25;let v=Math.pow(c.distanceToSquared(m),_),M=Math.pow(m.distanceToSquared(g),_),C=Math.pow(g.distanceToSquared(d),_);M<1e-4&&(M=1),v<1e-4&&(v=M),C<1e-4&&(C=M),Zc.initNonuniformCatmullRom(c.x,m.x,g.x,d.x,v,M,C),$c.initNonuniformCatmullRom(c.y,m.y,g.y,d.y,v,M,C),Jc.initNonuniformCatmullRom(c.z,m.z,g.z,d.z,v,M,C)}else this.curveType==="catmullrom"&&(Zc.initCatmullRom(c.x,m.x,g.x,d.x,this.tension),$c.initCatmullRom(c.y,m.y,g.y,d.y,this.tension),Jc.initCatmullRom(c.z,m.z,g.z,d.z,this.tension));return n.set(Zc.calc(l),$c.calc(l),Jc.calc(l)),n}copy(t){super.copy(t),this.points=[];for(let e=0,n=t.points.length;e<n;e++){const i=t.points[e];this.points.push(i.clone())}return this.closed=t.closed,this.curveType=t.curveType,this.tension=t.tension,this}toJSON(){const t=super.toJSON();t.points=[];for(let e=0,n=this.points.length;e<n;e++){const i=this.points[e];t.points.push(i.toArray())}return t.closed=this.closed,t.curveType=this.curveType,t.tension=this.tension,t}fromJSON(t){super.fromJSON(t),this.points=[];for(let e=0,n=t.points.length;e<n;e++){const i=t.points[e];this.points.push(new O().fromArray(i))}return this.closed=t.closed,this.curveType=t.curveType,this.tension=t.tension,this}}function vu(u,t,e,n,i){const s=(n-t)*.5,r=(i-e)*.5,a=u*u,l=u*a;return(2*e-2*n+s+r)*l+(-3*e+3*n-2*s-r)*a+s*u+e}function kf(u,t){const e=1-u;return e*e*t}function Gf(u,t){return 2*(1-u)*u*t}function Hf(u,t){return u*u*t}function Br(u,t,e,n){return kf(u,t)+Gf(u,e)+Hf(u,n)}function Wf(u,t){const e=1-u;return e*e*e*t}function Xf(u,t){const e=1-u;return 3*e*e*u*t}function qf(u,t){return 3*(1-u)*u*u*t}function Yf(u,t){return u*u*u*t}function zr(u,t,e,n,i){return Wf(u,t)+Xf(u,e)+qf(u,n)+Yf(u,i)}class yu extends di{constructor(t=new At,e=new At,n=new At,i=new At){super(),this.isCubicBezierCurve=!0,this.type="CubicBezierCurve",this.v0=t,this.v1=e,this.v2=n,this.v3=i}getPoint(t,e=new At){const n=e,i=this.v0,s=this.v1,r=this.v2,a=this.v3;return n.set(zr(t,i.x,s.x,r.x,a.x),zr(t,i.y,s.y,r.y,a.y)),n}copy(t){return super.copy(t),this.v0.copy(t.v0),this.v1.copy(t.v1),this.v2.copy(t.v2),this.v3.copy(t.v3),this}toJSON(){const t=super.toJSON();return t.v0=this.v0.toArray(),t.v1=this.v1.toArray(),t.v2=this.v2.toArray(),t.v3=this.v3.toArray(),t}fromJSON(t){return super.fromJSON(t),this.v0.fromArray(t.v0),this.v1.fromArray(t.v1),this.v2.fromArray(t.v2),this.v3.fromArray(t.v3),this}}class Zf extends di{constructor(t=new O,e=new O,n=new O,i=new O){super(),this.isCubicBezierCurve3=!0,this.type="CubicBezierCurve3",this.v0=t,this.v1=e,this.v2=n,this.v3=i}getPoint(t,e=new O){const n=e,i=this.v0,s=this.v1,r=this.v2,a=this.v3;return n.set(zr(t,i.x,s.x,r.x,a.x),zr(t,i.y,s.y,r.y,a.y),zr(t,i.z,s.z,r.z,a.z)),n}copy(t){return super.copy(t),this.v0.copy(t.v0),this.v1.copy(t.v1),this.v2.copy(t.v2),this.v3.copy(t.v3),this}toJSON(){const t=super.toJSON();return t.v0=this.v0.toArray(),t.v1=this.v1.toArray(),t.v2=this.v2.toArray(),t.v3=this.v3.toArray(),t}fromJSON(t){return super.fromJSON(t),this.v0.fromArray(t.v0),this.v1.fromArray(t.v1),this.v2.fromArray(t.v2),this.v3.fromArray(t.v3),this}}class Mu extends di{constructor(t=new At,e=new At){super(),this.isLineCurve=!0,this.type="LineCurve",this.v1=t,this.v2=e}getPoint(t,e=new At){const n=e;return t===1?n.copy(this.v2):(n.copy(this.v2).sub(this.v1),n.multiplyScalar(t).add(this.v1)),n}getPointAt(t,e){return this.getPoint(t,e)}getTangent(t,e=new At){return e.subVectors(this.v2,this.v1).normalize()}getTangentAt(t,e){return this.getTangent(t,e)}copy(t){return super.copy(t),this.v1.copy(t.v1),this.v2.copy(t.v2),this}toJSON(){const t=super.toJSON();return t.v1=this.v1.toArray(),t.v2=this.v2.toArray(),t}fromJSON(t){return super.fromJSON(t),this.v1.fromArray(t.v1),this.v2.fromArray(t.v2),this}}class $f extends di{constructor(t=new O,e=new O){super(),this.isLineCurve3=!0,this.type="LineCurve3",this.v1=t,this.v2=e}getPoint(t,e=new O){const n=e;return t===1?n.copy(this.v2):(n.copy(this.v2).sub(this.v1),n.multiplyScalar(t).add(this.v1)),n}getPointAt(t,e){return this.getPoint(t,e)}getTangent(t,e=new O){return e.subVectors(this.v2,this.v1).normalize()}getTangentAt(t,e){return this.getTangent(t,e)}copy(t){return super.copy(t),this.v1.copy(t.v1),this.v2.copy(t.v2),this}toJSON(){const t=super.toJSON();return t.v1=this.v1.toArray(),t.v2=this.v2.toArray(),t}fromJSON(t){return super.fromJSON(t),this.v1.fromArray(t.v1),this.v2.fromArray(t.v2),this}}class Su extends di{constructor(t=new At,e=new At,n=new At){super(),this.isQuadraticBezierCurve=!0,this.type="QuadraticBezierCurve",this.v0=t,this.v1=e,this.v2=n}getPoint(t,e=new At){const n=e,i=this.v0,s=this.v1,r=this.v2;return n.set(Br(t,i.x,s.x,r.x),Br(t,i.y,s.y,r.y)),n}copy(t){return super.copy(t),this.v0.copy(t.v0),this.v1.copy(t.v1),this.v2.copy(t.v2),this}toJSON(){const t=super.toJSON();return t.v0=this.v0.toArray(),t.v1=this.v1.toArray(),t.v2=this.v2.toArray(),t}fromJSON(t){return super.fromJSON(t),this.v0.fromArray(t.v0),this.v1.fromArray(t.v1),this.v2.fromArray(t.v2),this}}class bu extends di{constructor(t=new O,e=new O,n=new O){super(),this.isQuadraticBezierCurve3=!0,this.type="QuadraticBezierCurve3",this.v0=t,this.v1=e,this.v2=n}getPoint(t,e=new O){const n=e,i=this.v0,s=this.v1,r=this.v2;return n.set(Br(t,i.x,s.x,r.x),Br(t,i.y,s.y,r.y),Br(t,i.z,s.z,r.z)),n}copy(t){return super.copy(t),this.v0.copy(t.v0),this.v1.copy(t.v1),this.v2.copy(t.v2),this}toJSON(){const t=super.toJSON();return t.v0=this.v0.toArray(),t.v1=this.v1.toArray(),t.v2=this.v2.toArray(),t}fromJSON(t){return super.fromJSON(t),this.v0.fromArray(t.v0),this.v1.fromArray(t.v1),this.v2.fromArray(t.v2),this}}class Tu extends di{constructor(t=[]){super(),this.isSplineCurve=!0,this.type="SplineCurve",this.points=t}getPoint(t,e=new At){const n=e,i=this.points,s=(i.length-1)*t,r=Math.floor(s),a=s-r,l=i[r===0?r:r-1],c=i[r],d=i[r>i.length-2?i.length-1:r+1],m=i[r>i.length-3?i.length-1:r+2];return n.set(vu(a,l.x,c.x,d.x,m.x),vu(a,l.y,c.y,d.y,m.y)),n}copy(t){super.copy(t),this.points=[];for(let e=0,n=t.points.length;e<n;e++){const i=t.points[e];this.points.push(i.clone())}return this}toJSON(){const t=super.toJSON();t.points=[];for(let e=0,n=this.points.length;e<n;e++){const i=this.points[e];t.points.push(i.toArray())}return t}fromJSON(t){super.fromJSON(t),this.points=[];for(let e=0,n=t.points.length;e<n;e++){const i=t.points[e];this.points.push(new At().fromArray(i))}return this}}var po=Object.freeze({__proto__:null,ArcCurve:zf,CatmullRomCurve3:Vf,CubicBezierCurve:yu,CubicBezierCurve3:Zf,EllipseCurve:qc,LineCurve:Mu,LineCurve3:$f,QuadraticBezierCurve:Su,QuadraticBezierCurve3:bu,SplineCurve:Tu});class Jf extends di{constructor(){super(),this.type="CurvePath",this.curves=[],this.autoClose=!1}add(t){this.curves.push(t)}closePath(){const t=this.curves[0].getPoint(0),e=this.curves[this.curves.length-1].getPoint(1);if(!t.equals(e)){const n=t.isVector2===!0?"LineCurve":"LineCurve3";this.curves.push(new po[n](e,t))}return this}getPoint(t,e){const n=t*this.getLength(),i=this.getCurveLengths();let s=0;for(;s<i.length;){if(i[s]>=n){const r=i[s]-n,a=this.curves[s],l=a.getLength(),c=l===0?0:1-r/l;return a.getPointAt(c,e)}s++}return null}getLength(){const t=this.getCurveLengths();return t[t.length-1]}updateArcLengths(){this.needsUpdate=!0,this.cacheLengths=null,this.getCurveLengths()}getCurveLengths(){if(this.cacheLengths&&this.cacheLengths.length===this.curves.length)return this.cacheLengths;const t=[];let e=0;for(let n=0,i=this.curves.length;n<i;n++)e+=this.curves[n].getLength(),t.push(e);return this.cacheLengths=t,t}getSpacedPoints(t=40){const e=[];for(let n=0;n<=t;n++)e.push(this.getPoint(n/t));return this.autoClose&&e.push(e[0]),e}getPoints(t=12){const e=[];let n;for(let i=0,s=this.curves;i<s.length;i++){const r=s[i],a=r.isEllipseCurve?t*2:r.isLineCurve||r.isLineCurve3?1:r.isSplineCurve?t*r.points.length:t,l=r.getPoints(a);for(let c=0;c<l.length;c++){const d=l[c];n&&n.equals(d)||(e.push(d),n=d)}}return this.autoClose&&e.length>1&&!e[e.length-1].equals(e[0])&&e.push(e[0]),e}copy(t){super.copy(t),this.curves=[];for(let e=0,n=t.curves.length;e<n;e++){const i=t.curves[e];this.curves.push(i.clone())}return this.autoClose=t.autoClose,this}toJSON(){const t=super.toJSON();t.autoClose=this.autoClose,t.curves=[];for(let e=0,n=this.curves.length;e<n;e++){const i=this.curves[e];t.curves.push(i.toJSON())}return t}fromJSON(t){super.fromJSON(t),this.autoClose=t.autoClose,this.curves=[];for(let e=0,n=t.curves.length;e<n;e++){const i=t.curves[e];this.curves.push(new po[i.type]().fromJSON(i))}return this}}class Kc extends Jf{constructor(t){super(),this.type="Path",this.currentPoint=new At,t&&this.setFromPoints(t)}setFromPoints(t){this.moveTo(t[0].x,t[0].y);for(let e=1,n=t.length;e<n;e++)this.lineTo(t[e].x,t[e].y);return this}moveTo(t,e){return this.currentPoint.set(t,e),this}lineTo(t,e){const n=new Mu(this.currentPoint.clone(),new At(t,e));return this.curves.push(n),this.currentPoint.set(t,e),this}quadraticCurveTo(t,e,n,i){const s=new Su(this.currentPoint.clone(),new At(t,e),new At(n,i));return this.curves.push(s),this.currentPoint.set(n,i),this}bezierCurveTo(t,e,n,i,s,r){const a=new yu(this.currentPoint.clone(),new At(t,e),new At(n,i),new At(s,r));return this.curves.push(a),this.currentPoint.set(s,r),this}splineThru(t){const e=[this.currentPoint.clone()].concat(t),n=new Tu(e);return this.curves.push(n),this.currentPoint.copy(t[t.length-1]),this}arc(t,e,n,i,s,r){const a=this.currentPoint.x,l=this.currentPoint.y;return this.absarc(t+a,e+l,n,i,s,r),this}absarc(t,e,n,i,s,r){return this.absellipse(t,e,n,n,i,s,r),this}ellipse(t,e,n,i,s,r,a,l){const c=this.currentPoint.x,d=this.currentPoint.y;return this.absellipse(t+c,e+d,n,i,s,r,a,l),this}absellipse(t,e,n,i,s,r,a,l){const c=new qc(t,e,n,i,s,r,a,l);if(this.curves.length>0){const m=c.getPoint(0);m.equals(this.currentPoint)||this.lineTo(m.x,m.y)}this.curves.push(c);const d=c.getPoint(1);return this.currentPoint.copy(d),this}copy(t){return super.copy(t),this.currentPoint.copy(t.currentPoint),this}toJSON(){const t=super.toJSON();return t.currentPoint=this.currentPoint.toArray(),t}fromJSON(t){return super.fromJSON(t),this.currentPoint.fromArray(t.currentPoint),this}}class ks extends Kc{constructor(t){super(t),this.uuid=An(),this.type="Shape",this.holes=[]}getPointsHoles(t){const e=[];for(let n=0,i=this.holes.length;n<i;n++)e[n]=this.holes[n].getPoints(t);return e}extractPoints(t){return{shape:this.getPoints(t),holes:this.getPointsHoles(t)}}copy(t){super.copy(t),this.holes=[];for(let e=0,n=t.holes.length;e<n;e++){const i=t.holes[e];this.holes.push(i.clone())}return this}toJSON(){const t=super.toJSON();t.uuid=this.uuid,t.holes=[];for(let e=0,n=this.holes.length;e<n;e++){const i=this.holes[e];t.holes.push(i.toJSON())}return t}fromJSON(t){super.fromJSON(t),this.uuid=t.uuid,this.holes=[];for(let e=0,n=t.holes.length;e<n;e++){const i=t.holes[e];this.holes.push(new Kc().fromJSON(i))}return this}}function Kf(u,t,e=2){const n=t&&t.length,i=n?t[0]*e:u.length;let s=Au(u,0,i,e,!0);const r=[];if(!s||s.next===s.prev)return r;let a,l,c;if(n&&(s=nd(u,t,s,e)),u.length>80*e){a=u[0],l=u[1];let d=a,m=l;for(let g=e;g<i;g+=e){const _=u[g],v=u[g+1];_<a&&(a=_),v<l&&(l=v),_>d&&(d=_),v>m&&(m=v)}c=Math.max(d-a,m-l),c=c!==0?32767/c:0}return Vr(s,r,e,a,l,c,0),r}function Au(u,t,e,n,i){let s;if(i===dd(u,t,e,n)>0)for(let r=t;r<e;r+=n)s=Ru(r/n|0,u[r],u[r+1],s);else for(let r=e-n;r>=t;r-=n)s=Ru(r/n|0,u[r],u[r+1],s);return s&&Gs(s,s.next)&&(Hr(s),s=s.next),s}function ms(u,t){if(!u)return u;t||(t=u);let e=u,n;do if(n=!1,!e.steiner&&(Gs(e,e.next)||je(e.prev,e,e.next)===0)){if(Hr(e),e=t=e.prev,e===e.next)break;n=!0}else e=e.next;while(n||e!==t);return t}function Vr(u,t,e,n,i,s,r){if(!u)return;!r&&s&&od(u,n,i,s);let a=u;for(;u.prev!==u.next;){const l=u.prev,c=u.next;if(s?jf(u,n,i,s):Qf(u)){t.push(l.i,u.i,c.i),Hr(u),u=c.next,a=c.next;continue}if(u=c,u===a){r?r===1?(u=td(ms(u),t),Vr(u,t,e,n,i,s,2)):r===2&&ed(u,t,e,n,i,s):Vr(ms(u),t,e,n,i,s,1);break}}}function Qf(u){const t=u.prev,e=u,n=u.next;if(je(t,e,n)>=0)return!1;const i=t.x,s=e.x,r=n.x,a=t.y,l=e.y,c=n.y,d=Math.min(i,s,r),m=Math.min(a,l,c),g=Math.max(i,s,r),_=Math.max(a,l,c);let v=n.next;for(;v!==t;){if(v.x>=d&&v.x<=g&&v.y>=m&&v.y<=_&&kr(i,a,s,l,r,c,v.x,v.y)&&je(v.prev,v,v.next)>=0)return!1;v=v.next}return!0}function jf(u,t,e,n){const i=u.prev,s=u,r=u.next;if(je(i,s,r)>=0)return!1;const a=i.x,l=s.x,c=r.x,d=i.y,m=s.y,g=r.y,_=Math.min(a,l,c),v=Math.min(d,m,g),M=Math.max(a,l,c),C=Math.max(d,m,g),w=Qc(_,v,t,e,n),L=Qc(M,C,t,e,n);let D=u.prevZ,N=u.nextZ;for(;D&&D.z>=w&&N&&N.z<=L;){if(D.x>=_&&D.x<=M&&D.y>=v&&D.y<=C&&D!==i&&D!==r&&kr(a,d,l,m,c,g,D.x,D.y)&&je(D.prev,D,D.next)>=0||(D=D.prevZ,N.x>=_&&N.x<=M&&N.y>=v&&N.y<=C&&N!==i&&N!==r&&kr(a,d,l,m,c,g,N.x,N.y)&&je(N.prev,N,N.next)>=0))return!1;N=N.nextZ}for(;D&&D.z>=w;){if(D.x>=_&&D.x<=M&&D.y>=v&&D.y<=C&&D!==i&&D!==r&&kr(a,d,l,m,c,g,D.x,D.y)&&je(D.prev,D,D.next)>=0)return!1;D=D.prevZ}for(;N&&N.z<=L;){if(N.x>=_&&N.x<=M&&N.y>=v&&N.y<=C&&N!==i&&N!==r&&kr(a,d,l,m,c,g,N.x,N.y)&&je(N.prev,N,N.next)>=0)return!1;N=N.nextZ}return!0}function td(u,t){let e=u;do{const n=e.prev,i=e.next.next;!Gs(n,i)&&wu(n,e,e.next,i)&&Gr(n,i)&&Gr(i,n)&&(t.push(n.i,e.i,i.i),Hr(e),Hr(e.next),e=u=i),e=e.next}while(e!==u);return ms(e)}function ed(u,t,e,n,i,s){let r=u;do{let a=r.next.next;for(;a!==r.prev;){if(r.i!==a.i&&hd(r,a)){let l=Cu(r,a);r=ms(r,r.next),l=ms(l,l.next),Vr(r,t,e,n,i,s,0),Vr(l,t,e,n,i,s,0);return}a=a.next}r=r.next}while(r!==u)}function nd(u,t,e,n){const i=[];for(let s=0,r=t.length;s<r;s++){const a=t[s]*n,l=s<r-1?t[s+1]*n:u.length,c=Au(u,a,l,n,!1);c===c.next&&(c.steiner=!0),i.push(cd(c))}i.sort(id);for(let s=0;s<i.length;s++)e=sd(i[s],e);return e}function id(u,t){let e=u.x-t.x;if(e===0&&(e=u.y-t.y,e===0)){const n=(u.next.y-u.y)/(u.next.x-u.x),i=(t.next.y-t.y)/(t.next.x-t.x);e=n-i}return e}function sd(u,t){const e=rd(u,t);if(!e)return t;const n=Cu(e,u);return ms(n,n.next),ms(e,e.next)}function rd(u,t){let e=t;const n=u.x,i=u.y;let s=-1/0,r;if(Gs(u,e))return e;do{if(Gs(u,e.next))return e.next;if(i<=e.y&&i>=e.next.y&&e.next.y!==e.y){const m=e.x+(i-e.y)*(e.next.x-e.x)/(e.next.y-e.y);if(m<=n&&m>s&&(s=m,r=e.x<e.next.x?e:e.next,m===n))return r}e=e.next}while(e!==t);if(!r)return null;const a=r,l=r.x,c=r.y;let d=1/0;e=r;do{if(n>=e.x&&e.x>=l&&n!==e.x&&Eu(i<c?n:s,i,l,c,i<c?s:n,i,e.x,e.y)){const m=Math.abs(i-e.y)/(n-e.x);Gr(e,u)&&(m<d||m===d&&(e.x>r.x||e.x===r.x&&ad(r,e)))&&(r=e,d=m)}e=e.next}while(e!==a);return r}function ad(u,t){return je(u.prev,u,t.prev)<0&&je(t.next,u,u.next)<0}function od(u,t,e,n){let i=u;do i.z===0&&(i.z=Qc(i.x,i.y,t,e,n)),i.prevZ=i.prev,i.nextZ=i.next,i=i.next;while(i!==u);i.prevZ.nextZ=null,i.prevZ=null,ld(i)}function ld(u){let t,e=1;do{let n=u,i;u=null;let s=null;for(t=0;n;){t++;let r=n,a=0;for(let c=0;c<e&&(a++,r=r.nextZ,!!r);c++);let l=e;for(;a>0||l>0&&r;)a!==0&&(l===0||!r||n.z<=r.z)?(i=n,n=n.nextZ,a--):(i=r,r=r.nextZ,l--),s?s.nextZ=i:u=i,i.prevZ=s,s=i;n=r}s.nextZ=null,e*=2}while(t>1);return u}function Qc(u,t,e,n,i){return u=(u-e)*i|0,t=(t-n)*i|0,u=(u|u<<8)&16711935,u=(u|u<<4)&252645135,u=(u|u<<2)&858993459,u=(u|u<<1)&1431655765,t=(t|t<<8)&16711935,t=(t|t<<4)&252645135,t=(t|t<<2)&858993459,t=(t|t<<1)&1431655765,u|t<<1}function cd(u){let t=u,e=u;do(t.x<e.x||t.x===e.x&&t.y<e.y)&&(e=t),t=t.next;while(t!==u);return e}function Eu(u,t,e,n,i,s,r,a){return(i-r)*(t-a)>=(u-r)*(s-a)&&(u-r)*(n-a)>=(e-r)*(t-a)&&(e-r)*(s-a)>=(i-r)*(n-a)}function kr(u,t,e,n,i,s,r,a){return!(u===r&&t===a)&&Eu(u,t,e,n,i,s,r,a)}function hd(u,t){return u.next.i!==t.i&&u.prev.i!==t.i&&!ud(u,t)&&(Gr(u,t)&&Gr(t,u)&&fd(u,t)&&(je(u.prev,u,t.prev)||je(u,t.prev,t))||Gs(u,t)&&je(u.prev,u,u.next)>0&&je(t.prev,t,t.next)>0)}function je(u,t,e){return(t.y-u.y)*(e.x-t.x)-(t.x-u.x)*(e.y-t.y)}function Gs(u,t){return u.x===t.x&&u.y===t.y}function wu(u,t,e,n){const i=go(je(u,t,e)),s=go(je(u,t,n)),r=go(je(e,n,u)),a=go(je(e,n,t));return!!(i!==s&&r!==a||i===0&&mo(u,e,t)||s===0&&mo(u,n,t)||r===0&&mo(e,u,n)||a===0&&mo(e,t,n))}function mo(u,t,e){return t.x<=Math.max(u.x,e.x)&&t.x>=Math.min(u.x,e.x)&&t.y<=Math.max(u.y,e.y)&&t.y>=Math.min(u.y,e.y)}function go(u){return u>0?1:u<0?-1:0}function ud(u,t){let e=u;do{if(e.i!==u.i&&e.next.i!==u.i&&e.i!==t.i&&e.next.i!==t.i&&wu(e,e.next,u,t))return!0;e=e.next}while(e!==u);return!1}function Gr(u,t){return je(u.prev,u,u.next)<0?je(u,t,u.next)>=0&&je(u,u.prev,t)>=0:je(u,t,u.prev)<0||je(u,u.next,t)<0}function fd(u,t){let e=u,n=!1;const i=(u.x+t.x)/2,s=(u.y+t.y)/2;do e.y>s!=e.next.y>s&&e.next.y!==e.y&&i<(e.next.x-e.x)*(s-e.y)/(e.next.y-e.y)+e.x&&(n=!n),e=e.next;while(e!==u);return n}function Cu(u,t){const e=jc(u.i,u.x,u.y),n=jc(t.i,t.x,t.y),i=u.next,s=t.prev;return u.next=t,t.prev=u,e.next=i,i.prev=e,n.next=e,e.prev=n,s.next=n,n.prev=s,n}function Ru(u,t,e,n){const i=jc(u,t,e);return n?(i.next=n.next,i.prev=n,n.next.prev=i,n.next=i):(i.prev=i,i.next=i),i}function Hr(u){u.next.prev=u.prev,u.prev.next=u.next,u.prevZ&&(u.prevZ.nextZ=u.nextZ),u.nextZ&&(u.nextZ.prevZ=u.prevZ)}function jc(u,t,e){return{i:u,x:t,y:e,prev:null,next:null,z:0,prevZ:null,nextZ:null,steiner:!1}}function dd(u,t,e,n){let i=0;for(let s=t,r=e-n;s<e;s+=n)i+=(u[r]-u[s])*(u[s+1]+u[r+1]),r=s;return i}class pd{static triangulate(t,e,n=2){return Kf(t,e,n)}}class pi{static area(t){const e=t.length;let n=0;for(let i=e-1,s=0;s<e;i=s++)n+=t[i].x*t[s].y-t[s].x*t[i].y;return n*.5}static isClockWise(t){return pi.area(t)<0}static triangulateShape(t,e){const n=[],i=[],s=[];Iu(t),Pu(n,t);let r=t.length;e.forEach(Iu);for(let l=0;l<e.length;l++)i.push(r),r+=e[l].length,Pu(n,e[l]);const a=pd.triangulate(n,i);for(let l=0;l<a.length;l+=3)s.push(a.slice(l,l+3));return s}}function Iu(u){const t=u.length;t>2&&u[t-1].equals(u[0])&&u.pop()}function Pu(u,t){for(let e=0;e<t.length;e++)u.push(t[e].x),u.push(t[e].y)}class th extends y{constructor(t=new ks([new At(.5,.5),new At(-.5,.5),new At(-.5,-.5),new At(.5,-.5)]),e={}){super(),this.type="ExtrudeGeometry",this.parameters={shapes:t,options:e},t=Array.isArray(t)?t:[t];const n=this,i=[],s=[];for(let a=0,l=t.length;a<l;a++){const c=t[a];r(c)}this.setAttribute("position",new Ut(i,3)),this.setAttribute("uv",new Ut(s,2)),this.computeVertexNormals();function r(a){const l=[],c=e.curveSegments!==void 0?e.curveSegments:12,d=e.steps!==void 0?e.steps:1,m=e.depth!==void 0?e.depth:1;let g=e.bevelEnabled!==void 0?e.bevelEnabled:!0,_=e.bevelThickness!==void 0?e.bevelThickness:.2,v=e.bevelSize!==void 0?e.bevelSize:_-.1,M=e.bevelOffset!==void 0?e.bevelOffset:0,C=e.bevelSegments!==void 0?e.bevelSegments:3;const w=e.extrudePath,L=e.UVGenerator!==void 0?e.UVGenerator:md;let D,N=!1,et,rt,vt,at;if(w){D=w.getSpacedPoints(d),N=!0,g=!1;const Bt=w.isCatmullRomCurve3?w.closed:!1;et=w.computeFrenetFrames(d,Bt),rt=new O,vt=new O,at=new O}g||(C=0,_=0,v=0,M=0);const Mt=a.extractPoints(c);let St=Mt.shape;const Gt=Mt.holes;if(!pi.isClockWise(St)){St=St.reverse();for(let Bt=0,Kt=Gt.length;Bt<Kt;Bt++){const Qt=Gt[Bt];pi.isClockWise(Qt)&&(Gt[Bt]=Qt.reverse())}}function ge(Bt){const Qt=10000000000000001e-36;let fe=Bt[0];for(let oe=1;oe<=Bt.length;oe++){const Oe=oe%Bt.length,we=Bt[Oe],qe=we.x-fe.x,tn=we.y-fe.y,pn=qe*qe+tn*tn,qn=Math.max(Math.abs(we.x),Math.abs(we.y),Math.abs(fe.x),Math.abs(fe.y)),Ts=Qt*qn*qn;if(pn<=Ts){Bt.splice(Oe,1),oe--;continue}fe=we}}ge(St),Gt.forEach(ge);const Ne=Gt.length,xe=St;for(let Bt=0;Bt<Ne;Bt++){const Kt=Gt[Bt];St=St.concat(Kt)}function nn(Bt,Kt,Qt){return Kt||Re("ExtrudeGeometry: vec does not exist"),Bt.clone().addScaledVector(Kt,Qt)}const cn=St.length;function ii(Bt,Kt,Qt){let fe,oe,Oe;const we=Bt.x-Kt.x,qe=Bt.y-Kt.y,tn=Qt.x-Bt.x,pn=Qt.y-Bt.y,qn=we*we+qe*qe,Ts=we*pn-qe*tn;if(Math.abs(Ts)>Number.EPSILON){const Yn=Math.sqrt(qn),xf=Math.sqrt(tn*tn+pn*pn),vf=Kt.x-qe/Yn,yf=Kt.y+we/Yn,Pp=Qt.x-pn/xf,Lp=Qt.y+tn/xf,Mf=((Pp-vf)*pn-(Lp-yf)*tn)/(we*pn-qe*tn);fe=vf+we*Mf-Bt.x,oe=yf+qe*Mf-Bt.y;const Sf=fe*fe+oe*oe;if(Sf<=2)return new At(fe,oe);Oe=Math.sqrt(Sf/2)}else{let Yn=!1;we>Number.EPSILON?tn>Number.EPSILON&&(Yn=!0):we<-Number.EPSILON?tn<-Number.EPSILON&&(Yn=!0):Math.sign(qe)===Math.sign(pn)&&(Yn=!0),Yn?(fe=-qe,oe=we,Oe=Math.sqrt(qn)):(fe=we,oe=qe,Oe=Math.sqrt(qn/2))}return new At(fe/Oe,oe/Oe)}const sn=[];for(let Bt=0,Kt=xe.length,Qt=Kt-1,fe=Bt+1;Bt<Kt;Bt++,Qt++,fe++)Qt===Kt&&(Qt=0),fe===Kt&&(fe=0),sn[Bt]=ii(xe[Bt],xe[Qt],xe[fe]);const hn=[];let rn,si=sn.concat();for(let Bt=0,Kt=Ne;Bt<Kt;Bt++){const Qt=Gt[Bt];rn=[];for(let fe=0,oe=Qt.length,Oe=oe-1,we=fe+1;fe<oe;fe++,Oe++,we++)Oe===oe&&(Oe=0),we===oe&&(we=0),rn[fe]=ii(Qt[fe],Qt[Oe],Qt[we]);hn.push(rn),si=si.concat(rn)}let $n;if(C===0)$n=pi.triangulateShape(xe,Gt);else{const Bt=[],Kt=[];for(let Qt=0;Qt<C;Qt++){const fe=Qt/C,oe=_*Math.cos(fe*Math.PI/2),Oe=v*Math.sin(fe*Math.PI/2)+M;for(let we=0,qe=xe.length;we<qe;we++){const tn=nn(xe[we],sn[we],Oe);Ai(tn.x,tn.y,-oe),fe===0&&Bt.push(tn)}for(let we=0,qe=Ne;we<qe;we++){const tn=Gt[we];rn=hn[we];const pn=[];for(let qn=0,Ts=tn.length;qn<Ts;qn++){const Yn=nn(tn[qn],rn[qn],Oe);Ai(Yn.x,Yn.y,-oe),fe===0&&pn.push(Yn)}fe===0&&Kt.push(pn)}}$n=pi.triangulateShape(Bt,Kt)}const Js=$n.length,Xr=v+M;for(let Bt=0;Bt<cn;Bt++){const Kt=g?nn(St[Bt],si[Bt],Xr):St[Bt];N?(vt.copy(et.normals[0]).multiplyScalar(Kt.x),rt.copy(et.binormals[0]).multiplyScalar(Kt.y),at.copy(D[0]).add(vt).add(rt),Ai(at.x,at.y,at.z)):Ai(Kt.x,Kt.y,0)}for(let Bt=1;Bt<=d;Bt++)for(let Kt=0;Kt<cn;Kt++){const Qt=g?nn(St[Kt],si[Kt],Xr):St[Kt];N?(vt.copy(et.normals[Bt]).multiplyScalar(Qt.x),rt.copy(et.binormals[Bt]).multiplyScalar(Qt.y),at.copy(D[Bt]).add(vt).add(rt),Ai(at.x,at.y,at.z)):Ai(Qt.x,Qt.y,m/d*Bt)}for(let Bt=C-1;Bt>=0;Bt--){const Kt=Bt/C,Qt=_*Math.cos(Kt*Math.PI/2),fe=v*Math.sin(Kt*Math.PI/2)+M;for(let oe=0,Oe=xe.length;oe<Oe;oe++){const we=nn(xe[oe],sn[oe],fe);Ai(we.x,we.y,m+Qt)}for(let oe=0,Oe=Gt.length;oe<Oe;oe++){const we=Gt[oe];rn=hn[oe];for(let qe=0,tn=we.length;qe<tn;qe++){const pn=nn(we[qe],rn[qe],fe);N?Ai(pn.x,pn.y+D[d-1].y,D[d-1].x+Qt):Ai(pn.x,pn.y,m+Qt)}}}Cp(),Rp();function Cp(){const Bt=i.length/3;if(g){let Kt=0,Qt=cn*Kt;for(let fe=0;fe<Js;fe++){const oe=$n[fe];No(oe[2]+Qt,oe[1]+Qt,oe[0]+Qt)}Kt=d+C*2,Qt=cn*Kt;for(let fe=0;fe<Js;fe++){const oe=$n[fe];No(oe[0]+Qt,oe[1]+Qt,oe[2]+Qt)}}else{for(let Kt=0;Kt<Js;Kt++){const Qt=$n[Kt];No(Qt[2],Qt[1],Qt[0])}for(let Kt=0;Kt<Js;Kt++){const Qt=$n[Kt];No(Qt[0]+cn*d,Qt[1]+cn*d,Qt[2]+cn*d)}}n.addGroup(Bt,i.length/3-Bt,0)}function Rp(){const Bt=i.length/3;let Kt=0;_f(xe,Kt),Kt+=xe.length;for(let Qt=0,fe=Gt.length;Qt<fe;Qt++){const oe=Gt[Qt];_f(oe,Kt),Kt+=oe.length}n.addGroup(Bt,i.length/3-Bt,1)}function _f(Bt,Kt){let Qt=Bt.length;for(;--Qt>=0;){const fe=Qt;let oe=Qt-1;oe<0&&(oe=Bt.length-1);for(let Oe=0,we=d+C*2;Oe<we;Oe++){const qe=cn*Oe,tn=cn*(Oe+1),pn=Kt+fe+qe,qn=Kt+oe+qe,Ts=Kt+oe+tn,Yn=Kt+fe+tn;Ip(pn,qn,Ts,Yn)}}}function Ai(Bt,Kt,Qt){l.push(Bt),l.push(Kt),l.push(Qt)}function No(Bt,Kt,Qt){Ei(Bt),Ei(Kt),Ei(Qt);const fe=i.length/3,oe=L.generateTopUV(n,i,fe-3,fe-2,fe-1);wi(oe[0]),wi(oe[1]),wi(oe[2])}function Ip(Bt,Kt,Qt,fe){Ei(Bt),Ei(Kt),Ei(fe),Ei(Kt),Ei(Qt),Ei(fe);const oe=i.length/3,Oe=L.generateSideWallUV(n,i,oe-6,oe-3,oe-2,oe-1);wi(Oe[0]),wi(Oe[1]),wi(Oe[3]),wi(Oe[1]),wi(Oe[2]),wi(Oe[3])}function Ei(Bt){i.push(l[Bt*3+0]),i.push(l[Bt*3+1]),i.push(l[Bt*3+2])}function wi(Bt){s.push(Bt.x),s.push(Bt.y)}}}copy(t){return super.copy(t),this.parameters=Object.assign({},t.parameters),this}toJSON(){const t=super.toJSON(),e=this.parameters.shapes,n=this.parameters.options;return gd(e,n,t)}static fromJSON(t,e){const n=[];for(let s=0,r=t.shapes.length;s<r;s++){const a=e[t.shapes[s]];n.push(a)}const i=t.options.extrudePath;return i!==void 0&&(t.options.extrudePath=new po[i.type]().fromJSON(i)),new th(n,t.options)}}const md={generateTopUV:function(u,t,e,n,i){const s=t[e*3],r=t[e*3+1],a=t[n*3],l=t[n*3+1],c=t[i*3],d=t[i*3+1];return[new At(s,r),new At(a,l),new At(c,d)]},generateSideWallUV:function(u,t,e,n,i,s){const r=t[e*3],a=t[e*3+1],l=t[e*3+2],c=t[n*3],d=t[n*3+1],m=t[n*3+2],g=t[i*3],_=t[i*3+1],v=t[i*3+2],M=t[s*3],C=t[s*3+1],w=t[s*3+2];return Math.abs(a-d)<Math.abs(r-c)?[new At(r,1-l),new At(c,1-m),new At(g,1-v),new At(M,1-w)]:[new At(a,1-l),new At(d,1-m),new At(_,1-v),new At(C,1-w)]}};function gd(u,t,e){if(e.shapes=[],Array.isArray(u))for(let n=0,i=u.length;n<i;n++){const s=u[n];e.shapes.push(s.uuid)}else e.shapes.push(u.uuid);return e.options=Object.assign({},t),t.extrudePath!==void 0&&(e.options.extrudePath=t.extrudePath.toJSON()),e}class eh extends ps{constructor(t=1,e=0){const n=(1+Math.sqrt(5))/2,i=[-1,n,0,1,n,0,-1,-n,0,1,-n,0,0,-1,n,0,1,n,0,-1,-n,0,1,-n,n,0,-1,n,0,1,-n,0,-1,-n,0,1],s=[0,11,5,0,5,1,0,1,7,0,7,10,0,10,11,1,5,9,5,11,4,11,10,2,10,7,6,7,1,8,3,9,4,3,4,2,3,2,6,3,6,8,3,8,9,4,9,5,2,4,11,6,2,10,8,6,7,9,8,1];super(i,s,t,e),this.type="IcosahedronGeometry",this.parameters={radius:t,detail:e}}static fromJSON(t){return new eh(t.radius,t.detail)}}class nh extends y{constructor(t=[new At(0,-.5),new At(.5,0),new At(0,.5)],e=12,n=0,i=Math.PI*2){super(),this.type="LatheGeometry",this.parameters={points:t,segments:e,phiStart:n,phiLength:i},e=Math.floor(e),i=pe(i,0,Math.PI*2);const s=[],r=[],a=[],l=[],c=[],d=1/e,m=new O,g=new At,_=new O,v=new O,M=new O;let C=0,w=0;for(let L=0;L<=t.length-1;L++)switch(L){case 0:C=t[L+1].x-t[L].x,w=t[L+1].y-t[L].y,_.x=w*1,_.y=-C,_.z=w*0,M.copy(_),_.normalize(),l.push(_.x,_.y,_.z);break;case t.length-1:l.push(M.x,M.y,M.z);break;default:C=t[L+1].x-t[L].x,w=t[L+1].y-t[L].y,_.x=w*1,_.y=-C,_.z=w*0,v.copy(_),_.x+=M.x,_.y+=M.y,_.z+=M.z,_.normalize(),l.push(_.x,_.y,_.z),M.copy(v)}for(let L=0;L<=e;L++){const D=n+L*d*i,N=Math.sin(D),et=Math.cos(D);for(let rt=0;rt<=t.length-1;rt++){m.x=t[rt].x*N,m.y=t[rt].y,m.z=t[rt].x*et,r.push(m.x,m.y,m.z),g.x=L/e,g.y=rt/(t.length-1),a.push(g.x,g.y);const vt=l[3*rt+0]*N,at=l[3*rt+1],Mt=l[3*rt+0]*et;c.push(vt,at,Mt)}}for(let L=0;L<e;L++)for(let D=0;D<t.length-1;D++){const N=D+L*t.length,et=N,rt=N+t.length,vt=N+t.length+1,at=N+1;s.push(et,rt,at),s.push(vt,at,rt)}this.setIndex(s),this.setAttribute("position",new Ut(r,3)),this.setAttribute("uv",new Ut(a,2)),this.setAttribute("normal",new Ut(c,3))}copy(t){return super.copy(t),this.parameters=Object.assign({},t.parameters),this}static fromJSON(t){return new nh(t.points,t.segments,t.phiStart,t.phiLength)}}class _o extends ps{constructor(t=1,e=0){const n=[1,0,0,-1,0,0,0,1,0,0,-1,0,0,0,1,0,0,-1],i=[0,2,4,0,4,3,0,3,5,0,5,2,1,2,5,1,5,3,1,3,4,1,4,2];super(n,i,t,e),this.type="OctahedronGeometry",this.parameters={radius:t,detail:e}}static fromJSON(t){return new _o(t.radius,t.detail)}}class xo extends y{constructor(t=1,e=1,n=1,i=1){super(),this.type="PlaneGeometry",this.parameters={width:t,height:e,widthSegments:n,heightSegments:i};const s=t/2,r=e/2,a=Math.floor(n),l=Math.floor(i),c=a+1,d=l+1,m=t/a,g=e/l,_=[],v=[],M=[],C=[];for(let w=0;w<d;w++){const L=w*g-r;for(let D=0;D<c;D++){const N=D*m-s;v.push(N,-L,0),M.push(0,0,1),C.push(D/a),C.push(1-w/l)}}for(let w=0;w<l;w++)for(let L=0;L<a;L++){const D=L+c*w,N=L+c*(w+1),et=L+1+c*(w+1),rt=L+1+c*w;_.push(D,N,rt),_.push(N,et,rt)}this.setIndex(_),this.setAttribute("position",new Ut(v,3)),this.setAttribute("normal",new Ut(M,3)),this.setAttribute("uv",new Ut(C,2))}copy(t){return super.copy(t),this.parameters=Object.assign({},t.parameters),this}static fromJSON(t){return new xo(t.width,t.height,t.widthSegments,t.heightSegments)}}class ih extends y{constructor(t=.5,e=1,n=32,i=1,s=0,r=Math.PI*2){super(),this.type="RingGeometry",this.parameters={innerRadius:t,outerRadius:e,thetaSegments:n,phiSegments:i,thetaStart:s,thetaLength:r},n=Math.max(3,n),i=Math.max(1,i);const a=[],l=[],c=[],d=[];let m=t;const g=(e-t)/i,_=new O,v=new At;for(let M=0;M<=i;M++){for(let C=0;C<=n;C++){const w=s+C/n*r;_.x=m*Math.cos(w),_.y=m*Math.sin(w),l.push(_.x,_.y,_.z),c.push(0,0,1),v.x=(_.x/e+1)/2,v.y=(_.y/e+1)/2,d.push(v.x,v.y)}m+=g}for(let M=0;M<i;M++){const C=M*(n+1);for(let w=0;w<n;w++){const L=w+C,D=L,N=L+n+1,et=L+n+2,rt=L+1;a.push(D,N,rt),a.push(N,et,rt)}}this.setIndex(a),this.setAttribute("position",new Ut(l,3)),this.setAttribute("normal",new Ut(c,3)),this.setAttribute("uv",new Ut(d,2))}copy(t){return super.copy(t),this.parameters=Object.assign({},t.parameters),this}static fromJSON(t){return new ih(t.innerRadius,t.outerRadius,t.thetaSegments,t.phiSegments,t.thetaStart,t.thetaLength)}}class sh extends y{constructor(t=new ks([new At(0,.5),new At(-.5,-.5),new At(.5,-.5)]),e=12){super(),this.type="ShapeGeometry",this.parameters={shapes:t,curveSegments:e};const n=[],i=[],s=[],r=[];let a=0,l=0;if(Array.isArray(t)===!1)c(t);else for(let d=0;d<t.length;d++)c(t[d]),this.addGroup(a,l,d),a+=l,l=0;this.setIndex(n),this.setAttribute("position",new Ut(i,3)),this.setAttribute("normal",new Ut(s,3)),this.setAttribute("uv",new Ut(r,2));function c(d){const m=i.length/3,g=d.extractPoints(e);let _=g.shape;const v=g.holes;pi.isClockWise(_)===!1&&(_=_.reverse());for(let C=0,w=v.length;C<w;C++){const L=v[C];pi.isClockWise(L)===!0&&(v[C]=L.reverse())}const M=pi.triangulateShape(_,v);for(let C=0,w=v.length;C<w;C++){const L=v[C];_=_.concat(L)}for(let C=0,w=_.length;C<w;C++){const L=_[C];i.push(L.x,L.y,0),s.push(0,0,1),r.push(L.x,L.y)}for(let C=0,w=M.length;C<w;C++){const L=M[C],D=L[0]+m,N=L[1]+m,et=L[2]+m;n.push(D,N,et),l+=3}}}copy(t){return super.copy(t),this.parameters=Object.assign({},t.parameters),this}toJSON(){const t=super.toJSON(),e=this.parameters.shapes;return _d(e,t)}static fromJSON(t,e){const n=[];for(let i=0,s=t.shapes.length;i<s;i++){const r=e[t.shapes[i]];n.push(r)}return new sh(n,t.curveSegments)}}function _d(u,t){if(t.shapes=[],Array.isArray(u))for(let e=0,n=u.length;e<n;e++){const i=u[e];t.shapes.push(i.uuid)}else t.shapes.push(u.uuid);return t}class vo extends y{constructor(t=1,e=32,n=16,i=0,s=Math.PI*2,r=0,a=Math.PI){super(),this.type="SphereGeometry",this.parameters={radius:t,widthSegments:e,heightSegments:n,phiStart:i,phiLength:s,thetaStart:r,thetaLength:a},e=Math.max(3,Math.floor(e)),n=Math.max(2,Math.floor(n));const l=Math.min(r+a,Math.PI);let c=0;const d=[],m=new O,g=new O,_=[],v=[],M=[],C=[];for(let w=0;w<=n;w++){const L=[],D=w/n;let N=0;w===0&&r===0?N=.5/e:w===n&&l===Math.PI&&(N=-.5/e);for(let et=0;et<=e;et++){const rt=et/e;m.x=-t*Math.cos(i+rt*s)*Math.sin(r+D*a),m.y=t*Math.cos(r+D*a),m.z=t*Math.sin(i+rt*s)*Math.sin(r+D*a),v.push(m.x,m.y,m.z),g.copy(m).normalize(),M.push(g.x,g.y,g.z),C.push(rt+N,1-D),L.push(c++)}d.push(L)}for(let w=0;w<n;w++)for(let L=0;L<e;L++){const D=d[w][L+1],N=d[w][L],et=d[w+1][L],rt=d[w+1][L+1];(w!==0||r>0)&&_.push(D,N,rt),(w!==n-1||l<Math.PI)&&_.push(N,et,rt)}this.setIndex(_),this.setAttribute("position",new Ut(v,3)),this.setAttribute("normal",new Ut(M,3)),this.setAttribute("uv",new Ut(C,2))}copy(t){return super.copy(t),this.parameters=Object.assign({},t.parameters),this}static fromJSON(t){return new vo(t.radius,t.widthSegments,t.heightSegments,t.phiStart,t.phiLength,t.thetaStart,t.thetaLength)}}class rh extends ps{constructor(t=1,e=0){const n=[1,1,1,-1,-1,1,-1,1,-1,1,-1,-1],i=[2,1,0,0,3,2,1,3,0,2,3,1];super(n,i,t,e),this.type="TetrahedronGeometry",this.parameters={radius:t,detail:e}}static fromJSON(t){return new rh(t.radius,t.detail)}}class ah extends y{constructor(t=1,e=.4,n=12,i=48,s=Math.PI*2,r=0,a=Math.PI*2){super(),this.type="TorusGeometry",this.parameters={radius:t,tube:e,radialSegments:n,tubularSegments:i,arc:s,thetaStart:r,thetaLength:a},n=Math.floor(n),i=Math.floor(i);const l=[],c=[],d=[],m=[],g=new O,_=new O,v=new O;for(let M=0;M<=n;M++){const C=r+M/n*a;for(let w=0;w<=i;w++){const L=w/i*s;_.x=(t+e*Math.cos(C))*Math.cos(L),_.y=(t+e*Math.cos(C))*Math.sin(L),_.z=e*Math.sin(C),c.push(_.x,_.y,_.z),g.x=t*Math.cos(L),g.y=t*Math.sin(L),v.subVectors(_,g).normalize(),d.push(v.x,v.y,v.z),m.push(w/i),m.push(M/n)}}for(let M=1;M<=n;M++)for(let C=1;C<=i;C++){const w=(i+1)*M+C-1,L=(i+1)*(M-1)+C-1,D=(i+1)*(M-1)+C,N=(i+1)*M+C;l.push(w,L,N),l.push(L,D,N)}this.setIndex(l),this.setAttribute("position",new Ut(c,3)),this.setAttribute("normal",new Ut(d,3)),this.setAttribute("uv",new Ut(m,2))}copy(t){return super.copy(t),this.parameters=Object.assign({},t.parameters),this}static fromJSON(t){return new ah(t.radius,t.tube,t.radialSegments,t.tubularSegments,t.arc)}}class oh extends y{constructor(t=1,e=.4,n=64,i=8,s=2,r=3){super(),this.type="TorusKnotGeometry",this.parameters={radius:t,tube:e,tubularSegments:n,radialSegments:i,p:s,q:r},n=Math.floor(n),i=Math.floor(i);const a=[],l=[],c=[],d=[],m=new O,g=new O,_=new O,v=new O,M=new O,C=new O,w=new O;for(let D=0;D<=n;++D){const N=D/n*s*Math.PI*2;L(N,s,r,t,_),L(N+.01,s,r,t,v),C.subVectors(v,_),w.addVectors(v,_),M.crossVectors(C,w),w.crossVectors(M,C),M.normalize(),w.normalize();for(let et=0;et<=i;++et){const rt=et/i*Math.PI*2,vt=-e*Math.cos(rt),at=e*Math.sin(rt);m.x=_.x+(vt*w.x+at*M.x),m.y=_.y+(vt*w.y+at*M.y),m.z=_.z+(vt*w.z+at*M.z),l.push(m.x,m.y,m.z),g.subVectors(m,_).normalize(),c.push(g.x,g.y,g.z),d.push(D/n),d.push(et/i)}}for(let D=1;D<=n;D++)for(let N=1;N<=i;N++){const et=(i+1)*(D-1)+(N-1),rt=(i+1)*D+(N-1),vt=(i+1)*D+N,at=(i+1)*(D-1)+N;a.push(et,rt,at),a.push(rt,vt,at)}this.setIndex(a),this.setAttribute("position",new Ut(l,3)),this.setAttribute("normal",new Ut(c,3)),this.setAttribute("uv",new Ut(d,2));function L(D,N,et,rt,vt){const at=Math.cos(D),Mt=Math.sin(D),St=et/N*D,Gt=Math.cos(St);vt.x=rt*(2+Gt)*.5*at,vt.y=rt*(2+Gt)*Mt*.5,vt.z=rt*Math.sin(St)*.5}}copy(t){return super.copy(t),this.parameters=Object.assign({},t.parameters),this}static fromJSON(t){return new oh(t.radius,t.tube,t.tubularSegments,t.radialSegments,t.p,t.q)}}class lh extends y{constructor(t=new bu(new O(-1,-1,0),new O(-1,1,0),new O(1,1,0)),e=64,n=1,i=8,s=!1){super(),this.type="TubeGeometry",this.parameters={path:t,tubularSegments:e,radius:n,radialSegments:i,closed:s};const r=t.computeFrenetFrames(e,s);this.tangents=r.tangents,this.normals=r.normals,this.binormals=r.binormals;const a=new O,l=new O,c=new At;let d=new O;const m=[],g=[],_=[],v=[];M(),this.setIndex(v),this.setAttribute("position",new Ut(m,3)),this.setAttribute("normal",new Ut(g,3)),this.setAttribute("uv",new Ut(_,2));function M(){for(let D=0;D<e;D++)C(D);C(s===!1?e:0),L(),w()}function C(D){d=t.getPointAt(D/e,d);const N=r.normals[D],et=r.binormals[D];for(let rt=0;rt<=i;rt++){const vt=rt/i*Math.PI*2,at=Math.sin(vt),Mt=-Math.cos(vt);l.x=Mt*N.x+at*et.x,l.y=Mt*N.y+at*et.y,l.z=Mt*N.z+at*et.z,l.normalize(),g.push(l.x,l.y,l.z),a.x=d.x+n*l.x,a.y=d.y+n*l.y,a.z=d.z+n*l.z,m.push(a.x,a.y,a.z)}}function w(){for(let D=1;D<=e;D++)for(let N=1;N<=i;N++){const et=(i+1)*(D-1)+(N-1),rt=(i+1)*D+(N-1),vt=(i+1)*D+N,at=(i+1)*(D-1)+N;v.push(et,rt,at),v.push(rt,vt,at)}}function L(){for(let D=0;D<=e;D++)for(let N=0;N<=i;N++)c.x=D/e,c.y=N/i,_.push(c.x,c.y)}}copy(t){return super.copy(t),this.parameters=Object.assign({},t.parameters),this}toJSON(){const t=super.toJSON();return t.path=this.parameters.path.toJSON(),t}static fromJSON(t){return new lh(new po[t.path.type]().fromJSON(t.path),t.tubularSegments,t.radius,t.radialSegments,t.closed)}}class xd extends y{constructor(t=null){if(super(),this.type="WireframeGeometry",this.parameters={geometry:t},t!==null){const e=[],n=new Set,i=new O,s=new O;if(t.index!==null){const r=t.attributes.position,a=t.index;let l=t.groups;l.length===0&&(l=[{start:0,count:a.count,materialIndex:0}]);for(let c=0,d=l.length;c<d;++c){const m=l[c],g=m.start,_=m.count;for(let v=g,M=g+_;v<M;v+=3)for(let C=0;C<3;C++){const w=a.getX(v+C),L=a.getX(v+(C+1)%3);i.fromBufferAttribute(r,w),s.fromBufferAttribute(r,L),Lu(i,s,n)===!0&&(e.push(i.x,i.y,i.z),e.push(s.x,s.y,s.z))}}}else{const r=t.attributes.position;for(let a=0,l=r.count/3;a<l;a++)for(let c=0;c<3;c++){const d=3*a+c,m=3*a+(c+1)%3;i.fromBufferAttribute(r,d),s.fromBufferAttribute(r,m),Lu(i,s,n)===!0&&(e.push(i.x,i.y,i.z),e.push(s.x,s.y,s.z))}}this.setAttribute("position",new Ut(e,3))}}copy(t){return super.copy(t),this.parameters=Object.assign({},t.parameters),this}}function Lu(u,t,e){const n=`${u.x},${u.y},${u.z}-${t.x},${t.y},${t.z}`,i=`${t.x},${t.y},${t.z}-${u.x},${u.y},${u.z}`;return e.has(n)===!0||e.has(i)===!0?!1:(e.add(n),e.add(i),!0)}var Du=Object.freeze({__proto__:null,BoxGeometry:ao,CapsuleGeometry:Gc,CircleGeometry:Hc,ConeGeometry:lo,CylinderGeometry:oo,DodecahedronGeometry:Wc,EdgesGeometry:Bf,ExtrudeGeometry:th,IcosahedronGeometry:eh,LatheGeometry:nh,OctahedronGeometry:_o,PlaneGeometry:xo,PolyhedronGeometry:ps,RingGeometry:ih,ShapeGeometry:sh,SphereGeometry:vo,TetrahedronGeometry:rh,TorusGeometry:ah,TorusKnotGeometry:oh,TubeGeometry:lh,WireframeGeometry:xd});class vd extends Ft{constructor(t){super(),this.isShadowMaterial=!0,this.type="ShadowMaterial",this.color=new re(0),this.transparent=!0,this.fog=!0,this.setValues(t)}copy(t){return super.copy(t),this.color.copy(t.color),this.fog=t.fog,this}}function yo(u){const t={};for(const e in u){t[e]={};for(const n in u[e]){const i=u[e][n];i&&(i.isColor||i.isMatrix3||i.isMatrix4||i.isVector2||i.isVector3||i.isVector4||i.isTexture||i.isQuaternion)?i.isRenderTargetTexture?(le("UniformsUtils: Textures of render targets cannot be cloned via cloneUniforms() or mergeUniforms()."),t[e][n]=null):t[e][n]=i.clone():Array.isArray(i)?t[e][n]=i.slice():t[e][n]=i}}return t}function Uu(u){const t={};for(let e=0;e<u.length;e++){const n=yo(u[e]);for(const i in n)t[i]=n[i]}return t}function yd(u){const t=[];for(let e=0;e<u.length;e++)t.push(u[e].clone());return t}function Md(u){const t=u.getRenderTarget();return t===null?u.outputColorSpace:t.isXRRenderTarget===!0?t.texture.colorSpace:_n.workingColorSpace}const Sd={clone:yo,merge:Uu};var bd=`void main() {
	gl_Position = projectionMatrix * modelViewMatrix * vec4( position, 1.0 );
}`,Td=`void main() {
	gl_FragColor = vec4( 1.0, 0.0, 0.0, 1.0 );
}`;class ch extends Ft{constructor(t){super(),this.isShaderMaterial=!0,this.type="ShaderMaterial",this.defines={},this.uniforms={},this.uniformsGroups=[],this.vertexShader=bd,this.fragmentShader=Td,this.linewidth=1,this.wireframe=!1,this.wireframeLinewidth=1,this.fog=!1,this.lights=!1,this.clipping=!1,this.forceSinglePass=!0,this.extensions={clipCullDistance:!1,multiDraw:!1},this.defaultAttributeValues={color:[1,1,1],uv:[0,0],uv1:[0,0]},this.index0AttributeName=void 0,this.uniformsNeedUpdate=!1,this.glslVersion=null,t!==void 0&&this.setValues(t)}copy(t){return super.copy(t),this.fragmentShader=t.fragmentShader,this.vertexShader=t.vertexShader,this.uniforms=yo(t.uniforms),this.uniformsGroups=yd(t.uniformsGroups),this.defines=Object.assign({},t.defines),this.wireframe=t.wireframe,this.wireframeLinewidth=t.wireframeLinewidth,this.fog=t.fog,this.lights=t.lights,this.clipping=t.clipping,this.extensions=Object.assign({},t.extensions),this.glslVersion=t.glslVersion,this.defaultAttributeValues=Object.assign({},t.defaultAttributeValues),this.index0AttributeName=t.index0AttributeName,this.uniformsNeedUpdate=t.uniformsNeedUpdate,this}toJSON(t){const e=super.toJSON(t);e.glslVersion=this.glslVersion,e.uniforms={};for(const i in this.uniforms){const r=this.uniforms[i].value;r&&r.isTexture?e.uniforms[i]={type:"t",value:r.toJSON(t).uuid}:r&&r.isColor?e.uniforms[i]={type:"c",value:r.getHex()}:r&&r.isVector2?e.uniforms[i]={type:"v2",value:r.toArray()}:r&&r.isVector3?e.uniforms[i]={type:"v3",value:r.toArray()}:r&&r.isVector4?e.uniforms[i]={type:"v4",value:r.toArray()}:r&&r.isMatrix3?e.uniforms[i]={type:"m3",value:r.toArray()}:r&&r.isMatrix4?e.uniforms[i]={type:"m4",value:r.toArray()}:e.uniforms[i]={value:r}}Object.keys(this.defines).length>0&&(e.defines=this.defines),e.vertexShader=this.vertexShader,e.fragmentShader=this.fragmentShader,e.lights=this.lights,e.clipping=this.clipping;const n={};for(const i in this.extensions)this.extensions[i]===!0&&(n[i]=!0);return Object.keys(n).length>0&&(e.extensions=n),e}}class Nu extends ch{constructor(t){super(t),this.isRawShaderMaterial=!0,this.type="RawShaderMaterial"}}class Fu extends Ft{constructor(t){super(),this.isMeshStandardMaterial=!0,this.type="MeshStandardMaterial",this.defines={STANDARD:""},this.color=new re(16777215),this.roughness=1,this.metalness=0,this.map=null,this.lightMap=null,this.lightMapIntensity=1,this.aoMap=null,this.aoMapIntensity=1,this.emissive=new re(0),this.emissiveIntensity=1,this.emissiveMap=null,this.bumpMap=null,this.bumpScale=1,this.normalMap=null,this.normalMapType=Un,this.normalScale=new At(1,1),this.displacementMap=null,this.displacementScale=1,this.displacementBias=0,this.roughnessMap=null,this.metalnessMap=null,this.alphaMap=null,this.envMap=null,this.envMapRotation=new Bn,this.envMapIntensity=1,this.wireframe=!1,this.wireframeLinewidth=1,this.wireframeLinecap="round",this.wireframeLinejoin="round",this.flatShading=!1,this.fog=!0,this.setValues(t)}copy(t){return super.copy(t),this.defines={STANDARD:""},this.color.copy(t.color),this.roughness=t.roughness,this.metalness=t.metalness,this.map=t.map,this.lightMap=t.lightMap,this.lightMapIntensity=t.lightMapIntensity,this.aoMap=t.aoMap,this.aoMapIntensity=t.aoMapIntensity,this.emissive.copy(t.emissive),this.emissiveMap=t.emissiveMap,this.emissiveIntensity=t.emissiveIntensity,this.bumpMap=t.bumpMap,this.bumpScale=t.bumpScale,this.normalMap=t.normalMap,this.normalMapType=t.normalMapType,this.normalScale.copy(t.normalScale),this.displacementMap=t.displacementMap,this.displacementScale=t.displacementScale,this.displacementBias=t.displacementBias,this.roughnessMap=t.roughnessMap,this.metalnessMap=t.metalnessMap,this.alphaMap=t.alphaMap,this.envMap=t.envMap,this.envMapRotation.copy(t.envMapRotation),this.envMapIntensity=t.envMapIntensity,this.wireframe=t.wireframe,this.wireframeLinewidth=t.wireframeLinewidth,this.wireframeLinecap=t.wireframeLinecap,this.wireframeLinejoin=t.wireframeLinejoin,this.flatShading=t.flatShading,this.fog=t.fog,this}}class Ad extends Fu{constructor(t){super(),this.isMeshPhysicalMaterial=!0,this.defines={STANDARD:"",PHYSICAL:""},this.type="MeshPhysicalMaterial",this.anisotropyRotation=0,this.anisotropyMap=null,this.clearcoatMap=null,this.clearcoatRoughness=0,this.clearcoatRoughnessMap=null,this.clearcoatNormalScale=new At(1,1),this.clearcoatNormalMap=null,this.ior=1.5,Object.defineProperty(this,"reflectivity",{get:function(){return pe(2.5*(this.ior-1)/(this.ior+1),0,1)},set:function(e){this.ior=(1+.4*e)/(1-.4*e)}}),this.iridescenceMap=null,this.iridescenceIOR=1.3,this.iridescenceThicknessRange=[100,400],this.iridescenceThicknessMap=null,this.sheenColor=new re(0),this.sheenColorMap=null,this.sheenRoughness=1,this.sheenRoughnessMap=null,this.transmissionMap=null,this.thickness=0,this.thicknessMap=null,this.attenuationDistance=1/0,this.attenuationColor=new re(1,1,1),this.specularIntensity=1,this.specularIntensityMap=null,this.specularColor=new re(1,1,1),this.specularColorMap=null,this._anisotropy=0,this._clearcoat=0,this._dispersion=0,this._iridescence=0,this._sheen=0,this._transmission=0,this.setValues(t)}get anisotropy(){return this._anisotropy}set anisotropy(t){this._anisotropy>0!=t>0&&this.version++,this._anisotropy=t}get clearcoat(){return this._clearcoat}set clearcoat(t){this._clearcoat>0!=t>0&&this.version++,this._clearcoat=t}get iridescence(){return this._iridescence}set iridescence(t){this._iridescence>0!=t>0&&this.version++,this._iridescence=t}get dispersion(){return this._dispersion}set dispersion(t){this._dispersion>0!=t>0&&this.version++,this._dispersion=t}get sheen(){return this._sheen}set sheen(t){this._sheen>0!=t>0&&this.version++,this._sheen=t}get transmission(){return this._transmission}set transmission(t){this._transmission>0!=t>0&&this.version++,this._transmission=t}copy(t){return super.copy(t),this.defines={STANDARD:"",PHYSICAL:""},this.anisotropy=t.anisotropy,this.anisotropyRotation=t.anisotropyRotation,this.anisotropyMap=t.anisotropyMap,this.clearcoat=t.clearcoat,this.clearcoatMap=t.clearcoatMap,this.clearcoatRoughness=t.clearcoatRoughness,this.clearcoatRoughnessMap=t.clearcoatRoughnessMap,this.clearcoatNormalMap=t.clearcoatNormalMap,this.clearcoatNormalScale.copy(t.clearcoatNormalScale),this.dispersion=t.dispersion,this.ior=t.ior,this.iridescence=t.iridescence,this.iridescenceMap=t.iridescenceMap,this.iridescenceIOR=t.iridescenceIOR,this.iridescenceThicknessRange=[...t.iridescenceThicknessRange],this.iridescenceThicknessMap=t.iridescenceThicknessMap,this.sheen=t.sheen,this.sheenColor.copy(t.sheenColor),this.sheenColorMap=t.sheenColorMap,this.sheenRoughness=t.sheenRoughness,this.sheenRoughnessMap=t.sheenRoughnessMap,this.transmission=t.transmission,this.transmissionMap=t.transmissionMap,this.thickness=t.thickness,this.thicknessMap=t.thicknessMap,this.attenuationDistance=t.attenuationDistance,this.attenuationColor.copy(t.attenuationColor),this.specularIntensity=t.specularIntensity,this.specularIntensityMap=t.specularIntensityMap,this.specularColor.copy(t.specularColor),this.specularColorMap=t.specularColorMap,this}}class Ed extends Ft{constructor(t){super(),this.isMeshPhongMaterial=!0,this.type="MeshPhongMaterial",this.color=new re(16777215),this.specular=new re(1118481),this.shininess=30,this.map=null,this.lightMap=null,this.lightMapIntensity=1,this.aoMap=null,this.aoMapIntensity=1,this.emissive=new re(0),this.emissiveIntensity=1,this.emissiveMap=null,this.bumpMap=null,this.bumpScale=1,this.normalMap=null,this.normalMapType=Un,this.normalScale=new At(1,1),this.displacementMap=null,this.displacementScale=1,this.displacementBias=0,this.specularMap=null,this.alphaMap=null,this.envMap=null,this.envMapRotation=new Bn,this.combine=Es,this.reflectivity=1,this.envMapIntensity=1,this.refractionRatio=.98,this.wireframe=!1,this.wireframeLinewidth=1,this.wireframeLinecap="round",this.wireframeLinejoin="round",this.flatShading=!1,this.fog=!0,this.setValues(t)}copy(t){return super.copy(t),this.color.copy(t.color),this.specular.copy(t.specular),this.shininess=t.shininess,this.map=t.map,this.lightMap=t.lightMap,this.lightMapIntensity=t.lightMapIntensity,this.aoMap=t.aoMap,this.aoMapIntensity=t.aoMapIntensity,this.emissive.copy(t.emissive),this.emissiveMap=t.emissiveMap,this.emissiveIntensity=t.emissiveIntensity,this.bumpMap=t.bumpMap,this.bumpScale=t.bumpScale,this.normalMap=t.normalMap,this.normalMapType=t.normalMapType,this.normalScale.copy(t.normalScale),this.displacementMap=t.displacementMap,this.displacementScale=t.displacementScale,this.displacementBias=t.displacementBias,this.specularMap=t.specularMap,this.alphaMap=t.alphaMap,this.envMap=t.envMap,this.envMapRotation.copy(t.envMapRotation),this.combine=t.combine,this.reflectivity=t.reflectivity,this.envMapIntensity=t.envMapIntensity,this.refractionRatio=t.refractionRatio,this.wireframe=t.wireframe,this.wireframeLinewidth=t.wireframeLinewidth,this.wireframeLinecap=t.wireframeLinecap,this.wireframeLinejoin=t.wireframeLinejoin,this.flatShading=t.flatShading,this.fog=t.fog,this}}class wd extends Ft{constructor(t){super(),this.isMeshToonMaterial=!0,this.defines={TOON:""},this.type="MeshToonMaterial",this.color=new re(16777215),this.map=null,this.gradientMap=null,this.lightMap=null,this.lightMapIntensity=1,this.aoMap=null,this.aoMapIntensity=1,this.emissive=new re(0),this.emissiveIntensity=1,this.emissiveMap=null,this.bumpMap=null,this.bumpScale=1,this.normalMap=null,this.normalMapType=Un,this.normalScale=new At(1,1),this.displacementMap=null,this.displacementScale=1,this.displacementBias=0,this.alphaMap=null,this.wireframe=!1,this.wireframeLinewidth=1,this.wireframeLinecap="round",this.wireframeLinejoin="round",this.fog=!0,this.setValues(t)}copy(t){return super.copy(t),this.color.copy(t.color),this.map=t.map,this.gradientMap=t.gradientMap,this.lightMap=t.lightMap,this.lightMapIntensity=t.lightMapIntensity,this.aoMap=t.aoMap,this.aoMapIntensity=t.aoMapIntensity,this.emissive.copy(t.emissive),this.emissiveMap=t.emissiveMap,this.emissiveIntensity=t.emissiveIntensity,this.bumpMap=t.bumpMap,this.bumpScale=t.bumpScale,this.normalMap=t.normalMap,this.normalMapType=t.normalMapType,this.normalScale.copy(t.normalScale),this.displacementMap=t.displacementMap,this.displacementScale=t.displacementScale,this.displacementBias=t.displacementBias,this.alphaMap=t.alphaMap,this.wireframe=t.wireframe,this.wireframeLinewidth=t.wireframeLinewidth,this.wireframeLinecap=t.wireframeLinecap,this.wireframeLinejoin=t.wireframeLinejoin,this.fog=t.fog,this}}class Cd extends Ft{constructor(t){super(),this.isMeshNormalMaterial=!0,this.type="MeshNormalMaterial",this.bumpMap=null,this.bumpScale=1,this.normalMap=null,this.normalMapType=Un,this.normalScale=new At(1,1),this.displacementMap=null,this.displacementScale=1,this.displacementBias=0,this.wireframe=!1,this.wireframeLinewidth=1,this.flatShading=!1,this.setValues(t)}copy(t){return super.copy(t),this.bumpMap=t.bumpMap,this.bumpScale=t.bumpScale,this.normalMap=t.normalMap,this.normalMapType=t.normalMapType,this.normalScale.copy(t.normalScale),this.displacementMap=t.displacementMap,this.displacementScale=t.displacementScale,this.displacementBias=t.displacementBias,this.wireframe=t.wireframe,this.wireframeLinewidth=t.wireframeLinewidth,this.flatShading=t.flatShading,this}}class Rd extends Ft{constructor(t){super(),this.isMeshLambertMaterial=!0,this.type="MeshLambertMaterial",this.color=new re(16777215),this.map=null,this.lightMap=null,this.lightMapIntensity=1,this.aoMap=null,this.aoMapIntensity=1,this.emissive=new re(0),this.emissiveIntensity=1,this.emissiveMap=null,this.bumpMap=null,this.bumpScale=1,this.normalMap=null,this.normalMapType=Un,this.normalScale=new At(1,1),this.displacementMap=null,this.displacementScale=1,this.displacementBias=0,this.specularMap=null,this.alphaMap=null,this.envMap=null,this.envMapRotation=new Bn,this.combine=Es,this.reflectivity=1,this.envMapIntensity=1,this.refractionRatio=.98,this.wireframe=!1,this.wireframeLinewidth=1,this.wireframeLinecap="round",this.wireframeLinejoin="round",this.flatShading=!1,this.fog=!0,this.setValues(t)}copy(t){return super.copy(t),this.color.copy(t.color),this.map=t.map,this.lightMap=t.lightMap,this.lightMapIntensity=t.lightMapIntensity,this.aoMap=t.aoMap,this.aoMapIntensity=t.aoMapIntensity,this.emissive.copy(t.emissive),this.emissiveMap=t.emissiveMap,this.emissiveIntensity=t.emissiveIntensity,this.bumpMap=t.bumpMap,this.bumpScale=t.bumpScale,this.normalMap=t.normalMap,this.normalMapType=t.normalMapType,this.normalScale.copy(t.normalScale),this.displacementMap=t.displacementMap,this.displacementScale=t.displacementScale,this.displacementBias=t.displacementBias,this.specularMap=t.specularMap,this.alphaMap=t.alphaMap,this.envMap=t.envMap,this.envMapRotation.copy(t.envMapRotation),this.combine=t.combine,this.reflectivity=t.reflectivity,this.envMapIntensity=t.envMapIntensity,this.refractionRatio=t.refractionRatio,this.wireframe=t.wireframe,this.wireframeLinewidth=t.wireframeLinewidth,this.wireframeLinecap=t.wireframeLinecap,this.wireframeLinejoin=t.wireframeLinejoin,this.flatShading=t.flatShading,this.fog=t.fog,this}}class Ou extends Ft{constructor(t){super(),this.isMeshDepthMaterial=!0,this.type="MeshDepthMaterial",this.depthPacking=ri,this.map=null,this.alphaMap=null,this.displacementMap=null,this.displacementScale=1,this.displacementBias=0,this.wireframe=!1,this.wireframeLinewidth=1,this.setValues(t)}copy(t){return super.copy(t),this.depthPacking=t.depthPacking,this.map=t.map,this.alphaMap=t.alphaMap,this.displacementMap=t.displacementMap,this.displacementScale=t.displacementScale,this.displacementBias=t.displacementBias,this.wireframe=t.wireframe,this.wireframeLinewidth=t.wireframeLinewidth,this}}class Bu extends Ft{constructor(t){super(),this.isMeshDistanceMaterial=!0,this.type="MeshDistanceMaterial",this.map=null,this.alphaMap=null,this.displacementMap=null,this.displacementScale=1,this.displacementBias=0,this.setValues(t)}copy(t){return super.copy(t),this.map=t.map,this.alphaMap=t.alphaMap,this.displacementMap=t.displacementMap,this.displacementScale=t.displacementScale,this.displacementBias=t.displacementBias,this}}class Id extends Ft{constructor(t){super(),this.isMeshMatcapMaterial=!0,this.defines={MATCAP:""},this.type="MeshMatcapMaterial",this.color=new re(16777215),this.matcap=null,this.map=null,this.bumpMap=null,this.bumpScale=1,this.normalMap=null,this.normalMapType=Un,this.normalScale=new At(1,1),this.displacementMap=null,this.displacementScale=1,this.displacementBias=0,this.alphaMap=null,this.wireframe=!1,this.wireframeLinewidth=1,this.flatShading=!1,this.fog=!0,this.setValues(t)}copy(t){return super.copy(t),this.defines={MATCAP:""},this.color.copy(t.color),this.matcap=t.matcap,this.map=t.map,this.bumpMap=t.bumpMap,this.bumpScale=t.bumpScale,this.normalMap=t.normalMap,this.normalMapType=t.normalMapType,this.normalScale.copy(t.normalScale),this.displacementMap=t.displacementMap,this.displacementScale=t.displacementScale,this.displacementBias=t.displacementBias,this.alphaMap=t.alphaMap,this.wireframe=t.wireframe,this.wireframeLinewidth=t.wireframeLinewidth,this.flatShading=t.flatShading,this.fog=t.fog,this}}class Pd extends Pn{constructor(t){super(),this.isLineDashedMaterial=!0,this.type="LineDashedMaterial",this.scale=1,this.dashSize=3,this.gapSize=1,this.setValues(t)}copy(t){return super.copy(t),this.scale=t.scale,this.dashSize=t.dashSize,this.gapSize=t.gapSize,this}}function gs(u,t){return!u||u.constructor===t?u:typeof t.BYTES_PER_ELEMENT=="number"?new t(u):Array.prototype.slice.call(u)}function zu(u){function t(i,s){return u[i]-u[s]}const e=u.length,n=new Array(e);for(let i=0;i!==e;++i)n[i]=i;return n.sort(t),n}function hh(u,t,e){const n=u.length,i=new u.constructor(n);for(let s=0,r=0;r!==n;++s){const a=e[s]*t;for(let l=0;l!==t;++l)i[r++]=u[a+l]}return i}function uh(u,t,e,n){let i=1,s=u[0];for(;s!==void 0&&s[n]===void 0;)s=u[i++];if(s===void 0)return;let r=s[n];if(r!==void 0)if(Array.isArray(r))do r=s[n],r!==void 0&&(t.push(s.time),e.push(...r)),s=u[i++];while(s!==void 0);else if(r.toArray!==void 0)do r=s[n],r!==void 0&&(t.push(s.time),r.toArray(e,e.length)),s=u[i++];while(s!==void 0);else do r=s[n],r!==void 0&&(t.push(s.time),e.push(r)),s=u[i++];while(s!==void 0)}function Ld(u,t,e,n,i=30){const s=u.clone();s.name=t;const r=[];for(let l=0;l<s.tracks.length;++l){const c=s.tracks[l],d=c.getValueSize(),m=[],g=[];for(let _=0;_<c.times.length;++_){const v=c.times[_]*i;if(!(v<e||v>=n)){m.push(c.times[_]);for(let M=0;M<d;++M)g.push(c.values[_*d+M])}}m.length!==0&&(c.times=gs(m,c.times.constructor),c.values=gs(g,c.values.constructor),r.push(c))}s.tracks=r;let a=1/0;for(let l=0;l<s.tracks.length;++l)a>s.tracks[l].times[0]&&(a=s.tracks[l].times[0]);for(let l=0;l<s.tracks.length;++l)s.tracks[l].shift(-1*a);return s.resetDuration(),s}function Dd(u,t=0,e=u,n=30){n<=0&&(n=30);const i=e.tracks.length,s=t/n;for(let r=0;r<i;++r){const a=e.tracks[r],l=a.ValueTypeName;if(l==="bool"||l==="string")continue;const c=u.tracks.find(function(w){return w.name===a.name&&w.ValueTypeName===l});if(c===void 0)continue;let d=0;const m=a.getValueSize();a.createInterpolant.isInterpolantFactoryMethodGLTFCubicSpline&&(d=m/3);let g=0;const _=c.getValueSize();c.createInterpolant.isInterpolantFactoryMethodGLTFCubicSpline&&(g=_/3);const v=a.times.length-1;let M;if(s<=a.times[0]){const w=d,L=m-d;M=a.values.slice(w,L)}else if(s>=a.times[v]){const w=v*m+d,L=w+m-d;M=a.values.slice(w,L)}else{const w=a.createInterpolant(),L=d,D=m-d;w.evaluate(s),M=w.resultBuffer.slice(L,D)}l==="quaternion"&&new un().fromArray(M).normalize().conjugate().toArray(M);const C=c.times.length;for(let w=0;w<C;++w){const L=w*_+g;if(l==="quaternion")un.multiplyQuaternionsFlat(c.values,L,M,0,c.values,L);else{const D=_-g*2;for(let N=0;N<D;++N)c.values[L+N]-=M[N]}}}return u.blendMode=ba,u}class Bp{static convertArray(t,e){return gs(t,e)}static isTypedArray(t){return Da(t)}static getKeyframeOrder(t){return zu(t)}static sortedArray(t,e,n){return hh(t,e,n)}static flattenJSON(t,e,n,i){uh(t,e,n,i)}static subclip(t,e,n,i,s=30){return Ld(t,e,n,i,s)}static makeClipAdditive(t,e=0,n=t,i=30){return Dd(t,e,n,i)}}class Wr{constructor(t,e,n,i){this.parameterPositions=t,this._cachedIndex=0,this.resultBuffer=i!==void 0?i:new e.constructor(n),this.sampleValues=e,this.valueSize=n,this.settings=null,this.DefaultSettings_={}}evaluate(t){const e=this.parameterPositions;let n=this._cachedIndex,i=e[n],s=e[n-1];t:{e:{let r;n:{i:if(!(t<i)){for(let a=n+2;;){if(i===void 0){if(t<s)break i;return n=e.length,this._cachedIndex=n,this.copySampleValue_(n-1)}if(n===a)break;if(s=i,i=e[++n],t<i)break e}r=e.length;break n}if(!(t>=s)){const a=e[1];t<a&&(n=2,s=a);for(let l=n-2;;){if(s===void 0)return this._cachedIndex=0,this.copySampleValue_(0);if(n===l)break;if(i=s,s=e[--n-1],t>=s)break e}r=n,n=0;break n}break t}for(;n<r;){const a=n+r>>>1;t<e[a]?r=a:n=a+1}if(i=e[n],s=e[n-1],s===void 0)return this._cachedIndex=0,this.copySampleValue_(0);if(i===void 0)return n=e.length,this._cachedIndex=n,this.copySampleValue_(n-1)}this._cachedIndex=n,this.intervalChanged_(n,s,i)}return this.interpolate_(n,s,t,i)}getSettings_(){return this.settings||this.DefaultSettings_}copySampleValue_(t){const e=this.resultBuffer,n=this.sampleValues,i=this.valueSize,s=t*i;for(let r=0;r!==i;++r)e[r]=n[s+r];return e}interpolate_(){throw new Error("call to abstract method")}intervalChanged_(){}}class Ud extends Wr{constructor(t,e,n,i){super(t,e,n,i),this._weightPrev=-0,this._offsetPrev=-0,this._weightNext=-0,this._offsetNext=-0,this.DefaultSettings_={endingStart:Jn,endingEnd:Jn}}intervalChanged_(t,e,n){const i=this.parameterPositions;let s=t-2,r=t+1,a=i[s],l=i[r];if(a===void 0)switch(this.getSettings_().endingStart){case Rn:s=t,a=2*e-n;break;case Ls:s=i.length-2,a=e+i[s]-i[s+1];break;default:s=t,a=n}if(l===void 0)switch(this.getSettings_().endingEnd){case Rn:r=t,l=2*n-e;break;case Ls:r=1,l=n+i[1]-i[0];break;default:r=t-1,l=e}const c=(n-e)*.5,d=this.valueSize;this._weightPrev=c/(e-a),this._weightNext=c/(l-n),this._offsetPrev=s*d,this._offsetNext=r*d}interpolate_(t,e,n,i){const s=this.resultBuffer,r=this.sampleValues,a=this.valueSize,l=t*a,c=l-a,d=this._offsetPrev,m=this._offsetNext,g=this._weightPrev,_=this._weightNext,v=(n-e)/(i-e),M=v*v,C=M*v,w=-g*C+2*g*M-g*v,L=(1+g)*C+(-1.5-2*g)*M+(-.5+g)*v+1,D=(-1-_)*C+(1.5+_)*M+.5*v,N=_*C-_*M;for(let et=0;et!==a;++et)s[et]=w*r[d+et]+L*r[c+et]+D*r[l+et]+N*r[m+et];return s}}class Vu extends Wr{constructor(t,e,n,i){super(t,e,n,i)}interpolate_(t,e,n,i){const s=this.resultBuffer,r=this.sampleValues,a=this.valueSize,l=t*a,c=l-a,d=(n-e)/(i-e),m=1-d;for(let g=0;g!==a;++g)s[g]=r[c+g]*m+r[l+g]*d;return s}}class Nd extends Wr{constructor(t,e,n,i){super(t,e,n,i)}interpolate_(t){return this.copySampleValue_(t-1)}}class Fd extends Wr{interpolate_(t,e,n,i){const s=this.resultBuffer,r=this.sampleValues,a=this.valueSize,l=t*a,c=l-a,d=this.settings||this.DefaultSettings_,m=d.inTangents,g=d.outTangents;if(!m||!g){const M=(n-e)/(i-e),C=1-M;for(let w=0;w!==a;++w)s[w]=r[c+w]*C+r[l+w]*M;return s}const _=a*2,v=t-1;for(let M=0;M!==a;++M){const C=r[c+M],w=r[l+M],L=v*_+M*2,D=g[L],N=g[L+1],et=t*_+M*2,rt=m[et],vt=m[et+1];let at=(n-e)/(i-e),Mt,St,Gt,he,ge;for(let Ne=0;Ne<8;Ne++){Mt=at*at,St=Mt*at,Gt=1-at,he=Gt*Gt,ge=he*Gt;const nn=ge*e+3*he*at*D+3*Gt*Mt*rt+St*i-n;if(Math.abs(nn)<1e-10)break;const cn=3*he*(D-e)+6*Gt*at*(rt-D)+3*Mt*(i-rt);if(Math.abs(cn)<1e-10)break;at=at-nn/cn,at=Math.max(0,Math.min(1,at))}s[M]=ge*C+3*he*at*N+3*Gt*Mt*vt+St*w}return s}}class ni{constructor(t,e,n,i){if(t===void 0)throw new Error("THREE.KeyframeTrack: track name is undefined");if(e===void 0||e.length===0)throw new Error("THREE.KeyframeTrack: no keyframes in track named "+t);this.name=t,this.times=gs(e,this.TimeBufferType),this.values=gs(n,this.ValueBufferType),this.setInterpolation(i||this.DefaultInterpolation)}static toJSON(t){const e=t.constructor;let n;if(e.toJSON!==this.toJSON)n=e.toJSON(t);else{n={name:t.name,times:gs(t.times,Array),values:gs(t.values,Array)};const i=t.getInterpolation();i!==t.DefaultInterpolation&&(n.interpolation=i)}return n.type=t.ValueTypeName,n}InterpolantFactoryMethodDiscrete(t){return new Nd(this.times,this.values,this.getValueSize(),t)}InterpolantFactoryMethodLinear(t){return new Vu(this.times,this.values,this.getValueSize(),t)}InterpolantFactoryMethodSmooth(t){return new Ud(this.times,this.values,this.getValueSize(),t)}InterpolantFactoryMethodBezier(t){const e=new Fd(this.times,this.values,this.getValueSize(),t);return this.settings&&(e.settings=this.settings),e}setInterpolation(t){let e;switch(t){case ur:e=this.InterpolantFactoryMethodDiscrete;break;case ve:e=this.InterpolantFactoryMethodLinear;break;case Nt:e=this.InterpolantFactoryMethodSmooth;break;case Dn:e=this.InterpolantFactoryMethodBezier;break}if(e===void 0){const n="unsupported interpolation for "+this.ValueTypeName+" keyframe track named "+this.name;if(this.createInterpolant===void 0)if(t!==this.DefaultInterpolation)this.setInterpolation(this.DefaultInterpolation);else throw new Error(n);return le("KeyframeTrack:",n),this}return this.createInterpolant=e,this}getInterpolation(){switch(this.createInterpolant){case this.InterpolantFactoryMethodDiscrete:return ur;case this.InterpolantFactoryMethodLinear:return ve;case this.InterpolantFactoryMethodSmooth:return Nt;case this.InterpolantFactoryMethodBezier:return Dn}}getValueSize(){return this.values.length/this.times.length}shift(t){if(t!==0){const e=this.times;for(let n=0,i=e.length;n!==i;++n)e[n]+=t}return this}scale(t){if(t!==1){const e=this.times;for(let n=0,i=e.length;n!==i;++n)e[n]*=t}return this}trim(t,e){const n=this.times,i=n.length;let s=0,r=i-1;for(;s!==i&&n[s]<t;)++s;for(;r!==-1&&n[r]>e;)--r;if(++r,s!==0||r!==i){s>=r&&(r=Math.max(r,1),s=r-1);const a=this.getValueSize();this.times=n.slice(s,r),this.values=this.values.slice(s*a,r*a)}return this}validate(){let t=!0;const e=this.getValueSize();e-Math.floor(e)!==0&&(Re("KeyframeTrack: Invalid value size in track.",this),t=!1);const n=this.times,i=this.values,s=n.length;s===0&&(Re("KeyframeTrack: Track is empty.",this),t=!1);let r=null;for(let a=0;a!==s;a++){const l=n[a];if(typeof l=="number"&&isNaN(l)){Re("KeyframeTrack: Time is not a valid number.",this,a,l),t=!1;break}if(r!==null&&r>l){Re("KeyframeTrack: Out of order keys.",this,a,l,r),t=!1;break}r=l}if(i!==void 0&&Da(i))for(let a=0,l=i.length;a!==l;++a){const c=i[a];if(isNaN(c)){Re("KeyframeTrack: Value is not a valid number.",this,a,c),t=!1;break}}return t}optimize(){const t=this.times.slice(),e=this.values.slice(),n=this.getValueSize(),i=this.getInterpolation()===Nt,s=t.length-1;let r=1;for(let a=1;a<s;++a){let l=!1;const c=t[a],d=t[a+1];if(c!==d&&(a!==1||c!==t[0]))if(i)l=!0;else{const m=a*n,g=m-n,_=m+n;for(let v=0;v!==n;++v){const M=e[m+v];if(M!==e[g+v]||M!==e[_+v]){l=!0;break}}}if(l){if(a!==r){t[r]=t[a];const m=a*n,g=r*n;for(let _=0;_!==n;++_)e[g+_]=e[m+_]}++r}}if(s>0){t[r]=t[s];for(let a=s*n,l=r*n,c=0;c!==n;++c)e[l+c]=e[a+c];++r}return r!==t.length?(this.times=t.slice(0,r),this.values=e.slice(0,r*n)):(this.times=t,this.values=e),this}clone(){const t=this.times.slice(),e=this.values.slice(),n=this.constructor,i=new n(this.name,t,e);return i.createInterpolant=this.createInterpolant,i}}ni.prototype.ValueTypeName="",ni.prototype.TimeBufferType=Float32Array,ni.prototype.ValueBufferType=Float32Array,ni.prototype.DefaultInterpolation=ve;class Hs extends ni{constructor(t,e,n){super(t,e,n)}}Hs.prototype.ValueTypeName="bool",Hs.prototype.ValueBufferType=Array,Hs.prototype.DefaultInterpolation=ur,Hs.prototype.InterpolantFactoryMethodLinear=void 0,Hs.prototype.InterpolantFactoryMethodSmooth=void 0;class ku extends ni{constructor(t,e,n,i){super(t,e,n,i)}}ku.prototype.ValueTypeName="color";class Mo extends ni{constructor(t,e,n,i){super(t,e,n,i)}}Mo.prototype.ValueTypeName="number";class Od extends Wr{constructor(t,e,n,i){super(t,e,n,i)}interpolate_(t,e,n,i){const s=this.resultBuffer,r=this.sampleValues,a=this.valueSize,l=(n-e)/(i-e);let c=t*a;for(let d=c+a;c!==d;c+=4)un.slerpFlat(s,0,r,c-a,r,c,l);return s}}class So extends ni{constructor(t,e,n,i){super(t,e,n,i)}InterpolantFactoryMethodLinear(t){return new Od(this.times,this.values,this.getValueSize(),t)}}So.prototype.ValueTypeName="quaternion",So.prototype.InterpolantFactoryMethodSmooth=void 0;class Ws extends ni{constructor(t,e,n){super(t,e,n)}}Ws.prototype.ValueTypeName="string",Ws.prototype.ValueBufferType=Array,Ws.prototype.DefaultInterpolation=ur,Ws.prototype.InterpolantFactoryMethodLinear=void 0,Ws.prototype.InterpolantFactoryMethodSmooth=void 0;class bo extends ni{constructor(t,e,n,i){super(t,e,n,i)}}bo.prototype.ValueTypeName="vector";class To{constructor(t="",e=-1,n=[],i=fr){this.name=t,this.tracks=n,this.duration=e,this.blendMode=i,this.uuid=An(),this.userData={},this.duration<0&&this.resetDuration()}static parse(t){const e=[],n=t.tracks,i=1/(t.fps||1);for(let r=0,a=n.length;r!==a;++r)e.push(zd(n[r]).scale(i));const s=new this(t.name,t.duration,e,t.blendMode);return s.uuid=t.uuid,s.userData=JSON.parse(t.userData||"{}"),s}static toJSON(t){const e=[],n=t.tracks,i={name:t.name,duration:t.duration,tracks:e,uuid:t.uuid,blendMode:t.blendMode,userData:JSON.stringify(t.userData)};for(let s=0,r=n.length;s!==r;++s)e.push(ni.toJSON(n[s]));return i}static CreateFromMorphTargetSequence(t,e,n,i){const s=e.length,r=[];for(let a=0;a<s;a++){let l=[],c=[];l.push((a+s-1)%s,a,(a+1)%s),c.push(0,1,0);const d=zu(l);l=hh(l,1,d),c=hh(c,1,d),!i&&l[0]===0&&(l.push(s),c.push(c[0])),r.push(new Mo(".morphTargetInfluences["+e[a].name+"]",l,c).scale(1/n))}return new this(t,-1,r)}static findByName(t,e){let n=t;if(!Array.isArray(t)){const i=t;n=i.geometry&&i.geometry.animations||i.animations}for(let i=0;i<n.length;i++)if(n[i].name===e)return n[i];return null}static CreateClipsFromMorphTargetSequences(t,e,n){const i={},s=/^([\w-]*?)([\d]+)$/;for(let a=0,l=t.length;a<l;a++){const c=t[a],d=c.name.match(s);if(d&&d.length>1){const m=d[1];let g=i[m];g||(i[m]=g=[]),g.push(c)}}const r=[];for(const a in i)r.push(this.CreateFromMorphTargetSequence(a,i[a],e,n));return r}static parseAnimation(t,e){if(le("AnimationClip: parseAnimation() is deprecated and will be removed with r185"),!t)return Re("AnimationClip: No animation in JSONLoader data."),null;const n=function(m,g,_,v,M){if(_.length!==0){const C=[],w=[];uh(_,C,w,v),C.length!==0&&M.push(new m(g,C,w))}},i=[],s=t.name||"default",r=t.fps||30,a=t.blendMode;let l=t.length||-1;const c=t.hierarchy||[];for(let m=0;m<c.length;m++){const g=c[m].keys;if(!(!g||g.length===0))if(g[0].morphTargets){const _={};let v;for(v=0;v<g.length;v++)if(g[v].morphTargets)for(let M=0;M<g[v].morphTargets.length;M++)_[g[v].morphTargets[M]]=-1;for(const M in _){const C=[],w=[];for(let L=0;L!==g[v].morphTargets.length;++L){const D=g[v];C.push(D.time),w.push(D.morphTarget===M?1:0)}i.push(new Mo(".morphTargetInfluence["+M+"]",C,w))}l=_.length*r}else{const _=".bones["+e[m].name+"]";n(bo,_+".position",g,"pos",i),n(So,_+".quaternion",g,"rot",i),n(bo,_+".scale",g,"scl",i)}}return i.length===0?null:new this(s,l,i,a)}resetDuration(){const t=this.tracks;let e=0;for(let n=0,i=t.length;n!==i;++n){const s=this.tracks[n];e=Math.max(e,s.times[s.times.length-1])}return this.duration=e,this}trim(){for(let t=0;t<this.tracks.length;t++)this.tracks[t].trim(0,this.duration);return this}validate(){let t=!0;for(let e=0;e<this.tracks.length;e++)t=t&&this.tracks[e].validate();return t}optimize(){for(let t=0;t<this.tracks.length;t++)this.tracks[t].optimize();return this}clone(){const t=[];for(let n=0;n<this.tracks.length;n++)t.push(this.tracks[n].clone());const e=new this.constructor(this.name,this.duration,t,this.blendMode);return e.userData=JSON.parse(JSON.stringify(this.userData)),e}toJSON(){return this.constructor.toJSON(this)}}function Bd(u){switch(u.toLowerCase()){case"scalar":case"double":case"float":case"number":case"integer":return Mo;case"vector":case"vector2":case"vector3":case"vector4":return bo;case"color":return ku;case"quaternion":return So;case"bool":case"boolean":return Hs;case"string":return Ws}throw new Error("THREE.KeyframeTrack: Unsupported typeName: "+u)}function zd(u){if(u.type===void 0)throw new Error("THREE.KeyframeTrack: track type undefined, can not parse");const t=Bd(u.type);if(u.times===void 0){const e=[],n=[];uh(u.keys,e,n,"value"),u.times=e,u.values=n}return t.parse!==void 0?t.parse(u):new t(u.name,u.times,u.values,u.interpolation)}const bi={enabled:!1,files:{},add:function(u,t){this.enabled!==!1&&(Gu(u)||(this.files[u]=t))},get:function(u){if(this.enabled!==!1&&!Gu(u))return this.files[u]},remove:function(u){delete this.files[u]},clear:function(){this.files={}}};function Gu(u){try{const t=u.slice(u.indexOf(":")+1);return new URL(t).protocol==="blob:"}catch{return!1}}class Hu{constructor(t,e,n){const i=this;let s=!1,r=0,a=0,l;const c=[];this.onStart=void 0,this.onLoad=t,this.onProgress=e,this.onError=n,this._abortController=null,this.itemStart=function(d){a++,s===!1&&i.onStart!==void 0&&i.onStart(d,r,a),s=!0},this.itemEnd=function(d){r++,i.onProgress!==void 0&&i.onProgress(d,r,a),r===a&&(s=!1,i.onLoad!==void 0&&i.onLoad())},this.itemError=function(d){i.onError!==void 0&&i.onError(d)},this.resolveURL=function(d){return l?l(d):d},this.setURLModifier=function(d){return l=d,this},this.addHandler=function(d,m){return c.push(d,m),this},this.removeHandler=function(d){const m=c.indexOf(d);return m!==-1&&c.splice(m,2),this},this.getHandler=function(d){for(let m=0,g=c.length;m<g;m+=2){const _=c[m],v=c[m+1];if(_.global&&(_.lastIndex=0),_.test(d))return v}return null},this.abort=function(){return this.abortController.abort(),this._abortController=null,this}}get abortController(){return this._abortController||(this._abortController=new AbortController),this._abortController}}const Vd=new Hu;class Wn{constructor(t){this.manager=t!==void 0?t:Vd,this.crossOrigin="anonymous",this.withCredentials=!1,this.path="",this.resourcePath="",this.requestHeader={},typeof __THREE_DEVTOOLS__<"u"&&__THREE_DEVTOOLS__.dispatchEvent(new CustomEvent("observe",{detail:this}))}load(){}loadAsync(t,e){const n=this;return new Promise(function(i,s){n.load(t,i,e,s)})}parse(){}setCrossOrigin(t){return this.crossOrigin=t,this}setWithCredentials(t){return this.withCredentials=t,this}setPath(t){return this.path=t,this}setResourcePath(t){return this.resourcePath=t,this}setRequestHeader(t){return this.requestHeader=t,this}abort(){return this}}Wn.DEFAULT_MATERIAL_NAME="__DEFAULT";const Ti={};class kd extends Error{constructor(t,e){super(t),this.response=e}}class Gi extends Wn{constructor(t){super(t),this.mimeType="",this.responseType="",this._abortController=new AbortController}load(t,e,n,i){t===void 0&&(t=""),this.path!==void 0&&(t=this.path+t),t=this.manager.resolveURL(t);const s=bi.get(`file:${t}`);if(s!==void 0)return this.manager.itemStart(t),setTimeout(()=>{e&&e(s),this.manager.itemEnd(t)},0),s;if(Ti[t]!==void 0){Ti[t].push({onLoad:e,onProgress:n,onError:i});return}Ti[t]=[],Ti[t].push({onLoad:e,onProgress:n,onError:i});const r=new Request(t,{headers:new Headers(this.requestHeader),credentials:this.withCredentials?"include":"same-origin",signal:typeof AbortSignal.any=="function"?AbortSignal.any([this._abortController.signal,this.manager.abortController.signal]):this._abortController.signal}),a=this.mimeType,l=this.responseType;fetch(r).then(c=>{if(c.status===200||c.status===0){if(c.status===0&&le("FileLoader: HTTP Status 0 received."),typeof ReadableStream>"u"||c.body===void 0||c.body.getReader===void 0)return c;const d=Ti[t],m=c.body.getReader(),g=c.headers.get("X-File-Size")||c.headers.get("Content-Length"),_=g?parseInt(g):0,v=_!==0;let M=0;const C=new ReadableStream({start(w){L();function L(){m.read().then(({done:D,value:N})=>{if(D)w.close();else{M+=N.byteLength;const et=new ProgressEvent("progress",{lengthComputable:v,loaded:M,total:_});for(let rt=0,vt=d.length;rt<vt;rt++){const at=d[rt];at.onProgress&&at.onProgress(et)}w.enqueue(N),L()}},D=>{w.error(D)})}}});return new Response(C)}else throw new kd(`fetch for "${c.url}" responded with ${c.status}: ${c.statusText}`,c)}).then(c=>{switch(l){case"arraybuffer":return c.arrayBuffer();case"blob":return c.blob();case"document":return c.text().then(d=>new DOMParser().parseFromString(d,a));case"json":return c.json();default:if(a==="")return c.text();{const m=/charset="?([^;"\s]*)"?/i.exec(a),g=m&&m[1]?m[1].toLowerCase():void 0,_=new TextDecoder(g);return c.arrayBuffer().then(v=>_.decode(v))}}}).then(c=>{bi.add(`file:${t}`,c);const d=Ti[t];delete Ti[t];for(let m=0,g=d.length;m<g;m++){const _=d[m];_.onLoad&&_.onLoad(c)}}).catch(c=>{const d=Ti[t];if(d===void 0)throw this.manager.itemError(t),c;delete Ti[t];for(let m=0,g=d.length;m<g;m++){const _=d[m];_.onError&&_.onError(c)}this.manager.itemError(t)}).finally(()=>{this.manager.itemEnd(t)}),this.manager.itemStart(t)}setResponseType(t){return this.responseType=t,this}setMimeType(t){return this.mimeType=t,this}abort(){return this._abortController.abort(),this._abortController=new AbortController,this}}class zp extends Wn{constructor(t){super(t)}load(t,e,n,i){const s=this,r=new Gi(this.manager);r.setPath(this.path),r.setRequestHeader(this.requestHeader),r.setWithCredentials(this.withCredentials),r.load(t,function(a){try{e(s.parse(JSON.parse(a)))}catch(l){i?i(l):Re(l),s.manager.itemError(t)}},n,i)}parse(t){const e=[];for(let n=0;n<t.length;n++){const i=To.parse(t[n]);e.push(i)}return e}}class Vp extends Wn{constructor(t){super(t)}load(t,e,n,i){const s=this,r=[],a=new Vc,l=new Gi(this.manager);l.setPath(this.path),l.setResponseType("arraybuffer"),l.setRequestHeader(this.requestHeader),l.setWithCredentials(s.withCredentials);let c=0;function d(m){l.load(t[m],function(g){const _=s.parse(g,!0);r[m]={width:_.width,height:_.height,format:_.format,mipmaps:_.mipmaps},c+=1,c===6&&(_.mipmapCount===1&&(a.minFilter=Cn),a.image=r,a.format=_.format,a.needsUpdate=!0,e&&e(a))},n,i)}if(Array.isArray(t))for(let m=0,g=t.length;m<g;++m)d(m);else l.load(t,function(m){const g=s.parse(m,!0);if(g.isCubemap){const _=g.mipmaps.length/g.mipmapCount;for(let v=0;v<_;v++){r[v]={mipmaps:[]};for(let M=0;M<g.mipmapCount;M++)r[v].mipmaps.push(g.mipmaps[v*g.mipmapCount+M]),r[v].format=g.format,r[v].width=g.width,r[v].height=g.height}a.image=r}else a.image.width=g.width,a.image.height=g.height,a.mipmaps=g.mipmaps;g.mipmapCount===1&&(a.minFilter=Cn),a.format=g.format,a.needsUpdate=!0,e&&e(a)},n,i);return a}}const Xs=new WeakMap;class Ao extends Wn{constructor(t){super(t)}load(t,e,n,i){this.path!==void 0&&(t=this.path+t),t=this.manager.resolveURL(t);const s=this,r=bi.get(`image:${t}`);if(r!==void 0){if(r.complete===!0)s.manager.itemStart(t),setTimeout(function(){e&&e(r),s.manager.itemEnd(t)},0);else{let m=Xs.get(r);m===void 0&&(m=[],Xs.set(r,m)),m.push({onLoad:e,onError:i})}return r}const a=es("img");function l(){d(),e&&e(this);const m=Xs.get(this)||[];for(let g=0;g<m.length;g++){const _=m[g];_.onLoad&&_.onLoad(this)}Xs.delete(this),s.manager.itemEnd(t)}function c(m){d(),i&&i(m),bi.remove(`image:${t}`);const g=Xs.get(this)||[];for(let _=0;_<g.length;_++){const v=g[_];v.onError&&v.onError(m)}Xs.delete(this),s.manager.itemError(t),s.manager.itemEnd(t)}function d(){a.removeEventListener("load",l,!1),a.removeEventListener("error",c,!1)}return a.addEventListener("load",l,!1),a.addEventListener("error",c,!1),t.slice(0,5)!=="data:"&&this.crossOrigin!==void 0&&(a.crossOrigin=this.crossOrigin),bi.add(`image:${t}`,a),s.manager.itemStart(t),a.src=t,a}}class kp extends Wn{constructor(t){super(t)}load(t,e,n,i){const s=new kc;s.colorSpace=yn;const r=new Ao(this.manager);r.setCrossOrigin(this.crossOrigin),r.setPath(this.path);let a=0;function l(c){r.load(t[c],function(d){s.images[c]=d,a++,a===6&&(s.needsUpdate=!0,e&&e(s))},void 0,i)}for(let c=0;c<t.length;++c)l(c);return s}}class Gp extends Wn{constructor(t){super(t)}load(t,e,n,i){const s=this,r=new Yt,a=new Gi(this.manager);return a.setResponseType("arraybuffer"),a.setRequestHeader(this.requestHeader),a.setPath(this.path),a.setWithCredentials(s.withCredentials),a.load(t,function(l){let c;try{c=s.parse(l)}catch(d){if(i!==void 0)i(d);else{d(d);return}}c.image!==void 0?r.image=c.image:c.data!==void 0&&(r.image.width=c.width,r.image.height=c.height,r.image.data=c.data),r.wrapS=c.wrapS!==void 0?c.wrapS:Ln,r.wrapT=c.wrapT!==void 0?c.wrapT:Ln,r.magFilter=c.magFilter!==void 0?c.magFilter:Cn,r.minFilter=c.minFilter!==void 0?c.minFilter:Cn,r.anisotropy=c.anisotropy!==void 0?c.anisotropy:1,c.colorSpace!==void 0&&(r.colorSpace=c.colorSpace),c.flipY!==void 0&&(r.flipY=c.flipY),c.format!==void 0&&(r.format=c.format),c.type!==void 0&&(r.type=c.type),c.mipmaps!==void 0&&(r.mipmaps=c.mipmaps,r.minFilter=Rs),c.mipmapCount===1&&(r.minFilter=Cn),c.generateMipmaps!==void 0&&(r.generateMipmaps=c.generateMipmaps),r.needsUpdate=!0,e&&e(r,c)},n,i),r}}class Hp extends Wn{constructor(t){super(t)}load(t,e,n,i){const s=new Ye,r=new Ao(this.manager);return r.setCrossOrigin(this.crossOrigin),r.setPath(this.path),r.load(t,function(a){s.image=a,s.needsUpdate=!0,e!==void 0&&e(s)},n,i),s}}class _s extends De{constructor(t,e=1){super(),this.isLight=!0,this.type="Light",this.color=new re(t),this.intensity=e}dispose(){this.dispatchEvent({type:"dispose"})}copy(t,e){return super.copy(t,e),this.color.copy(t.color),this.intensity=t.intensity,this}toJSON(t){const e=super.toJSON(t);return e.object.color=this.color.getHex(),e.object.intensity=this.intensity,e}}class Gd extends _s{constructor(t,e,n){super(t,n),this.isHemisphereLight=!0,this.type="HemisphereLight",this.position.copy(De.DEFAULT_UP),this.updateMatrix(),this.groundColor=new re(e)}copy(t,e){return super.copy(t,e),this.groundColor.copy(t.groundColor),this}toJSON(t){const e=super.toJSON(t);return e.object.groundColor=this.groundColor.getHex(),e}}const fh=new Me,Wu=new O,Xu=new O;class dh{constructor(t){this.camera=t,this.intensity=1,this.bias=0,this.biasNode=null,this.normalBias=0,this.radius=1,this.blurSamples=8,this.mapSize=new At(512,512),this.mapType=Is,this.map=null,this.mapPass=null,this.matrix=new Me,this.autoUpdate=!0,this.needsUpdate=!1,this._frustum=new Ka,this._frameExtents=new At(1,1),this._viewportCount=1,this._viewports=[new En(0,0,1,1)]}getViewportCount(){return this._viewportCount}getFrustum(){return this._frustum}updateMatrices(t){const e=this.camera,n=this.matrix;Wu.setFromMatrixPosition(t.matrixWorld),e.position.copy(Wu),Xu.setFromMatrixPosition(t.target.matrixWorld),e.lookAt(Xu),e.updateMatrixWorld(),fh.multiplyMatrices(e.projectionMatrix,e.matrixWorldInverse),this._frustum.setFromProjectionMatrix(fh,e.coordinateSystem,e.reversedDepth),e.coordinateSystem===Pi||e.reversedDepth?n.set(.5,0,0,.5,0,.5,0,.5,0,0,1,0,0,0,0,1):n.set(.5,0,0,.5,0,.5,0,.5,0,0,.5,.5,0,0,0,1),n.multiply(fh)}getViewport(t){return this._viewports[t]}getFrameExtents(){return this._frameExtents}dispose(){this.map&&this.map.dispose(),this.mapPass&&this.mapPass.dispose()}copy(t){return this.camera=t.camera.clone(),this.intensity=t.intensity,this.bias=t.bias,this.radius=t.radius,this.autoUpdate=t.autoUpdate,this.needsUpdate=t.needsUpdate,this.normalBias=t.normalBias,this.blurSamples=t.blurSamples,this.mapSize.copy(t.mapSize),this.biasNode=t.biasNode,this}clone(){return new this.constructor().copy(this)}toJSON(){const t={};return this.intensity!==1&&(t.intensity=this.intensity),this.bias!==0&&(t.bias=this.bias),this.normalBias!==0&&(t.normalBias=this.normalBias),this.radius!==1&&(t.radius=this.radius),(this.mapSize.x!==512||this.mapSize.y!==512)&&(t.mapSize=this.mapSize.toArray()),t.camera=this.camera.toJSON(!1).object,delete t.camera.matrix,t}}const Eo=new O,wo=new un,mi=new O;class ph extends De{constructor(){super(),this.isCamera=!0,this.type="Camera",this.matrixWorldInverse=new Me,this.projectionMatrix=new Me,this.projectionMatrixInverse=new Me,this.coordinateSystem=Nn,this._reversedDepth=!1}get reversedDepth(){return this._reversedDepth}copy(t,e){return super.copy(t,e),this.matrixWorldInverse.copy(t.matrixWorldInverse),this.projectionMatrix.copy(t.projectionMatrix),this.projectionMatrixInverse.copy(t.projectionMatrixInverse),this.coordinateSystem=t.coordinateSystem,this}getWorldDirection(t){return super.getWorldDirection(t).negate()}updateMatrixWorld(t){super.updateMatrixWorld(t),this.matrixWorld.decompose(Eo,wo,mi),mi.x===1&&mi.y===1&&mi.z===1?this.matrixWorldInverse.copy(this.matrixWorld).invert():this.matrixWorldInverse.compose(Eo,wo,mi.set(1,1,1)).invert()}updateWorldMatrix(t,e){super.updateWorldMatrix(t,e),this.matrixWorld.decompose(Eo,wo,mi),mi.x===1&&mi.y===1&&mi.z===1?this.matrixWorldInverse.copy(this.matrixWorld).invert():this.matrixWorldInverse.compose(Eo,wo,mi.set(1,1,1)).invert()}clone(){return new this.constructor().copy(this)}}const Hi=new O,qu=new At,Yu=new At;class Xn extends ph{constructor(t=50,e=1,n=.1,i=2e3){super(),this.isPerspectiveCamera=!0,this.type="PerspectiveCamera",this.fov=t,this.zoom=1,this.near=n,this.far=i,this.focus=10,this.aspect=e,this.view=null,this.filmGauge=35,this.filmOffset=0,this.updateProjectionMatrix()}copy(t,e){return super.copy(t,e),this.fov=t.fov,this.zoom=t.zoom,this.near=t.near,this.far=t.far,this.focus=t.focus,this.aspect=t.aspect,this.view=t.view===null?null:Object.assign({},t.view),this.filmGauge=t.filmGauge,this.filmOffset=t.filmOffset,this}setFocalLength(t){const e=.5*this.getFilmHeight()/t;this.fov=Di*2*Math.atan(e),this.updateProjectionMatrix()}getFocalLength(){const t=Math.tan(_i*.5*this.fov);return .5*this.getFilmHeight()/t}getEffectiveFOV(){return Di*2*Math.atan(Math.tan(_i*.5*this.fov)/this.zoom)}getFilmWidth(){return this.filmGauge*Math.min(this.aspect,1)}getFilmHeight(){return this.filmGauge/Math.max(this.aspect,1)}getViewBounds(t,e,n){Hi.set(-1,-1,.5).applyMatrix4(this.projectionMatrixInverse),e.set(Hi.x,Hi.y).multiplyScalar(-t/Hi.z),Hi.set(1,1,.5).applyMatrix4(this.projectionMatrixInverse),n.set(Hi.x,Hi.y).multiplyScalar(-t/Hi.z)}getViewSize(t,e){return this.getViewBounds(t,qu,Yu),e.subVectors(Yu,qu)}setViewOffset(t,e,n,i,s,r){this.aspect=t/e,this.view===null&&(this.view={enabled:!0,fullWidth:1,fullHeight:1,offsetX:0,offsetY:0,width:1,height:1}),this.view.enabled=!0,this.view.fullWidth=t,this.view.fullHeight=e,this.view.offsetX=n,this.view.offsetY=i,this.view.width=s,this.view.height=r,this.updateProjectionMatrix()}clearViewOffset(){this.view!==null&&(this.view.enabled=!1),this.updateProjectionMatrix()}updateProjectionMatrix(){const t=this.near;let e=t*Math.tan(_i*.5*this.fov)/this.zoom,n=2*e,i=this.aspect*n,s=-.5*i;const r=this.view;if(this.view!==null&&this.view.enabled){const l=r.fullWidth,c=r.fullHeight;s+=r.offsetX*i/l,e-=r.offsetY*n/c,i*=r.width/l,n*=r.height/c}const a=this.filmOffset;a!==0&&(s+=t*a/this.getFilmWidth()),this.projectionMatrix.makePerspective(s,s+i,e,e-n,t,this.far,this.coordinateSystem,this.reversedDepth),this.projectionMatrixInverse.copy(this.projectionMatrix).invert()}toJSON(t){const e=super.toJSON(t);return e.object.fov=this.fov,e.object.zoom=this.zoom,e.object.near=this.near,e.object.far=this.far,e.object.focus=this.focus,e.object.aspect=this.aspect,this.view!==null&&(e.object.view=Object.assign({},this.view)),e.object.filmGauge=this.filmGauge,e.object.filmOffset=this.filmOffset,e}}class Hd extends dh{constructor(){super(new Xn(50,1,.5,500)),this.isSpotLightShadow=!0,this.focus=1,this.aspect=1}updateMatrices(t){const e=this.camera,n=Di*2*t.angle*this.focus,i=this.mapSize.width/this.mapSize.height*this.aspect,s=t.distance||e.far;(n!==e.fov||i!==e.aspect||s!==e.far)&&(e.fov=n,e.aspect=i,e.far=s,e.updateProjectionMatrix()),super.updateMatrices(t)}copy(t){return super.copy(t),this.focus=t.focus,this}}class Wd extends _s{constructor(t,e,n=0,i=Math.PI/3,s=0,r=2){super(t,e),this.isSpotLight=!0,this.type="SpotLight",this.position.copy(De.DEFAULT_UP),this.updateMatrix(),this.target=new De,this.distance=n,this.angle=i,this.penumbra=s,this.decay=r,this.map=null,this.shadow=new Hd}get power(){return this.intensity*Math.PI}set power(t){this.intensity=t/Math.PI}dispose(){super.dispose(),this.shadow.dispose()}copy(t,e){return super.copy(t,e),this.distance=t.distance,this.angle=t.angle,this.penumbra=t.penumbra,this.decay=t.decay,this.target=t.target.clone(),this.map=t.map,this.shadow=t.shadow.clone(),this}toJSON(t){const e=super.toJSON(t);return e.object.distance=this.distance,e.object.angle=this.angle,e.object.decay=this.decay,e.object.penumbra=this.penumbra,e.object.target=this.target.uuid,this.map&&this.map.isTexture&&(e.object.map=this.map.toJSON(t).uuid),e.object.shadow=this.shadow.toJSON(),e}}class Xd extends dh{constructor(){super(new Xn(90,1,.5,500)),this.isPointLightShadow=!0}}class qd extends _s{constructor(t,e,n=0,i=2){super(t,e),this.isPointLight=!0,this.type="PointLight",this.distance=n,this.decay=i,this.shadow=new Xd}get power(){return this.intensity*4*Math.PI}set power(t){this.intensity=t/(4*Math.PI)}dispose(){super.dispose(),this.shadow.dispose()}copy(t,e){return super.copy(t,e),this.distance=t.distance,this.decay=t.decay,this.shadow=t.shadow.clone(),this}toJSON(t){const e=super.toJSON(t);return e.object.distance=this.distance,e.object.decay=this.decay,e.object.shadow=this.shadow.toJSON(),e}}class mh extends ph{constructor(t=-1,e=1,n=1,i=-1,s=.1,r=2e3){super(),this.isOrthographicCamera=!0,this.type="OrthographicCamera",this.zoom=1,this.view=null,this.left=t,this.right=e,this.top=n,this.bottom=i,this.near=s,this.far=r,this.updateProjectionMatrix()}copy(t,e){return super.copy(t,e),this.left=t.left,this.right=t.right,this.top=t.top,this.bottom=t.bottom,this.near=t.near,this.far=t.far,this.zoom=t.zoom,this.view=t.view===null?null:Object.assign({},t.view),this}setViewOffset(t,e,n,i,s,r){this.view===null&&(this.view={enabled:!0,fullWidth:1,fullHeight:1,offsetX:0,offsetY:0,width:1,height:1}),this.view.enabled=!0,this.view.fullWidth=t,this.view.fullHeight=e,this.view.offsetX=n,this.view.offsetY=i,this.view.width=s,this.view.height=r,this.updateProjectionMatrix()}clearViewOffset(){this.view!==null&&(this.view.enabled=!1),this.updateProjectionMatrix()}updateProjectionMatrix(){const t=(this.right-this.left)/(2*this.zoom),e=(this.top-this.bottom)/(2*this.zoom),n=(this.right+this.left)/2,i=(this.top+this.bottom)/2;let s=n-t,r=n+t,a=i+e,l=i-e;if(this.view!==null&&this.view.enabled){const c=(this.right-this.left)/this.view.fullWidth/this.zoom,d=(this.top-this.bottom)/this.view.fullHeight/this.zoom;s+=c*this.view.offsetX,r=s+c*this.view.width,a-=d*this.view.offsetY,l=a-d*this.view.height}this.projectionMatrix.makeOrthographic(s,r,a,l,this.near,this.far,this.coordinateSystem,this.reversedDepth),this.projectionMatrixInverse.copy(this.projectionMatrix).invert()}toJSON(t){const e=super.toJSON(t);return e.object.zoom=this.zoom,e.object.left=this.left,e.object.right=this.right,e.object.top=this.top,e.object.bottom=this.bottom,e.object.near=this.near,e.object.far=this.far,this.view!==null&&(e.object.view=Object.assign({},this.view)),e}}class Yd extends dh{constructor(){super(new mh(-5,5,5,-5,.5,500)),this.isDirectionalLightShadow=!0}}class Zd extends _s{constructor(t,e){super(t,e),this.isDirectionalLight=!0,this.type="DirectionalLight",this.position.copy(De.DEFAULT_UP),this.updateMatrix(),this.target=new De,this.shadow=new Yd}dispose(){super.dispose(),this.shadow.dispose()}copy(t){return super.copy(t),this.target=t.target.clone(),this.shadow=t.shadow.clone(),this}toJSON(t){const e=super.toJSON(t);return e.object.shadow=this.shadow.toJSON(),e.object.target=this.target.uuid,e}}class $d extends _s{constructor(t,e){super(t,e),this.isAmbientLight=!0,this.type="AmbientLight"}}class Jd extends _s{constructor(t,e,n=10,i=10){super(t,e),this.isRectAreaLight=!0,this.type="RectAreaLight",this.width=n,this.height=i}get power(){return this.intensity*this.width*this.height*Math.PI}set power(t){this.intensity=t/(this.width*this.height*Math.PI)}copy(t){return super.copy(t),this.width=t.width,this.height=t.height,this}toJSON(t){const e=super.toJSON(t);return e.object.width=this.width,e.object.height=this.height,e}}class Zu{constructor(){this.isSphericalHarmonics3=!0,this.coefficients=[];for(let t=0;t<9;t++)this.coefficients.push(new O)}set(t){for(let e=0;e<9;e++)this.coefficients[e].copy(t[e]);return this}zero(){for(let t=0;t<9;t++)this.coefficients[t].set(0,0,0);return this}getAt(t,e){const n=t.x,i=t.y,s=t.z,r=this.coefficients;return e.copy(r[0]).multiplyScalar(.282095),e.addScaledVector(r[1],.488603*i),e.addScaledVector(r[2],.488603*s),e.addScaledVector(r[3],.488603*n),e.addScaledVector(r[4],1.092548*(n*i)),e.addScaledVector(r[5],1.092548*(i*s)),e.addScaledVector(r[6],.315392*(3*s*s-1)),e.addScaledVector(r[7],1.092548*(n*s)),e.addScaledVector(r[8],.546274*(n*n-i*i)),e}getIrradianceAt(t,e){const n=t.x,i=t.y,s=t.z,r=this.coefficients;return e.copy(r[0]).multiplyScalar(.886227),e.addScaledVector(r[1],2*.511664*i),e.addScaledVector(r[2],2*.511664*s),e.addScaledVector(r[3],2*.511664*n),e.addScaledVector(r[4],2*.429043*n*i),e.addScaledVector(r[5],2*.429043*i*s),e.addScaledVector(r[6],.743125*s*s-.247708),e.addScaledVector(r[7],2*.429043*n*s),e.addScaledVector(r[8],.429043*(n*n-i*i)),e}add(t){for(let e=0;e<9;e++)this.coefficients[e].add(t.coefficients[e]);return this}addScaledSH(t,e){for(let n=0;n<9;n++)this.coefficients[n].addScaledVector(t.coefficients[n],e);return this}scale(t){for(let e=0;e<9;e++)this.coefficients[e].multiplyScalar(t);return this}lerp(t,e){for(let n=0;n<9;n++)this.coefficients[n].lerp(t.coefficients[n],e);return this}equals(t){for(let e=0;e<9;e++)if(!this.coefficients[e].equals(t.coefficients[e]))return!1;return!0}copy(t){return this.set(t.coefficients)}clone(){return new this.constructor().copy(this)}fromArray(t,e=0){const n=this.coefficients;for(let i=0;i<9;i++)n[i].fromArray(t,e+i*3);return this}toArray(t=[],e=0){const n=this.coefficients;for(let i=0;i<9;i++)n[i].toArray(t,e+i*3);return t}static getBasisAt(t,e){const n=t.x,i=t.y,s=t.z;e[0]=.282095,e[1]=.488603*i,e[2]=.488603*s,e[3]=.488603*n,e[4]=1.092548*n*i,e[5]=1.092548*i*s,e[6]=.315392*(3*s*s-1),e[7]=1.092548*n*s,e[8]=.546274*(n*n-i*i)}}class Kd extends _s{constructor(t=new Zu,e=1){super(void 0,e),this.isLightProbe=!0,this.sh=t}copy(t){return super.copy(t),this.sh.copy(t.sh),this}toJSON(t){const e=super.toJSON(t);return e.object.sh=this.sh.toArray(),e}}class gh extends Wn{constructor(t){super(t),this.textures={}}load(t,e,n,i){const s=this,r=new Gi(s.manager);r.setPath(s.path),r.setRequestHeader(s.requestHeader),r.setWithCredentials(s.withCredentials),r.load(t,function(a){try{e(s.parse(JSON.parse(a)))}catch(l){i?i(l):Re(l),s.manager.itemError(t)}},n,i)}parse(t){const e=this.textures;function n(s){return e[s]===void 0&&le("MaterialLoader: Undefined texture",s),e[s]}const i=this.createMaterialFromType(t.type);if(t.uuid!==void 0&&(i.uuid=t.uuid),t.name!==void 0&&(i.name=t.name),t.color!==void 0&&i.color!==void 0&&i.color.setHex(t.color),t.roughness!==void 0&&(i.roughness=t.roughness),t.metalness!==void 0&&(i.metalness=t.metalness),t.sheen!==void 0&&(i.sheen=t.sheen),t.sheenColor!==void 0&&(i.sheenColor=new re().setHex(t.sheenColor)),t.sheenRoughness!==void 0&&(i.sheenRoughness=t.sheenRoughness),t.emissive!==void 0&&i.emissive!==void 0&&i.emissive.setHex(t.emissive),t.specular!==void 0&&i.specular!==void 0&&i.specular.setHex(t.specular),t.specularIntensity!==void 0&&(i.specularIntensity=t.specularIntensity),t.specularColor!==void 0&&i.specularColor!==void 0&&i.specularColor.setHex(t.specularColor),t.shininess!==void 0&&(i.shininess=t.shininess),t.clearcoat!==void 0&&(i.clearcoat=t.clearcoat),t.clearcoatRoughness!==void 0&&(i.clearcoatRoughness=t.clearcoatRoughness),t.dispersion!==void 0&&(i.dispersion=t.dispersion),t.iridescence!==void 0&&(i.iridescence=t.iridescence),t.iridescenceIOR!==void 0&&(i.iridescenceIOR=t.iridescenceIOR),t.iridescenceThicknessRange!==void 0&&(i.iridescenceThicknessRange=t.iridescenceThicknessRange),t.transmission!==void 0&&(i.transmission=t.transmission),t.thickness!==void 0&&(i.thickness=t.thickness),t.attenuationDistance!==void 0&&(i.attenuationDistance=t.attenuationDistance),t.attenuationColor!==void 0&&i.attenuationColor!==void 0&&i.attenuationColor.setHex(t.attenuationColor),t.anisotropy!==void 0&&(i.anisotropy=t.anisotropy),t.anisotropyRotation!==void 0&&(i.anisotropyRotation=t.anisotropyRotation),t.fog!==void 0&&(i.fog=t.fog),t.flatShading!==void 0&&(i.flatShading=t.flatShading),t.blending!==void 0&&(i.blending=t.blending),t.combine!==void 0&&(i.combine=t.combine),t.side!==void 0&&(i.side=t.side),t.shadowSide!==void 0&&(i.shadowSide=t.shadowSide),t.opacity!==void 0&&(i.opacity=t.opacity),t.transparent!==void 0&&(i.transparent=t.transparent),t.alphaTest!==void 0&&(i.alphaTest=t.alphaTest),t.alphaHash!==void 0&&(i.alphaHash=t.alphaHash),t.depthFunc!==void 0&&(i.depthFunc=t.depthFunc),t.depthTest!==void 0&&(i.depthTest=t.depthTest),t.depthWrite!==void 0&&(i.depthWrite=t.depthWrite),t.colorWrite!==void 0&&(i.colorWrite=t.colorWrite),t.blendSrc!==void 0&&(i.blendSrc=t.blendSrc),t.blendDst!==void 0&&(i.blendDst=t.blendDst),t.blendEquation!==void 0&&(i.blendEquation=t.blendEquation),t.blendSrcAlpha!==void 0&&(i.blendSrcAlpha=t.blendSrcAlpha),t.blendDstAlpha!==void 0&&(i.blendDstAlpha=t.blendDstAlpha),t.blendEquationAlpha!==void 0&&(i.blendEquationAlpha=t.blendEquationAlpha),t.blendColor!==void 0&&i.blendColor!==void 0&&i.blendColor.setHex(t.blendColor),t.blendAlpha!==void 0&&(i.blendAlpha=t.blendAlpha),t.stencilWriteMask!==void 0&&(i.stencilWriteMask=t.stencilWriteMask),t.stencilFunc!==void 0&&(i.stencilFunc=t.stencilFunc),t.stencilRef!==void 0&&(i.stencilRef=t.stencilRef),t.stencilFuncMask!==void 0&&(i.stencilFuncMask=t.stencilFuncMask),t.stencilFail!==void 0&&(i.stencilFail=t.stencilFail),t.stencilZFail!==void 0&&(i.stencilZFail=t.stencilZFail),t.stencilZPass!==void 0&&(i.stencilZPass=t.stencilZPass),t.stencilWrite!==void 0&&(i.stencilWrite=t.stencilWrite),t.wireframe!==void 0&&(i.wireframe=t.wireframe),t.wireframeLinewidth!==void 0&&(i.wireframeLinewidth=t.wireframeLinewidth),t.wireframeLinecap!==void 0&&(i.wireframeLinecap=t.wireframeLinecap),t.wireframeLinejoin!==void 0&&(i.wireframeLinejoin=t.wireframeLinejoin),t.rotation!==void 0&&(i.rotation=t.rotation),t.linewidth!==void 0&&(i.linewidth=t.linewidth),t.dashSize!==void 0&&(i.dashSize=t.dashSize),t.gapSize!==void 0&&(i.gapSize=t.gapSize),t.scale!==void 0&&(i.scale=t.scale),t.polygonOffset!==void 0&&(i.polygonOffset=t.polygonOffset),t.polygonOffsetFactor!==void 0&&(i.polygonOffsetFactor=t.polygonOffsetFactor),t.polygonOffsetUnits!==void 0&&(i.polygonOffsetUnits=t.polygonOffsetUnits),t.dithering!==void 0&&(i.dithering=t.dithering),t.alphaToCoverage!==void 0&&(i.alphaToCoverage=t.alphaToCoverage),t.premultipliedAlpha!==void 0&&(i.premultipliedAlpha=t.premultipliedAlpha),t.forceSinglePass!==void 0&&(i.forceSinglePass=t.forceSinglePass),t.allowOverride!==void 0&&(i.allowOverride=t.allowOverride),t.visible!==void 0&&(i.visible=t.visible),t.toneMapped!==void 0&&(i.toneMapped=t.toneMapped),t.userData!==void 0&&(i.userData=t.userData),t.vertexColors!==void 0&&(typeof t.vertexColors=="number"?i.vertexColors=t.vertexColors>0:i.vertexColors=t.vertexColors),t.uniforms!==void 0)for(const s in t.uniforms){const r=t.uniforms[s];switch(i.uniforms[s]={},r.type){case"t":i.uniforms[s].value=n(r.value);break;case"c":i.uniforms[s].value=new re().setHex(r.value);break;case"v2":i.uniforms[s].value=new At().fromArray(r.value);break;case"v3":i.uniforms[s].value=new O().fromArray(r.value);break;case"v4":i.uniforms[s].value=new En().fromArray(r.value);break;case"m3":i.uniforms[s].value=new Fn().fromArray(r.value);break;case"m4":i.uniforms[s].value=new Me().fromArray(r.value);break;default:i.uniforms[s].value=r.value}}if(t.defines!==void 0&&(i.defines=t.defines),t.vertexShader!==void 0&&(i.vertexShader=t.vertexShader),t.fragmentShader!==void 0&&(i.fragmentShader=t.fragmentShader),t.glslVersion!==void 0&&(i.glslVersion=t.glslVersion),t.extensions!==void 0)for(const s in t.extensions)i.extensions[s]=t.extensions[s];if(t.lights!==void 0&&(i.lights=t.lights),t.clipping!==void 0&&(i.clipping=t.clipping),t.size!==void 0&&(i.size=t.size),t.sizeAttenuation!==void 0&&(i.sizeAttenuation=t.sizeAttenuation),t.map!==void 0&&(i.map=n(t.map)),t.matcap!==void 0&&(i.matcap=n(t.matcap)),t.alphaMap!==void 0&&(i.alphaMap=n(t.alphaMap)),t.bumpMap!==void 0&&(i.bumpMap=n(t.bumpMap)),t.bumpScale!==void 0&&(i.bumpScale=t.bumpScale),t.normalMap!==void 0&&(i.normalMap=n(t.normalMap)),t.normalMapType!==void 0&&(i.normalMapType=t.normalMapType),t.normalScale!==void 0){let s=t.normalScale;Array.isArray(s)===!1&&(s=[s,s]),i.normalScale=new At().fromArray(s)}return t.displacementMap!==void 0&&(i.displacementMap=n(t.displacementMap)),t.displacementScale!==void 0&&(i.displacementScale=t.displacementScale),t.displacementBias!==void 0&&(i.displacementBias=t.displacementBias),t.roughnessMap!==void 0&&(i.roughnessMap=n(t.roughnessMap)),t.metalnessMap!==void 0&&(i.metalnessMap=n(t.metalnessMap)),t.emissiveMap!==void 0&&(i.emissiveMap=n(t.emissiveMap)),t.emissiveIntensity!==void 0&&(i.emissiveIntensity=t.emissiveIntensity),t.specularMap!==void 0&&(i.specularMap=n(t.specularMap)),t.specularIntensityMap!==void 0&&(i.specularIntensityMap=n(t.specularIntensityMap)),t.specularColorMap!==void 0&&(i.specularColorMap=n(t.specularColorMap)),t.envMap!==void 0&&(i.envMap=n(t.envMap)),t.envMapRotation!==void 0&&i.envMapRotation.fromArray(t.envMapRotation),t.envMapIntensity!==void 0&&(i.envMapIntensity=t.envMapIntensity),t.reflectivity!==void 0&&(i.reflectivity=t.reflectivity),t.refractionRatio!==void 0&&(i.refractionRatio=t.refractionRatio),t.lightMap!==void 0&&(i.lightMap=n(t.lightMap)),t.lightMapIntensity!==void 0&&(i.lightMapIntensity=t.lightMapIntensity),t.aoMap!==void 0&&(i.aoMap=n(t.aoMap)),t.aoMapIntensity!==void 0&&(i.aoMapIntensity=t.aoMapIntensity),t.gradientMap!==void 0&&(i.gradientMap=n(t.gradientMap)),t.clearcoatMap!==void 0&&(i.clearcoatMap=n(t.clearcoatMap)),t.clearcoatRoughnessMap!==void 0&&(i.clearcoatRoughnessMap=n(t.clearcoatRoughnessMap)),t.clearcoatNormalMap!==void 0&&(i.clearcoatNormalMap=n(t.clearcoatNormalMap)),t.clearcoatNormalScale!==void 0&&(i.clearcoatNormalScale=new At().fromArray(t.clearcoatNormalScale)),t.iridescenceMap!==void 0&&(i.iridescenceMap=n(t.iridescenceMap)),t.iridescenceThicknessMap!==void 0&&(i.iridescenceThicknessMap=n(t.iridescenceThicknessMap)),t.transmissionMap!==void 0&&(i.transmissionMap=n(t.transmissionMap)),t.thicknessMap!==void 0&&(i.thicknessMap=n(t.thicknessMap)),t.anisotropyMap!==void 0&&(i.anisotropyMap=n(t.anisotropyMap)),t.sheenColorMap!==void 0&&(i.sheenColorMap=n(t.sheenColorMap)),t.sheenRoughnessMap!==void 0&&(i.sheenRoughnessMap=n(t.sheenRoughnessMap)),i}setTextures(t){return this.textures=t,this}createMaterialFromType(t){return gh.createMaterialFromType(t)}static createMaterialFromType(t){const e={ShadowMaterial:vd,SpriteMaterial:Pt,RawShaderMaterial:Nu,ShaderMaterial:ch,PointsMaterial:mu,MeshPhysicalMaterial:Ad,MeshStandardMaterial:Fu,MeshPhongMaterial:Ed,MeshToonMaterial:wd,MeshNormalMaterial:Cd,MeshLambertMaterial:Rd,MeshDepthMaterial:Ou,MeshDistanceMaterial:Bu,MeshBasicMaterial:ei,MeshMatcapMaterial:Id,LineDashedMaterial:Pd,LineBasicMaterial:Pn,Material:Ft};return new e[t]}}class $u{static extractUrlBase(t){const e=t.lastIndexOf("/");return e===-1?"./":t.slice(0,e+1)}static resolveURL(t,e){return typeof t!="string"||t===""?"":(/^https?:\/\//i.test(e)&&/^\//.test(t)&&(e=e.replace(/(^https?:\/\/[^\/]+).*/i,"$1")),/^(https?:)?\/\//i.test(t)||/^data:.*,.*$/i.test(t)||/^blob:.*$/i.test(t)?t:e+t)}}class Qd extends y{constructor(){super(),this.isInstancedBufferGeometry=!0,this.type="InstancedBufferGeometry",this.instanceCount=1/0}copy(t){return super.copy(t),this.instanceCount=t.instanceCount,this}toJSON(){const t=super.toJSON();return t.instanceCount=this.instanceCount,t.isInstancedBufferGeometry=!0,t}}class jd extends Wn{constructor(t){super(t)}load(t,e,n,i){const s=this,r=new Gi(s.manager);r.setPath(s.path),r.setRequestHeader(s.requestHeader),r.setWithCredentials(s.withCredentials),r.load(t,function(a){try{e(s.parse(JSON.parse(a)))}catch(l){i?i(l):Re(l),s.manager.itemError(t)}},n,i)}parse(t){const e={},n={};function i(_,v){if(e[v]!==void 0)return e[v];const C=_.interleavedBuffers[v],w=s(_,C.buffer),L=Li(C.type,w),D=new W(L,C.stride);return D.uuid=C.uuid,e[v]=D,D}function s(_,v){if(n[v]!==void 0)return n[v];const C=_.arrayBuffers[v],w=new Uint32Array(C).buffer;return n[v]=w,w}const r=t.isInstancedBufferGeometry?new Qd:new y,a=t.data.index;if(a!==void 0){const _=Li(a.type,a.array);r.setIndex(new se(_,1))}const l=t.data.attributes;for(const _ in l){const v=l[_];let M;if(v.isInterleavedBufferAttribute){const C=i(t.data,v.data);M=new yt(C,v.itemSize,v.offset,v.normalized)}else{const C=Li(v.type,v.array),w=v.isInstancedBufferAttribute?Ue:se;M=new w(C,v.itemSize,v.normalized)}v.name!==void 0&&(M.name=v.name),v.usage!==void 0&&M.setUsage(v.usage),r.setAttribute(_,M)}const c=t.data.morphAttributes;if(c)for(const _ in c){const v=c[_],M=[];for(let C=0,w=v.length;C<w;C++){const L=v[C];let D;if(L.isInterleavedBufferAttribute){const N=i(t.data,L.data);D=new yt(N,L.itemSize,L.offset,L.normalized)}else{const N=Li(L.type,L.array);D=new se(N,L.itemSize,L.normalized)}L.name!==void 0&&(D.name=L.name),M.push(D)}r.morphAttributes[_]=M}t.data.morphTargetsRelative&&(r.morphTargetsRelative=!0);const m=t.data.groups||t.data.drawcalls||t.data.offsets;if(m!==void 0)for(let _=0,v=m.length;_!==v;++_){const M=m[_];r.addGroup(M.start,M.count,M.materialIndex)}const g=t.data.boundingSphere;return g!==void 0&&(r.boundingSphere=new ae().fromJSON(g)),t.name&&(r.name=t.name),t.userData&&(r.userData=t.userData),r}}class Wp extends Wn{constructor(t){super(t)}load(t,e,n,i){const s=this,r=this.path===""?$u.extractUrlBase(t):this.path;this.resourcePath=this.resourcePath||r;const a=new Gi(this.manager);a.setPath(this.path),a.setRequestHeader(this.requestHeader),a.setWithCredentials(this.withCredentials),a.load(t,function(l){let c=null;try{c=JSON.parse(l)}catch(m){i!==void 0&&i(m),m("ObjectLoader: Can't parse "+t+".",m.message);return}const d=c.metadata;if(d===void 0||d.type===void 0||d.type.toLowerCase()==="geometry"){i!==void 0&&i(new Error("THREE.ObjectLoader: Can't load "+t)),Re("ObjectLoader: Can't load "+t);return}s.parse(c,e)},n,i)}async loadAsync(t,e){const n=this,i=this.path===""?$u.extractUrlBase(t):this.path;this.resourcePath=this.resourcePath||i;const s=new Gi(this.manager);s.setPath(this.path),s.setRequestHeader(this.requestHeader),s.setWithCredentials(this.withCredentials);const r=await s.loadAsync(t,e);let a;try{a=JSON.parse(r)}catch(c){throw new Error("ObjectLoader: Can't parse "+t+". "+c.message)}const l=a.metadata;if(l===void 0||l.type===void 0||l.type.toLowerCase()==="geometry")throw new Error("THREE.ObjectLoader: Can't load "+t);return await n.parseAsync(a)}parse(t,e){const n=this.parseAnimations(t.animations),i=this.parseShapes(t.shapes),s=this.parseGeometries(t.geometries,i),r=this.parseImages(t.images,function(){e!==void 0&&e(c)}),a=this.parseTextures(t.textures,r),l=this.parseMaterials(t.materials,a),c=this.parseObject(t.object,s,l,a,n),d=this.parseSkeletons(t.skeletons,c);if(this.bindSkeletons(c,d),this.bindLightTargets(c),e!==void 0){let m=!1;for(const g in r)if(r[g].data instanceof HTMLImageElement){m=!0;break}m===!1&&e(c)}return c}async parseAsync(t){const e=this.parseAnimations(t.animations),n=this.parseShapes(t.shapes),i=this.parseGeometries(t.geometries,n),s=await this.parseImagesAsync(t.images),r=this.parseTextures(t.textures,s),a=this.parseMaterials(t.materials,r),l=this.parseObject(t.object,i,a,r,e),c=this.parseSkeletons(t.skeletons,l);return this.bindSkeletons(l,c),this.bindLightTargets(l),l}parseShapes(t){const e={};if(t!==void 0)for(let n=0,i=t.length;n<i;n++){const s=new ks().fromJSON(t[n]);e[s.uuid]=s}return e}parseSkeletons(t,e){const n={},i={};if(e.traverse(function(s){s.isBone&&(i[s.uuid]=s)}),t!==void 0)for(let s=0,r=t.length;s<r;s++){const a=new He().fromJSON(t[s],i);n[a.uuid]=a}return n}parseGeometries(t,e){const n={};if(t!==void 0){const i=new jd;for(let s=0,r=t.length;s<r;s++){let a;const l=t[s];switch(l.type){case"BufferGeometry":case"InstancedBufferGeometry":a=i.parse(l);break;default:l.type in Du?a=Du[l.type].fromJSON(l,e):le(`ObjectLoader: Unsupported geometry type "${l.type}"`)}a.uuid=l.uuid,l.name!==void 0&&(a.name=l.name),l.userData!==void 0&&(a.userData=l.userData),n[l.uuid]=a}}return n}parseMaterials(t,e){const n={},i={};if(t!==void 0){const s=new gh;s.setTextures(e);for(let r=0,a=t.length;r<a;r++){const l=t[r];n[l.uuid]===void 0&&(n[l.uuid]=s.parse(l)),i[l.uuid]=n[l.uuid]}}return i}parseAnimations(t){const e={};if(t!==void 0)for(let n=0;n<t.length;n++){const i=t[n],s=To.parse(i);e[s.uuid]=s}return e}parseImages(t,e){const n=this,i={};let s;function r(l){return n.manager.itemStart(l),s.load(l,function(){n.manager.itemEnd(l)},void 0,function(){n.manager.itemError(l),n.manager.itemEnd(l)})}function a(l){if(typeof l=="string"){const c=l,d=/^(\/\/)|([a-z]+:(\/\/)?)/i.test(c)?c:n.resourcePath+c;return r(d)}else return l.data?{data:Li(l.type,l.data),width:l.width,height:l.height}:null}if(t!==void 0&&t.length>0){const l=new Hu(e);s=new Ao(l),s.setCrossOrigin(this.crossOrigin);for(let c=0,d=t.length;c<d;c++){const m=t[c],g=m.url;if(Array.isArray(g)){const _=[];for(let v=0,M=g.length;v<M;v++){const C=g[v],w=a(C);w!==null&&(w instanceof HTMLImageElement?_.push(w):_.push(new Yt(w.data,w.width,w.height)))}i[m.uuid]=new oi(_)}else{const _=a(m.url);i[m.uuid]=new oi(_)}}}return i}async parseImagesAsync(t){const e=this,n={};let i;async function s(r){if(typeof r=="string"){const a=r,l=/^(\/\/)|([a-z]+:(\/\/)?)/i.test(a)?a:e.resourcePath+a;return await i.loadAsync(l)}else return r.data?{data:Li(r.type,r.data),width:r.width,height:r.height}:null}if(t!==void 0&&t.length>0){i=new Ao(this.manager),i.setCrossOrigin(this.crossOrigin);for(let r=0,a=t.length;r<a;r++){const l=t[r],c=l.url;if(Array.isArray(c)){const d=[];for(let m=0,g=c.length;m<g;m++){const _=c[m],v=await s(_);v!==null&&(v instanceof HTMLImageElement?d.push(v):d.push(new Yt(v.data,v.width,v.height)))}n[l.uuid]=new oi(d)}else{const d=await s(l.url);n[l.uuid]=new oi(d)}}}return n}parseTextures(t,e){function n(s,r){return typeof s=="number"?s:(le("ObjectLoader.parseTexture: Constant should be in numeric form.",s),r[s])}const i={};if(t!==void 0)for(let s=0,r=t.length;s<r;s++){const a=t[s];a.image===void 0&&le('ObjectLoader: No "image" specified for',a.uuid),e[a.image]===void 0&&le("ObjectLoader: Undefined image",a.image);const l=e[a.image],c=l.data;let d;Array.isArray(c)?(d=new kc,c.length===6&&(d.needsUpdate=!0)):(c&&c.data?d=new Yt:d=new Ye,c&&(d.needsUpdate=!0)),d.source=l,d.uuid=a.uuid,a.name!==void 0&&(d.name=a.name),a.mapping!==void 0&&(d.mapping=n(a.mapping,tp)),a.channel!==void 0&&(d.channel=a.channel),a.offset!==void 0&&d.offset.fromArray(a.offset),a.repeat!==void 0&&d.repeat.fromArray(a.repeat),a.center!==void 0&&d.center.fromArray(a.center),a.rotation!==void 0&&(d.rotation=a.rotation),a.wrap!==void 0&&(d.wrapS=n(a.wrap[0],Ju),d.wrapT=n(a.wrap[1],Ju)),a.format!==void 0&&(d.format=a.format),a.internalFormat!==void 0&&(d.internalFormat=a.internalFormat),a.type!==void 0&&(d.type=a.type),a.colorSpace!==void 0&&(d.colorSpace=a.colorSpace),a.minFilter!==void 0&&(d.minFilter=n(a.minFilter,Ku)),a.magFilter!==void 0&&(d.magFilter=n(a.magFilter,Ku)),a.anisotropy!==void 0&&(d.anisotropy=a.anisotropy),a.flipY!==void 0&&(d.flipY=a.flipY),a.generateMipmaps!==void 0&&(d.generateMipmaps=a.generateMipmaps),a.premultiplyAlpha!==void 0&&(d.premultiplyAlpha=a.premultiplyAlpha),a.unpackAlignment!==void 0&&(d.unpackAlignment=a.unpackAlignment),a.compareFunction!==void 0&&(d.compareFunction=a.compareFunction),a.userData!==void 0&&(d.userData=a.userData),i[a.uuid]=d}return i}parseObject(t,e,n,i,s){let r;function a(g){return e[g]===void 0&&le("ObjectLoader: Undefined geometry",g),e[g]}function l(g){if(g!==void 0){if(Array.isArray(g)){const _=[];for(let v=0,M=g.length;v<M;v++){const C=g[v];n[C]===void 0&&le("ObjectLoader: Undefined material",C),_.push(n[C])}return _}return n[g]===void 0&&le("ObjectLoader: Undefined material",g),n[g]}}function c(g){return i[g]===void 0&&le("ObjectLoader: Undefined texture",g),i[g]}let d,m;switch(t.type){case"Scene":r=new p,t.background!==void 0&&(Number.isInteger(t.background)?r.background=new re(t.background):r.background=c(t.background)),t.environment!==void 0&&(r.environment=c(t.environment)),t.fog!==void 0&&(t.fog.type==="Fog"?r.fog=new o(t.fog.color,t.fog.near,t.fog.far):t.fog.type==="FogExp2"&&(r.fog=new Pr(t.fog.color,t.fog.density)),t.fog.name!==""&&(r.fog.name=t.fog.name)),t.backgroundBlurriness!==void 0&&(r.backgroundBlurriness=t.backgroundBlurriness),t.backgroundIntensity!==void 0&&(r.backgroundIntensity=t.backgroundIntensity),t.backgroundRotation!==void 0&&r.backgroundRotation.fromArray(t.backgroundRotation),t.environmentIntensity!==void 0&&(r.environmentIntensity=t.environmentIntensity),t.environmentRotation!==void 0&&r.environmentRotation.fromArray(t.environmentRotation);break;case"PerspectiveCamera":r=new Xn(t.fov,t.aspect,t.near,t.far),t.focus!==void 0&&(r.focus=t.focus),t.zoom!==void 0&&(r.zoom=t.zoom),t.filmGauge!==void 0&&(r.filmGauge=t.filmGauge),t.filmOffset!==void 0&&(r.filmOffset=t.filmOffset),t.view!==void 0&&(r.view=Object.assign({},t.view));break;case"OrthographicCamera":r=new mh(t.left,t.right,t.top,t.bottom,t.near,t.far),t.zoom!==void 0&&(r.zoom=t.zoom),t.view!==void 0&&(r.view=Object.assign({},t.view));break;case"AmbientLight":r=new $d(t.color,t.intensity);break;case"DirectionalLight":r=new Zd(t.color,t.intensity),r.target=t.target||"";break;case"PointLight":r=new qd(t.color,t.intensity,t.distance,t.decay);break;case"RectAreaLight":r=new Jd(t.color,t.intensity,t.width,t.height);break;case"SpotLight":r=new Wd(t.color,t.intensity,t.distance,t.angle,t.penumbra,t.decay),r.target=t.target||"";break;case"HemisphereLight":r=new Gd(t.color,t.groundColor,t.intensity);break;case"LightProbe":const g=new Zu().fromArray(t.sh);r=new Kd(g,t.intensity);break;case"SkinnedMesh":d=a(t.geometry),m=l(t.material),r=new be(d,m),t.bindMode!==void 0&&(r.bindMode=t.bindMode),t.bindMatrix!==void 0&&r.bindMatrix.fromArray(t.bindMatrix),t.skeleton!==void 0&&(r.skeleton=t.skeleton);break;case"Mesh":d=a(t.geometry),m=l(t.material),r=new kn(d,m);break;case"InstancedMesh":d=a(t.geometry),m=l(t.material);const _=t.count,v=t.instanceMatrix,M=t.instanceColor;r=new Vi(d,m,_),r.instanceMatrix=new Ue(new Float32Array(v.array),16),M!==void 0&&(r.instanceColor=new Ue(new Float32Array(M.array),M.itemSize));break;case"BatchedMesh":d=a(t.geometry),m=l(t.material),r=new Lf(t.maxInstanceCount,t.maxVertexCount,t.maxIndexCount,m),r.geometry=d,r.perObjectFrustumCulled=t.perObjectFrustumCulled,r.sortObjects=t.sortObjects,r._drawRanges=t.drawRanges,r._reservedRanges=t.reservedRanges,r._geometryInfo=t.geometryInfo.map(C=>{let w=null,L=null;return C.boundingBox!==void 0&&(w=new A().fromJSON(C.boundingBox)),C.boundingSphere!==void 0&&(L=new ae().fromJSON(C.boundingSphere)),{...C,boundingBox:w,boundingSphere:L}}),r._instanceInfo=t.instanceInfo,r._availableInstanceIds=t._availableInstanceIds,r._availableGeometryIds=t._availableGeometryIds,r._nextIndexStart=t.nextIndexStart,r._nextVertexStart=t.nextVertexStart,r._geometryCount=t.geometryCount,r._maxInstanceCount=t.maxInstanceCount,r._maxVertexCount=t.maxVertexCount,r._maxIndexCount=t.maxIndexCount,r._geometryInitialized=t.geometryInitialized,r._matricesTexture=c(t.matricesTexture.uuid),r._indirectTexture=c(t.indirectTexture.uuid),t.colorsTexture!==void 0&&(r._colorsTexture=c(t.colorsTexture.uuid)),t.boundingSphere!==void 0&&(r.boundingSphere=new ae().fromJSON(t.boundingSphere)),t.boundingBox!==void 0&&(r.boundingBox=new A().fromJSON(t.boundingBox));break;case"LOD":r=new ze;break;case"Line":r=new ds(a(t.geometry),l(t.material));break;case"LineLoop":r=new Df(a(t.geometry),l(t.material));break;case"LineSegments":r=new Si(a(t.geometry),l(t.material));break;case"PointCloud":case"Points":r=new Uf(a(t.geometry),l(t.material));break;case"Sprite":r=new ft(l(t.material));break;case"Group":r=new rs;break;case"Bone":r=new Ae;break;default:r=new De}if(r.uuid=t.uuid,t.name!==void 0&&(r.name=t.name),t.matrix!==void 0?(r.matrix.fromArray(t.matrix),t.matrixAutoUpdate!==void 0&&(r.matrixAutoUpdate=t.matrixAutoUpdate),r.matrixAutoUpdate&&r.matrix.decompose(r.position,r.quaternion,r.scale)):(t.position!==void 0&&r.position.fromArray(t.position),t.rotation!==void 0&&r.rotation.fromArray(t.rotation),t.quaternion!==void 0&&r.quaternion.fromArray(t.quaternion),t.scale!==void 0&&r.scale.fromArray(t.scale)),t.up!==void 0&&r.up.fromArray(t.up),t.pivot!==void 0&&(r.pivot=new O().fromArray(t.pivot)),t.morphTargetDictionary!==void 0&&(r.morphTargetDictionary=Object.assign({},t.morphTargetDictionary)),t.morphTargetInfluences!==void 0&&(r.morphTargetInfluences=t.morphTargetInfluences.slice()),t.castShadow!==void 0&&(r.castShadow=t.castShadow),t.receiveShadow!==void 0&&(r.receiveShadow=t.receiveShadow),t.shadow&&(t.shadow.intensity!==void 0&&(r.shadow.intensity=t.shadow.intensity),t.shadow.bias!==void 0&&(r.shadow.bias=t.shadow.bias),t.shadow.normalBias!==void 0&&(r.shadow.normalBias=t.shadow.normalBias),t.shadow.radius!==void 0&&(r.shadow.radius=t.shadow.radius),t.shadow.mapSize!==void 0&&r.shadow.mapSize.fromArray(t.shadow.mapSize),t.shadow.camera!==void 0&&(r.shadow.camera=this.parseObject(t.shadow.camera))),t.visible!==void 0&&(r.visible=t.visible),t.frustumCulled!==void 0&&(r.frustumCulled=t.frustumCulled),t.renderOrder!==void 0&&(r.renderOrder=t.renderOrder),t.static!==void 0&&(r.static=t.static),t.userData!==void 0&&(r.userData=t.userData),t.layers!==void 0&&(r.layers.mask=t.layers),t.children!==void 0){const g=t.children;for(let _=0;_<g.length;_++)r.add(this.parseObject(g[_],e,n,i,s))}if(t.animations!==void 0){const g=t.animations;for(let _=0;_<g.length;_++){const v=g[_];r.animations.push(s[v])}}if(t.type==="LOD"){t.autoUpdate!==void 0&&(r.autoUpdate=t.autoUpdate);const g=t.levels;for(let _=0;_<g.length;_++){const v=g[_],M=r.getObjectByProperty("uuid",v.object);M!==void 0&&r.addLevel(M,v.distance,v.hysteresis)}}return r}bindSkeletons(t,e){Object.keys(e).length!==0&&t.traverse(function(n){if(n.isSkinnedMesh===!0&&n.skeleton!==void 0){const i=e[n.skeleton];i===void 0?le("ObjectLoader: No skeleton found with UUID:",n.skeleton):n.bind(i,n.bindMatrix)}})}bindLightTargets(t){t.traverse(function(e){if(e.isDirectionalLight||e.isSpotLight){const n=e.target,i=t.getObjectByProperty("uuid",n);i!==void 0?e.target=i:e.target=new De}})}}const tp={UVMapping:lr,CubeReflectionMapping:Yi,CubeRefractionMapping:$r,EquirectangularReflectionMapping:Jr,EquirectangularRefractionMapping:Kr,CubeUVReflectionMapping:Qr},Ju={RepeatWrapping:ws,ClampToEdgeWrapping:Ln,MirroredRepeatWrapping:Cs},Ku={NearestFilter:Mn,NearestMipmapNearestFilter:jr,NearestMipmapLinearFilter:ta,LinearFilter:Cn,LinearMipmapNearestFilter:ea,LinearMipmapLinearFilter:Rs},_h=new WeakMap;class Xp extends Wn{constructor(t){super(t),this.isImageBitmapLoader=!0,typeof createImageBitmap>"u"&&le("ImageBitmapLoader: createImageBitmap() not supported."),typeof fetch>"u"&&le("ImageBitmapLoader: fetch() not supported."),this.options={premultiplyAlpha:"none"},this._abortController=new AbortController}setOptions(t){return this.options=t,this}load(t,e,n,i){t===void 0&&(t=""),this.path!==void 0&&(t=this.path+t),t=this.manager.resolveURL(t);const s=this,r=bi.get(`image-bitmap:${t}`);if(r!==void 0){if(s.manager.itemStart(t),r.then){r.then(c=>{if(_h.has(r)===!0)i&&i(_h.get(r)),s.manager.itemError(t),s.manager.itemEnd(t);else return e&&e(c),s.manager.itemEnd(t),c});return}return setTimeout(function(){e&&e(r),s.manager.itemEnd(t)},0),r}const a={};a.credentials=this.crossOrigin==="anonymous"?"same-origin":"include",a.headers=this.requestHeader,a.signal=typeof AbortSignal.any=="function"?AbortSignal.any([this._abortController.signal,this.manager.abortController.signal]):this._abortController.signal;const l=fetch(t,a).then(function(c){return c.blob()}).then(function(c){return createImageBitmap(c,Object.assign(s.options,{colorSpaceConversion:"none"}))}).then(function(c){return bi.add(`image-bitmap:${t}`,c),e&&e(c),s.manager.itemEnd(t),c}).catch(function(c){i&&i(c),_h.set(l,c),bi.remove(`image-bitmap:${t}`),s.manager.itemError(t),s.manager.itemEnd(t)});bi.add(`image-bitmap:${t}`,l),s.manager.itemStart(t)}abort(){return this._abortController.abort(),this._abortController=new AbortController,this}}let Co;class Qu{static getContext(){return Co===void 0&&(Co=new(window.AudioContext||window.webkitAudioContext)),Co}static setContext(t){Co=t}}class qp extends Wn{constructor(t){super(t)}load(t,e,n,i){const s=this,r=new Gi(this.manager);r.setResponseType("arraybuffer"),r.setPath(this.path),r.setRequestHeader(this.requestHeader),r.setWithCredentials(this.withCredentials),r.load(t,function(l){try{const c=l.slice(0);Qu.getContext().decodeAudioData(c,function(m){e(m)}).catch(a)}catch(c){a(c)}},n,i);function a(l){i?i(l):Re(l),s.manager.itemError(t)}}}const ju=new Me,tf=new Me,xs=new Me;class Yp{constructor(){this.type="StereoCamera",this.aspect=1,this.eyeSep=.064,this.cameraL=new Xn,this.cameraL.layers.enable(1),this.cameraL.matrixAutoUpdate=!1,this.cameraR=new Xn,this.cameraR.layers.enable(2),this.cameraR.matrixAutoUpdate=!1,this._cache={focus:null,fov:null,aspect:null,near:null,far:null,zoom:null,eyeSep:null}}update(t){const e=this._cache;if(e.focus!==t.focus||e.fov!==t.fov||e.aspect!==t.aspect*this.aspect||e.near!==t.near||e.far!==t.far||e.zoom!==t.zoom||e.eyeSep!==this.eyeSep){e.focus=t.focus,e.fov=t.fov,e.aspect=t.aspect*this.aspect,e.near=t.near,e.far=t.far,e.zoom=t.zoom,e.eyeSep=this.eyeSep,xs.copy(t.projectionMatrix);const i=e.eyeSep/2,s=i*e.near/e.focus,r=e.near*Math.tan(_i*e.fov*.5)/e.zoom;let a,l;tf.elements[12]=-i,ju.elements[12]=i,a=-r*e.aspect+s,l=r*e.aspect+s,xs.elements[0]=2*e.near/(l-a),xs.elements[8]=(l+a)/(l-a),this.cameraL.projectionMatrix.copy(xs),a=-r*e.aspect-s,l=r*e.aspect-s,xs.elements[0]=2*e.near/(l-a),xs.elements[8]=(l+a)/(l-a),this.cameraR.projectionMatrix.copy(xs)}this.cameraL.matrixWorld.copy(t.matrixWorld).multiply(tf),this.cameraR.matrixWorld.copy(t.matrixWorld).multiply(ju)}}const qs=-90,Ys=1;class ep extends De{constructor(t,e,n){super(),this.type="CubeCamera",this.renderTarget=n,this.coordinateSystem=null,this.activeMipmapLevel=0;const i=new Xn(qs,Ys,t,e);i.layers=this.layers,this.add(i);const s=new Xn(qs,Ys,t,e);s.layers=this.layers,this.add(s);const r=new Xn(qs,Ys,t,e);r.layers=this.layers,this.add(r);const a=new Xn(qs,Ys,t,e);a.layers=this.layers,this.add(a);const l=new Xn(qs,Ys,t,e);l.layers=this.layers,this.add(l);const c=new Xn(qs,Ys,t,e);c.layers=this.layers,this.add(c)}updateCoordinateSystem(){const t=this.coordinateSystem,e=this.children.concat(),[n,i,s,r,a,l]=e;for(const c of e)this.remove(c);if(t===Nn)n.up.set(0,1,0),n.lookAt(1,0,0),i.up.set(0,1,0),i.lookAt(-1,0,0),s.up.set(0,0,-1),s.lookAt(0,1,0),r.up.set(0,0,1),r.lookAt(0,-1,0),a.up.set(0,1,0),a.lookAt(0,0,1),l.up.set(0,1,0),l.lookAt(0,0,-1);else if(t===Pi)n.up.set(0,-1,0),n.lookAt(-1,0,0),i.up.set(0,-1,0),i.lookAt(1,0,0),s.up.set(0,0,1),s.lookAt(0,1,0),r.up.set(0,0,-1),r.lookAt(0,-1,0),a.up.set(0,-1,0),a.lookAt(0,0,1),l.up.set(0,-1,0),l.lookAt(0,0,-1);else throw new Error("THREE.CubeCamera.updateCoordinateSystem(): Invalid coordinate system: "+t);for(const c of e)this.add(c),c.updateMatrixWorld()}update(t,e){this.parent===null&&this.updateMatrixWorld();const{renderTarget:n,activeMipmapLevel:i}=this;this.coordinateSystem!==t.coordinateSystem&&(this.coordinateSystem=t.coordinateSystem,this.updateCoordinateSystem());const[s,r,a,l,c,d]=this.children,m=t.getRenderTarget(),g=t.getActiveCubeFace(),_=t.getActiveMipmapLevel(),v=t.xr.enabled;t.xr.enabled=!1;const M=n.texture.generateMipmaps;n.texture.generateMipmaps=!1;let C=!1;t.isWebGLRenderer===!0?C=t.state.buffers.depth.getReversed():C=t.reversedDepthBuffer,t.setRenderTarget(n,0,i),C&&t.autoClear===!1&&t.clearDepth(),t.render(e,s),t.setRenderTarget(n,1,i),C&&t.autoClear===!1&&t.clearDepth(),t.render(e,r),t.setRenderTarget(n,2,i),C&&t.autoClear===!1&&t.clearDepth(),t.render(e,a),t.setRenderTarget(n,3,i),C&&t.autoClear===!1&&t.clearDepth(),t.render(e,l),t.setRenderTarget(n,4,i),C&&t.autoClear===!1&&t.clearDepth(),t.render(e,c),n.texture.generateMipmaps=M,t.setRenderTarget(n,5,i),C&&t.autoClear===!1&&t.clearDepth(),t.render(e,d),t.setRenderTarget(m,g,_),t.xr.enabled=v,n.texture.needsPMREMUpdate=!0}}class np extends Xn{constructor(t=[]){super(),this.isArrayCamera=!0,this.isMultiViewCamera=!1,this.cameras=t}}class ip{constructor(){this._previousTime=0,this._currentTime=0,this._startTime=performance.now(),this._delta=0,this._elapsed=0,this._timescale=1,this._document=null,this._pageVisibilityHandler=null}connect(t){this._document=t,t.hidden!==void 0&&(this._pageVisibilityHandler=sp.bind(this),t.addEventListener("visibilitychange",this._pageVisibilityHandler,!1))}disconnect(){this._pageVisibilityHandler!==null&&(this._document.removeEventListener("visibilitychange",this._pageVisibilityHandler),this._pageVisibilityHandler=null),this._document=null}getDelta(){return this._delta/1e3}getElapsed(){return this._elapsed/1e3}getTimescale(){return this._timescale}setTimescale(t){return this._timescale=t,this}reset(){return this._currentTime=performance.now()-this._startTime,this}dispose(){this.disconnect()}update(t){return this._pageVisibilityHandler!==null&&this._document.hidden===!0?this._delta=0:(this._previousTime=this._currentTime,this._currentTime=(t!==void 0?t:performance.now())-this._startTime,this._delta=(this._currentTime-this._previousTime)*this._timescale,this._elapsed+=this._delta),this}}function sp(){this._document.hidden===!1&&this.reset()}const vs=new O,xh=new un,rp=new O,ys=new O,Ms=new O;class Zp extends De{constructor(){super(),this.type="AudioListener",this.context=Qu.getContext(),this.gain=this.context.createGain(),this.gain.connect(this.context.destination),this.filter=null,this.timeDelta=0,this._timer=new ip}getInput(){return this.gain}removeFilter(){return this.filter!==null&&(this.gain.disconnect(this.filter),this.filter.disconnect(this.context.destination),this.gain.connect(this.context.destination),this.filter=null),this}getFilter(){return this.filter}setFilter(t){return this.filter!==null?(this.gain.disconnect(this.filter),this.filter.disconnect(this.context.destination)):this.gain.disconnect(this.context.destination),this.filter=t,this.gain.connect(this.filter),this.filter.connect(this.context.destination),this}getMasterVolume(){return this.gain.gain.value}setMasterVolume(t){return this.gain.gain.setTargetAtTime(t,this.context.currentTime,.01),this}updateMatrixWorld(t){super.updateMatrixWorld(t),this._timer.update();const e=this.context.listener;if(this.timeDelta=this._timer.getDelta(),this.matrixWorld.decompose(vs,xh,rp),ys.set(0,0,-1).applyQuaternion(xh),Ms.set(0,1,0).applyQuaternion(xh),e.positionX){const n=this.context.currentTime+this.timeDelta;e.positionX.linearRampToValueAtTime(vs.x,n),e.positionY.linearRampToValueAtTime(vs.y,n),e.positionZ.linearRampToValueAtTime(vs.z,n),e.forwardX.linearRampToValueAtTime(ys.x,n),e.forwardY.linearRampToValueAtTime(ys.y,n),e.forwardZ.linearRampToValueAtTime(ys.z,n),e.upX.linearRampToValueAtTime(Ms.x,n),e.upY.linearRampToValueAtTime(Ms.y,n),e.upZ.linearRampToValueAtTime(Ms.z,n)}else e.setPosition(vs.x,vs.y,vs.z),e.setOrientation(ys.x,ys.y,ys.z,Ms.x,Ms.y,Ms.z)}}class ap extends De{constructor(t){super(),this.type="Audio",this.listener=t,this.context=t.context,this.gain=this.context.createGain(),this.gain.connect(t.getInput()),this.autoplay=!1,this.buffer=null,this.detune=0,this.loop=!1,this.loopStart=0,this.loopEnd=0,this.offset=0,this.duration=void 0,this.playbackRate=1,this.isPlaying=!1,this.hasPlaybackControl=!0,this.source=null,this.sourceType="empty",this._startedAt=0,this._progress=0,this._connected=!1,this.filters=[]}getOutput(){return this.gain}setNodeSource(t){return this.hasPlaybackControl=!1,this.sourceType="audioNode",this.source=t,this.connect(),this}setMediaElementSource(t){return this.hasPlaybackControl=!1,this.sourceType="mediaNode",this.source=this.context.createMediaElementSource(t),this.connect(),this}setMediaStreamSource(t){return this.hasPlaybackControl=!1,this.sourceType="mediaStreamNode",this.source=this.context.createMediaStreamSource(t),this.connect(),this}setBuffer(t){return this.buffer=t,this.sourceType="buffer",this.autoplay&&this.play(),this}play(t=0){if(this.isPlaying===!0){le("Audio: Audio is already playing.");return}if(this.hasPlaybackControl===!1){le("Audio: this Audio has no playback control.");return}this._startedAt=this.context.currentTime+t;const e=this.context.createBufferSource();return e.buffer=this.buffer,e.loop=this.loop,e.loopStart=this.loopStart,e.loopEnd=this.loopEnd,e.onended=this.onEnded.bind(this),e.start(this._startedAt,this._progress+this.offset,this.duration),this.isPlaying=!0,this.source=e,this.setDetune(this.detune),this.setPlaybackRate(this.playbackRate),this.connect()}pause(){if(this.hasPlaybackControl===!1){le("Audio: this Audio has no playback control.");return}return this.isPlaying===!0&&(this._progress+=Math.max(this.context.currentTime-this._startedAt,0)*this.playbackRate,this.loop===!0&&(this._progress=this._progress%(this.duration||this.buffer.duration)),this.source.stop(),this.source.onended=null,this.isPlaying=!1),this}stop(t=0){if(this.hasPlaybackControl===!1){le("Audio: this Audio has no playback control.");return}return this._progress=0,this.source!==null&&(this.source.stop(this.context.currentTime+t),this.source.onended=null),this.isPlaying=!1,this}connect(){if(this.filters.length>0){this.source.connect(this.filters[0]);for(let t=1,e=this.filters.length;t<e;t++)this.filters[t-1].connect(this.filters[t]);this.filters[this.filters.length-1].connect(this.getOutput())}else this.source.connect(this.getOutput());return this._connected=!0,this}disconnect(){if(this._connected!==!1){if(this.filters.length>0){this.source.disconnect(this.filters[0]);for(let t=1,e=this.filters.length;t<e;t++)this.filters[t-1].disconnect(this.filters[t]);this.filters[this.filters.length-1].disconnect(this.getOutput())}else this.source.disconnect(this.getOutput());return this._connected=!1,this}}getFilters(){return this.filters}setFilters(t){return t||(t=[]),this._connected===!0?(this.disconnect(),this.filters=t.slice(),this.connect()):this.filters=t.slice(),this}setDetune(t){return this.detune=t,this.isPlaying===!0&&this.source.detune!==void 0&&this.source.detune.setTargetAtTime(this.detune,this.context.currentTime,.01),this}getDetune(){return this.detune}getFilter(){return this.getFilters()[0]}setFilter(t){return this.setFilters(t?[t]:[])}setPlaybackRate(t){if(this.hasPlaybackControl===!1){le("Audio: this Audio has no playback control.");return}return this.playbackRate=t,this.isPlaying===!0&&this.source.playbackRate.setTargetAtTime(this.playbackRate,this.context.currentTime,.01),this}getPlaybackRate(){return this.playbackRate}onEnded(){this.isPlaying=!1,this._progress=0}getLoop(){return this.hasPlaybackControl===!1?(le("Audio: this Audio has no playback control."),!1):this.loop}setLoop(t){if(this.hasPlaybackControl===!1){le("Audio: this Audio has no playback control.");return}return this.loop=t,this.isPlaying===!0&&(this.source.loop=this.loop),this}setLoopStart(t){return this.loopStart=t,this}setLoopEnd(t){return this.loopEnd=t,this}getVolume(){return this.gain.gain.value}setVolume(t){return this.gain.gain.setTargetAtTime(t,this.context.currentTime,.01),this}copy(t,e){return super.copy(t,e),t.sourceType!=="buffer"?(le("Audio: Audio source type cannot be copied."),this):(this.autoplay=t.autoplay,this.buffer=t.buffer,this.detune=t.detune,this.loop=t.loop,this.loopStart=t.loopStart,this.loopEnd=t.loopEnd,this.offset=t.offset,this.duration=t.duration,this.playbackRate=t.playbackRate,this.hasPlaybackControl=t.hasPlaybackControl,this.sourceType=t.sourceType,this.filters=t.filters.slice(),this)}clone(t){return new this.constructor(this.listener).copy(this,t)}}const Ss=new O,ef=new un,op=new O,bs=new O;class $p extends ap{constructor(t){super(t),this.panner=this.context.createPanner(),this.panner.panningModel="HRTF",this.panner.connect(this.gain)}connect(){return super.connect(),this.panner.connect(this.gain),this}disconnect(){return super.disconnect(),this.panner.disconnect(this.gain),this}getOutput(){return this.panner}getRefDistance(){return this.panner.refDistance}setRefDistance(t){return this.panner.refDistance=t,this}getRolloffFactor(){return this.panner.rolloffFactor}setRolloffFactor(t){return this.panner.rolloffFactor=t,this}getDistanceModel(){return this.panner.distanceModel}setDistanceModel(t){return this.panner.distanceModel=t,this}getMaxDistance(){return this.panner.maxDistance}setMaxDistance(t){return this.panner.maxDistance=t,this}setDirectionalCone(t,e,n){return this.panner.coneInnerAngle=t,this.panner.coneOuterAngle=e,this.panner.coneOuterGain=n,this}updateMatrixWorld(t){if(super.updateMatrixWorld(t),this.hasPlaybackControl===!0&&this.isPlaying===!1)return;this.matrixWorld.decompose(Ss,ef,op),bs.set(0,0,1).applyQuaternion(ef);const e=this.panner;if(e.positionX){const n=this.context.currentTime+this.listener.timeDelta;e.positionX.linearRampToValueAtTime(Ss.x,n),e.positionY.linearRampToValueAtTime(Ss.y,n),e.positionZ.linearRampToValueAtTime(Ss.z,n),e.orientationX.linearRampToValueAtTime(bs.x,n),e.orientationY.linearRampToValueAtTime(bs.y,n),e.orientationZ.linearRampToValueAtTime(bs.z,n)}else e.setPosition(Ss.x,Ss.y,Ss.z),e.setOrientation(bs.x,bs.y,bs.z)}}class Jp{constructor(t,e=2048){this.analyser=t.context.createAnalyser(),this.analyser.fftSize=e,this.data=new Uint8Array(this.analyser.frequencyBinCount),t.getOutput().connect(this.analyser)}getFrequencyData(){return this.analyser.getByteFrequencyData(this.data),this.data}getAverageFrequency(){let t=0;const e=this.getFrequencyData();for(let n=0;n<e.length;n++)t+=e[n];return t/e.length}}class lp{constructor(t,e,n){this.binding=t,this.valueSize=n;let i,s,r;switch(e){case"quaternion":i=this._slerp,s=this._slerpAdditive,r=this._setAdditiveIdentityQuaternion,this.buffer=new Float64Array(n*6),this._workIndex=5;break;case"string":case"bool":i=this._select,s=this._select,r=this._setAdditiveIdentityOther,this.buffer=new Array(n*5);break;default:i=this._lerp,s=this._lerpAdditive,r=this._setAdditiveIdentityNumeric,this.buffer=new Float64Array(n*5)}this._mixBufferRegion=i,this._mixBufferRegionAdditive=s,this._setIdentity=r,this._origIndex=3,this._addIndex=4,this.cumulativeWeight=0,this.cumulativeWeightAdditive=0,this.useCount=0,this.referenceCount=0}accumulate(t,e){const n=this.buffer,i=this.valueSize,s=t*i+i;let r=this.cumulativeWeight;if(r===0){for(let a=0;a!==i;++a)n[s+a]=n[a];r=e}else{r+=e;const a=e/r;this._mixBufferRegion(n,s,0,a,i)}this.cumulativeWeight=r}accumulateAdditive(t){const e=this.buffer,n=this.valueSize,i=n*this._addIndex;this.cumulativeWeightAdditive===0&&this._setIdentity(),this._mixBufferRegionAdditive(e,i,0,t,n),this.cumulativeWeightAdditive+=t}apply(t){const e=this.valueSize,n=this.buffer,i=t*e+e,s=this.cumulativeWeight,r=this.cumulativeWeightAdditive,a=this.binding;if(this.cumulativeWeight=0,this.cumulativeWeightAdditive=0,s<1){const l=e*this._origIndex;this._mixBufferRegion(n,i,l,1-s,e)}r>0&&this._mixBufferRegionAdditive(n,i,this._addIndex*e,1,e);for(let l=e,c=e+e;l!==c;++l)if(n[l]!==n[l+e]){a.setValue(n,i);break}}saveOriginalState(){const t=this.binding,e=this.buffer,n=this.valueSize,i=n*this._origIndex;t.getValue(e,i);for(let s=n,r=i;s!==r;++s)e[s]=e[i+s%n];this._setIdentity(),this.cumulativeWeight=0,this.cumulativeWeightAdditive=0}restoreOriginalState(){const t=this.valueSize*3;this.binding.setValue(this.buffer,t)}_setAdditiveIdentityNumeric(){const t=this._addIndex*this.valueSize,e=t+this.valueSize;for(let n=t;n<e;n++)this.buffer[n]=0}_setAdditiveIdentityQuaternion(){this._setAdditiveIdentityNumeric(),this.buffer[this._addIndex*this.valueSize+3]=1}_setAdditiveIdentityOther(){const t=this._origIndex*this.valueSize,e=this._addIndex*this.valueSize;for(let n=0;n<this.valueSize;n++)this.buffer[e+n]=this.buffer[t+n]}_select(t,e,n,i,s){if(i>=.5)for(let r=0;r!==s;++r)t[e+r]=t[n+r]}_slerp(t,e,n,i){un.slerpFlat(t,e,t,e,t,n,i)}_slerpAdditive(t,e,n,i,s){const r=this._workIndex*s;un.multiplyQuaternionsFlat(t,r,t,e,t,n),un.slerpFlat(t,e,t,e,t,r,i)}_lerp(t,e,n,i,s){const r=1-i;for(let a=0;a!==s;++a){const l=e+a;t[l]=t[l]*r+t[n+a]*i}}_lerpAdditive(t,e,n,i,s){for(let r=0;r!==s;++r){const a=e+r;t[a]=t[a]+t[n+r]*i}}}const vh="\\[\\]\\.:\\/",cp=new RegExp("["+vh+"]","g"),yh="[^"+vh+"]",hp="[^"+vh.replace("\\.","")+"]",up=/((?:WC+[\/:])*)/.source.replace("WC",yh),fp=/(WCOD+)?/.source.replace("WCOD",hp),dp=/(?:\.(WC+)(?:\[(.+)\])?)?/.source.replace("WC",yh),pp=/\.(WC+)(?:\[(.+)\])?/.source.replace("WC",yh),mp=new RegExp("^"+up+fp+dp+pp+"$"),gp=["material","materials","bones","map"];class _p{constructor(t,e,n){const i=n||Fe.parseTrackName(e);this._targetGroup=t,this._bindings=t.subscribe_(e,i)}getValue(t,e){this.bind();const n=this._targetGroup.nCachedObjects_,i=this._bindings[n];i!==void 0&&i.getValue(t,e)}setValue(t,e){const n=this._bindings;for(let i=this._targetGroup.nCachedObjects_,s=n.length;i!==s;++i)n[i].setValue(t,e)}bind(){const t=this._bindings;for(let e=this._targetGroup.nCachedObjects_,n=t.length;e!==n;++e)t[e].bind()}unbind(){const t=this._bindings;for(let e=this._targetGroup.nCachedObjects_,n=t.length;e!==n;++e)t[e].unbind()}}class Fe{constructor(t,e,n){this.path=e,this.parsedPath=n||Fe.parseTrackName(e),this.node=Fe.findNode(t,this.parsedPath.nodeName),this.rootNode=t,this.getValue=this._getValue_unbound,this.setValue=this._setValue_unbound}static create(t,e,n){return t&&t.isAnimationObjectGroup?new Fe.Composite(t,e,n):new Fe(t,e,n)}static sanitizeNodeName(t){return t.replace(/\s/g,"_").replace(cp,"")}static parseTrackName(t){const e=mp.exec(t);if(e===null)throw new Error("PropertyBinding: Cannot parse trackName: "+t);const n={nodeName:e[2],objectName:e[3],objectIndex:e[4],propertyName:e[5],propertyIndex:e[6]},i=n.nodeName&&n.nodeName.lastIndexOf(".");if(i!==void 0&&i!==-1){const s=n.nodeName.substring(i+1);gp.indexOf(s)!==-1&&(n.nodeName=n.nodeName.substring(0,i),n.objectName=s)}if(n.propertyName===null||n.propertyName.length===0)throw new Error("PropertyBinding: can not parse propertyName from trackName: "+t);return n}static findNode(t,e){if(e===void 0||e===""||e==="."||e===-1||e===t.name||e===t.uuid)return t;if(t.skeleton){const n=t.skeleton.getBoneByName(e);if(n!==void 0)return n}if(t.children){const n=function(s){for(let r=0;r<s.length;r++){const a=s[r];if(a.name===e||a.uuid===e)return a;const l=n(a.children);if(l)return l}return null},i=n(t.children);if(i)return i}return null}_getValue_unavailable(){}_setValue_unavailable(){}_getValue_direct(t,e){t[e]=this.targetObject[this.propertyName]}_getValue_array(t,e){const n=this.resolvedProperty;for(let i=0,s=n.length;i!==s;++i)t[e++]=n[i]}_getValue_arrayElement(t,e){t[e]=this.resolvedProperty[this.propertyIndex]}_getValue_toArray(t,e){this.resolvedProperty.toArray(t,e)}_setValue_direct(t,e){this.targetObject[this.propertyName]=t[e]}_setValue_direct_setNeedsUpdate(t,e){this.targetObject[this.propertyName]=t[e],this.targetObject.needsUpdate=!0}_setValue_direct_setMatrixWorldNeedsUpdate(t,e){this.targetObject[this.propertyName]=t[e],this.targetObject.matrixWorldNeedsUpdate=!0}_setValue_array(t,e){const n=this.resolvedProperty;for(let i=0,s=n.length;i!==s;++i)n[i]=t[e++]}_setValue_array_setNeedsUpdate(t,e){const n=this.resolvedProperty;for(let i=0,s=n.length;i!==s;++i)n[i]=t[e++];this.targetObject.needsUpdate=!0}_setValue_array_setMatrixWorldNeedsUpdate(t,e){const n=this.resolvedProperty;for(let i=0,s=n.length;i!==s;++i)n[i]=t[e++];this.targetObject.matrixWorldNeedsUpdate=!0}_setValue_arrayElement(t,e){this.resolvedProperty[this.propertyIndex]=t[e]}_setValue_arrayElement_setNeedsUpdate(t,e){this.resolvedProperty[this.propertyIndex]=t[e],this.targetObject.needsUpdate=!0}_setValue_arrayElement_setMatrixWorldNeedsUpdate(t,e){this.resolvedProperty[this.propertyIndex]=t[e],this.targetObject.matrixWorldNeedsUpdate=!0}_setValue_fromArray(t,e){this.resolvedProperty.fromArray(t,e)}_setValue_fromArray_setNeedsUpdate(t,e){this.resolvedProperty.fromArray(t,e),this.targetObject.needsUpdate=!0}_setValue_fromArray_setMatrixWorldNeedsUpdate(t,e){this.resolvedProperty.fromArray(t,e),this.targetObject.matrixWorldNeedsUpdate=!0}_getValue_unbound(t,e){this.bind(),this.getValue(t,e)}_setValue_unbound(t,e){this.bind(),this.setValue(t,e)}bind(){let t=this.node;const e=this.parsedPath,n=e.objectName,i=e.propertyName;let s=e.propertyIndex;if(t||(t=Fe.findNode(this.rootNode,e.nodeName),this.node=t),this.getValue=this._getValue_unavailable,this.setValue=this._setValue_unavailable,!t){le("PropertyBinding: No target node found for track: "+this.path+".");return}if(n){let c=e.objectIndex;switch(n){case"materials":if(!t.material){Re("PropertyBinding: Can not bind to material as node does not have a material.",this);return}if(!t.material.materials){Re("PropertyBinding: Can not bind to material.materials as node.material does not have a materials array.",this);return}t=t.material.materials;break;case"bones":if(!t.skeleton){Re("PropertyBinding: Can not bind to bones as node does not have a skeleton.",this);return}t=t.skeleton.bones;for(let d=0;d<t.length;d++)if(t[d].name===c){c=d;break}break;case"map":if("map"in t){t=t.map;break}if(!t.material){Re("PropertyBinding: Can not bind to material as node does not have a material.",this);return}if(!t.material.map){Re("PropertyBinding: Can not bind to material.map as node.material does not have a map.",this);return}t=t.material.map;break;default:if(t[n]===void 0){Re("PropertyBinding: Can not bind to objectName of node undefined.",this);return}t=t[n]}if(c!==void 0){if(t[c]===void 0){Re("PropertyBinding: Trying to bind to objectIndex of objectName, but is undefined.",this,t);return}t=t[c]}}const r=t[i];if(r===void 0){const c=e.nodeName;Re("PropertyBinding: Trying to update property for track: "+c+"."+i+" but it wasn't found.",t);return}let a=this.Versioning.None;this.targetObject=t,t.isMaterial===!0?a=this.Versioning.NeedsUpdate:t.isObject3D===!0&&(a=this.Versioning.MatrixWorldNeedsUpdate);let l=this.BindingType.Direct;if(s!==void 0){if(i==="morphTargetInfluences"){if(!t.geometry){Re("PropertyBinding: Can not bind to morphTargetInfluences because node does not have a geometry.",this);return}if(!t.geometry.morphAttributes){Re("PropertyBinding: Can not bind to morphTargetInfluences because node does not have a geometry.morphAttributes.",this);return}t.morphTargetDictionary[s]!==void 0&&(s=t.morphTargetDictionary[s])}l=this.BindingType.ArrayElement,this.resolvedProperty=r,this.propertyIndex=s}else r.fromArray!==void 0&&r.toArray!==void 0?(l=this.BindingType.HasFromToArray,this.resolvedProperty=r):Array.isArray(r)?(l=this.BindingType.EntireArray,this.resolvedProperty=r):this.propertyName=i;this.getValue=this.GetterByBindingType[l],this.setValue=this.SetterByBindingTypeAndVersioning[l][a]}unbind(){this.node=null,this.getValue=this._getValue_unbound,this.setValue=this._setValue_unbound}}Fe.Composite=_p,Fe.prototype.BindingType={Direct:0,EntireArray:1,ArrayElement:2,HasFromToArray:3},Fe.prototype.Versioning={None:0,NeedsUpdate:1,MatrixWorldNeedsUpdate:2},Fe.prototype.GetterByBindingType=[Fe.prototype._getValue_direct,Fe.prototype._getValue_array,Fe.prototype._getValue_arrayElement,Fe.prototype._getValue_toArray],Fe.prototype.SetterByBindingTypeAndVersioning=[[Fe.prototype._setValue_direct,Fe.prototype._setValue_direct_setNeedsUpdate,Fe.prototype._setValue_direct_setMatrixWorldNeedsUpdate],[Fe.prototype._setValue_array,Fe.prototype._setValue_array_setNeedsUpdate,Fe.prototype._setValue_array_setMatrixWorldNeedsUpdate],[Fe.prototype._setValue_arrayElement,Fe.prototype._setValue_arrayElement_setNeedsUpdate,Fe.prototype._setValue_arrayElement_setMatrixWorldNeedsUpdate],[Fe.prototype._setValue_fromArray,Fe.prototype._setValue_fromArray_setNeedsUpdate,Fe.prototype._setValue_fromArray_setMatrixWorldNeedsUpdate]];class Kp{constructor(){this.isAnimationObjectGroup=!0,this.uuid=An(),this._objects=Array.prototype.slice.call(arguments),this.nCachedObjects_=0;const t={};this._indicesByUUID=t;for(let n=0,i=arguments.length;n!==i;++n)t[arguments[n].uuid]=n;this._paths=[],this._parsedPaths=[],this._bindings=[],this._bindingsIndicesByPath={};const e=this;this.stats={objects:{get total(){return e._objects.length},get inUse(){return this.total-e.nCachedObjects_}},get bindingsPerObject(){return e._bindings.length}}}add(){const t=this._objects,e=this._indicesByUUID,n=this._paths,i=this._parsedPaths,s=this._bindings,r=s.length;let a,l=t.length,c=this.nCachedObjects_;for(let d=0,m=arguments.length;d!==m;++d){const g=arguments[d],_=g.uuid;let v=e[_];if(v===void 0){v=l++,e[_]=v,t.push(g);for(let M=0,C=r;M!==C;++M)s[M].push(new Fe(g,n[M],i[M]))}else if(v<c){a=t[v];const M=--c,C=t[M];e[C.uuid]=v,t[v]=C,e[_]=M,t[M]=g;for(let w=0,L=r;w!==L;++w){const D=s[w],N=D[M];let et=D[v];D[v]=N,et===void 0&&(et=new Fe(g,n[w],i[w])),D[M]=et}}else t[v]!==a&&Re("AnimationObjectGroup: Different objects with the same UUID detected. Clean the caches or recreate your infrastructure when reloading scenes.")}this.nCachedObjects_=c}remove(){const t=this._objects,e=this._indicesByUUID,n=this._bindings,i=n.length;let s=this.nCachedObjects_;for(let r=0,a=arguments.length;r!==a;++r){const l=arguments[r],c=l.uuid,d=e[c];if(d!==void 0&&d>=s){const m=s++,g=t[m];e[g.uuid]=d,t[d]=g,e[c]=m,t[m]=l;for(let _=0,v=i;_!==v;++_){const M=n[_],C=M[m],w=M[d];M[d]=C,M[m]=w}}}this.nCachedObjects_=s}uncache(){const t=this._objects,e=this._indicesByUUID,n=this._bindings,i=n.length;let s=this.nCachedObjects_,r=t.length;for(let a=0,l=arguments.length;a!==l;++a){const c=arguments[a],d=c.uuid,m=e[d];if(m!==void 0)if(delete e[d],m<s){const g=--s,_=t[g],v=--r,M=t[v];e[_.uuid]=m,t[m]=_,e[M.uuid]=g,t[g]=M,t.pop();for(let C=0,w=i;C!==w;++C){const L=n[C],D=L[g],N=L[v];L[m]=D,L[g]=N,L.pop()}}else{const g=--r,_=t[g];g>0&&(e[_.uuid]=m),t[m]=_,t.pop();for(let v=0,M=i;v!==M;++v){const C=n[v];C[m]=C[g],C.pop()}}}this.nCachedObjects_=s}subscribe_(t,e){const n=this._bindingsIndicesByPath;let i=n[t];const s=this._bindings;if(i!==void 0)return s[i];const r=this._paths,a=this._parsedPaths,l=this._objects,c=l.length,d=this.nCachedObjects_,m=new Array(c);i=s.length,n[t]=i,r.push(t),a.push(e),s.push(m);for(let g=d,_=l.length;g!==_;++g){const v=l[g];m[g]=new Fe(v,t,e)}return m}unsubscribe_(t){const e=this._bindingsIndicesByPath,n=e[t];if(n!==void 0){const i=this._paths,s=this._parsedPaths,r=this._bindings,a=r.length-1,l=r[a],c=t[a];e[c]=n,r[n]=l,r.pop(),s[n]=s[a],s.pop(),i[n]=i[a],i.pop()}}}class xp{constructor(t,e,n=null,i=e.blendMode){this._mixer=t,this._clip=e,this._localRoot=n,this.blendMode=i;const s=e.tracks,r=s.length,a=new Array(r),l={endingStart:Jn,endingEnd:Jn};for(let c=0;c!==r;++c){const d=s[c].createInterpolant(null);a[c]=d,d.settings=l}this._interpolantSettings=l,this._interpolants=a,this._propertyBindings=new Array(r),this._cacheIndex=null,this._byClipCacheIndex=null,this._timeScaleInterpolant=null,this._weightInterpolant=null,this.loop=Uh,this._loopCount=-1,this._startTime=null,this.time=0,this.timeScale=1,this._effectiveTimeScale=1,this.weight=1,this._effectiveWeight=1,this.repetitions=1/0,this.paused=!1,this.enabled=!0,this.clampWhenFinished=!1,this.zeroSlopeAtStart=!0,this.zeroSlopeAtEnd=!0}play(){return this._mixer._activateAction(this),this}stop(){return this._mixer._deactivateAction(this),this.reset()}reset(){return this.paused=!1,this.enabled=!0,this.time=0,this._loopCount=-1,this._startTime=null,this.stopFading().stopWarping()}isRunning(){return this.enabled&&!this.paused&&this.timeScale!==0&&this._startTime===null&&this._mixer._isActiveAction(this)}isScheduled(){return this._mixer._isActiveAction(this)}startAt(t){return this._startTime=t,this}setLoop(t,e){return this.loop=t,this.repetitions=e,this}setEffectiveWeight(t){return this.weight=t,this._effectiveWeight=this.enabled?t:0,this.stopFading()}getEffectiveWeight(){return this._effectiveWeight}fadeIn(t){return this._scheduleFading(t,0,1)}fadeOut(t){return this._scheduleFading(t,1,0)}crossFadeFrom(t,e,n=!1){if(t.fadeOut(e),this.fadeIn(e),n===!0){const i=this._clip.duration,s=t._clip.duration,r=s/i,a=i/s;t.warp(1,r,e),this.warp(a,1,e)}return this}crossFadeTo(t,e,n=!1){return t.crossFadeFrom(this,e,n)}stopFading(){const t=this._weightInterpolant;return t!==null&&(this._weightInterpolant=null,this._mixer._takeBackControlInterpolant(t)),this}setEffectiveTimeScale(t){return this.timeScale=t,this._effectiveTimeScale=this.paused?0:t,this.stopWarping()}getEffectiveTimeScale(){return this._effectiveTimeScale}setDuration(t){return this.timeScale=this._clip.duration/t,this.stopWarping()}syncWith(t){return this.time=t.time,this.timeScale=t.timeScale,this.stopWarping()}halt(t){return this.warp(this._effectiveTimeScale,0,t)}warp(t,e,n){const i=this._mixer,s=i.time,r=this.timeScale;let a=this._timeScaleInterpolant;a===null&&(a=i._lendControlInterpolant(),this._timeScaleInterpolant=a);const l=a.parameterPositions,c=a.sampleValues;return l[0]=s,l[1]=s+n,c[0]=t/r,c[1]=e/r,this}stopWarping(){const t=this._timeScaleInterpolant;return t!==null&&(this._timeScaleInterpolant=null,this._mixer._takeBackControlInterpolant(t)),this}getMixer(){return this._mixer}getClip(){return this._clip}getRoot(){return this._localRoot||this._mixer._root}_update(t,e,n,i){if(!this.enabled){this._updateWeight(t);return}const s=this._startTime;if(s!==null){const l=(t-s)*n;l<0||n===0?e=0:(this._startTime=null,e=n*l)}e*=this._updateTimeScale(t);const r=this._updateTime(e),a=this._updateWeight(t);if(a>0){const l=this._interpolants,c=this._propertyBindings;switch(this.blendMode){case ba:for(let d=0,m=l.length;d!==m;++d)l[d].evaluate(r),c[d].accumulateAdditive(a);break;case fr:default:for(let d=0,m=l.length;d!==m;++d)l[d].evaluate(r),c[d].accumulate(i,a)}}}_updateWeight(t){let e=0;if(this.enabled){e=this.weight;const n=this._weightInterpolant;if(n!==null){const i=n.evaluate(t)[0];e*=i,t>n.parameterPositions[1]&&(this.stopFading(),i===0&&(this.enabled=!1))}}return this._effectiveWeight=e,e}_updateTimeScale(t){let e=0;if(!this.paused){e=this.timeScale;const n=this._timeScaleInterpolant;if(n!==null){const i=n.evaluate(t)[0];e*=i,t>n.parameterPositions[1]&&(this.stopWarping(),e===0?this.paused=!0:this.timeScale=e)}}return this._effectiveTimeScale=e,e}_updateTime(t){const e=this._clip.duration,n=this.loop;let i=this.time+t,s=this._loopCount;const r=n===Nh;if(t===0)return s===-1?i:r&&(s&1)===1?e-i:i;if(n===Dh){s===-1&&(this._loopCount=0,this._setEndings(!0,!0,!1));t:{if(i>=e)i=e;else if(i<0)i=0;else{this.time=i;break t}this.clampWhenFinished?this.paused=!0:this.enabled=!1,this.time=i,this._mixer.dispatchEvent({type:"finished",action:this,direction:t<0?-1:1})}}else{if(s===-1&&(t>=0?(s=0,this._setEndings(!0,this.repetitions===0,r)):this._setEndings(this.repetitions===0,!0,r)),i>=e||i<0){const a=Math.floor(i/e);i-=e*a,s+=Math.abs(a);const l=this.repetitions-s;if(l<=0)this.clampWhenFinished?this.paused=!0:this.enabled=!1,i=t>0?e:0,this.time=i,this._mixer.dispatchEvent({type:"finished",action:this,direction:t>0?1:-1});else{if(l===1){const c=t<0;this._setEndings(c,!c,r)}else this._setEndings(!1,!1,r);this._loopCount=s,this.time=i,this._mixer.dispatchEvent({type:"loop",action:this,loopDelta:a})}}else this.time=i;if(r&&(s&1)===1)return e-i}return i}_setEndings(t,e,n){const i=this._interpolantSettings;n?(i.endingStart=Rn,i.endingEnd=Rn):(t?i.endingStart=this.zeroSlopeAtStart?Rn:Jn:i.endingStart=Ls,e?i.endingEnd=this.zeroSlopeAtEnd?Rn:Jn:i.endingEnd=Ls)}_scheduleFading(t,e,n){const i=this._mixer,s=i.time;let r=this._weightInterpolant;r===null&&(r=i._lendControlInterpolant(),this._weightInterpolant=r);const a=r.parameterPositions,l=r.sampleValues;return a[0]=s,l[0]=e,a[1]=s+t,l[1]=n,this}}const vp=new Float32Array(1);class Qp extends Kn{constructor(t){super(),this._root=t,this._initMemoryManager(),this._accuIndex=0,this.time=0,this.timeScale=1,typeof __THREE_DEVTOOLS__<"u"&&__THREE_DEVTOOLS__.dispatchEvent(new CustomEvent("observe",{detail:this}))}_bindAction(t,e){const n=t._localRoot||this._root,i=t._clip.tracks,s=i.length,r=t._propertyBindings,a=t._interpolants,l=n.uuid,c=this._bindingsByRootAndName;let d=c[l];d===void 0&&(d={},c[l]=d);for(let m=0;m!==s;++m){const g=i[m],_=g.name;let v=d[_];if(v!==void 0)++v.referenceCount,r[m]=v;else{if(v=r[m],v!==void 0){v._cacheIndex===null&&(++v.referenceCount,this._addInactiveBinding(v,l,_));continue}const M=e&&e._propertyBindings[m].binding.parsedPath;v=new lp(Fe.create(n,_,M),g.ValueTypeName,g.getValueSize()),++v.referenceCount,this._addInactiveBinding(v,l,_),r[m]=v}a[m].resultBuffer=v.buffer}}_activateAction(t){if(!this._isActiveAction(t)){if(t._cacheIndex===null){const n=(t._localRoot||this._root).uuid,i=t._clip.uuid,s=this._actionsByClip[i];this._bindAction(t,s&&s.knownActions[0]),this._addInactiveAction(t,i,n)}const e=t._propertyBindings;for(let n=0,i=e.length;n!==i;++n){const s=e[n];s.useCount++===0&&(this._lendBinding(s),s.saveOriginalState())}this._lendAction(t)}}_deactivateAction(t){if(this._isActiveAction(t)){const e=t._propertyBindings;for(let n=0,i=e.length;n!==i;++n){const s=e[n];--s.useCount===0&&(s.restoreOriginalState(),this._takeBackBinding(s))}this._takeBackAction(t)}}_initMemoryManager(){this._actions=[],this._nActiveActions=0,this._actionsByClip={},this._bindings=[],this._nActiveBindings=0,this._bindingsByRootAndName={},this._controlInterpolants=[],this._nActiveControlInterpolants=0;const t=this;this.stats={actions:{get total(){return t._actions.length},get inUse(){return t._nActiveActions}},bindings:{get total(){return t._bindings.length},get inUse(){return t._nActiveBindings}},controlInterpolants:{get total(){return t._controlInterpolants.length},get inUse(){return t._nActiveControlInterpolants}}}}_isActiveAction(t){const e=t._cacheIndex;return e!==null&&e<this._nActiveActions}_addInactiveAction(t,e,n){const i=this._actions,s=this._actionsByClip;let r=s[e];if(r===void 0)r={knownActions:[t],actionByRoot:{}},t._byClipCacheIndex=0,s[e]=r;else{const a=r.knownActions;t._byClipCacheIndex=a.length,a.push(t)}t._cacheIndex=i.length,i.push(t),r.actionByRoot[n]=t}_removeInactiveAction(t){const e=this._actions,n=e[e.length-1],i=t._cacheIndex;n._cacheIndex=i,e[i]=n,e.pop(),t._cacheIndex=null;const s=t._clip.uuid,r=this._actionsByClip,a=r[s],l=a.knownActions,c=l[l.length-1],d=t._byClipCacheIndex;c._byClipCacheIndex=d,l[d]=c,l.pop(),t._byClipCacheIndex=null;const m=a.actionByRoot,g=(t._localRoot||this._root).uuid;delete m[g],l.length===0&&delete r[s],this._removeInactiveBindingsForAction(t)}_removeInactiveBindingsForAction(t){const e=t._propertyBindings;for(let n=0,i=e.length;n!==i;++n){const s=e[n];--s.referenceCount===0&&this._removeInactiveBinding(s)}}_lendAction(t){const e=this._actions,n=t._cacheIndex,i=this._nActiveActions++,s=e[i];t._cacheIndex=i,e[i]=t,s._cacheIndex=n,e[n]=s}_takeBackAction(t){const e=this._actions,n=t._cacheIndex,i=--this._nActiveActions,s=e[i];t._cacheIndex=i,e[i]=t,s._cacheIndex=n,e[n]=s}_addInactiveBinding(t,e,n){const i=this._bindingsByRootAndName,s=this._bindings;let r=i[e];r===void 0&&(r={},i[e]=r),r[n]=t,t._cacheIndex=s.length,s.push(t)}_removeInactiveBinding(t){const e=this._bindings,n=t.binding,i=n.rootNode.uuid,s=n.path,r=this._bindingsByRootAndName,a=r[i],l=e[e.length-1],c=t._cacheIndex;l._cacheIndex=c,e[c]=l,e.pop(),delete a[s],Object.keys(a).length===0&&delete r[i]}_lendBinding(t){const e=this._bindings,n=t._cacheIndex,i=this._nActiveBindings++,s=e[i];t._cacheIndex=i,e[i]=t,s._cacheIndex=n,e[n]=s}_takeBackBinding(t){const e=this._bindings,n=t._cacheIndex,i=--this._nActiveBindings,s=e[i];t._cacheIndex=i,e[i]=t,s._cacheIndex=n,e[n]=s}_lendControlInterpolant(){const t=this._controlInterpolants,e=this._nActiveControlInterpolants++;let n=t[e];return n===void 0&&(n=new Vu(new Float32Array(2),new Float32Array(2),1,vp),n.__cacheIndex=e,t[e]=n),n}_takeBackControlInterpolant(t){const e=this._controlInterpolants,n=t.__cacheIndex,i=--this._nActiveControlInterpolants,s=e[i];t.__cacheIndex=i,e[i]=t,s.__cacheIndex=n,e[n]=s}clipAction(t,e,n){const i=e||this._root,s=i.uuid;let r=typeof t=="string"?To.findByName(i,t):t;const a=r!==null?r.uuid:t,l=this._actionsByClip[a];let c=null;if(n===void 0&&(r!==null?n=r.blendMode:n=fr),l!==void 0){const m=l.actionByRoot[s];if(m!==void 0&&m.blendMode===n)return m;c=l.knownActions[0],r===null&&(r=c._clip)}if(r===null)return null;const d=new xp(this,r,e,n);return this._bindAction(d,c),this._addInactiveAction(d,a,s),d}existingAction(t,e){const n=e||this._root,i=n.uuid,s=typeof t=="string"?To.findByName(n,t):t,r=s?s.uuid:t,a=this._actionsByClip[r];return a!==void 0&&a.actionByRoot[i]||null}stopAllAction(){const t=this._actions,e=this._nActiveActions;for(let n=e-1;n>=0;--n)t[n].stop();return this}update(t){t*=this.timeScale;const e=this._actions,n=this._nActiveActions,i=this.time+=t,s=Math.sign(t),r=this._accuIndex^=1;for(let c=0;c!==n;++c)e[c]._update(i,t,s,r);const a=this._bindings,l=this._nActiveBindings;for(let c=0;c!==l;++c)a[c].apply(r);return this}setTime(t){this.time=0;for(let e=0;e<this._actions.length;e++)this._actions[e].time=0;return this.update(t)}getRoot(){return this._root}uncacheClip(t){const e=this._actions,n=t.uuid,i=this._actionsByClip,s=i[n];if(s!==void 0){const r=s.knownActions;for(let a=0,l=r.length;a!==l;++a){const c=r[a];this._deactivateAction(c);const d=c._cacheIndex,m=e[e.length-1];c._cacheIndex=null,c._byClipCacheIndex=null,m._cacheIndex=d,e[d]=m,e.pop(),this._removeInactiveBindingsForAction(c)}delete i[n]}}uncacheRoot(t){const e=t.uuid,n=this._actionsByClip;for(const r in n){const a=n[r].actionByRoot,l=a[e];l!==void 0&&(this._deactivateAction(l),this._removeInactiveAction(l))}const i=this._bindingsByRootAndName,s=i[e];if(s!==void 0)for(const r in s){const a=s[r];a.restoreOriginalState(),this._removeInactiveBinding(a)}}uncacheAction(t,e){const n=this.existingAction(t,e);n!==null&&(this._deactivateAction(n),this._removeInactiveAction(n))}}class jp extends Ga{constructor(t=1,e=1,n=1,i={}){super(t,e,i),this.isRenderTarget3D=!0,this.depth=n,this.texture=new Er(null,t,e,n),this._setTextureOptions(i),this.texture.isRenderTargetTexture=!0}}class nf{constructor(t){this.value=t}clone(){return new nf(this.value.clone===void 0?this.value:this.value.clone())}}let yp=0;class tm extends Kn{constructor(){super(),this.isUniformsGroup=!0,Object.defineProperty(this,"id",{value:yp++}),this.name="",this.usage=ji,this.uniforms=[]}add(t){return this.uniforms.push(t),this}remove(t){const e=this.uniforms.indexOf(t);return e!==-1&&this.uniforms.splice(e,1),this}setName(t){return this.name=t,this}setUsage(t){return this.usage=t,this}dispose(){this.dispatchEvent({type:"dispose"})}copy(t){this.name=t.name,this.usage=t.usage;const e=t.uniforms;this.uniforms.length=0;for(let n=0,i=e.length;n<i;n++){const s=Array.isArray(e[n])?e[n]:[e[n]];for(let r=0;r<s.length;r++)this.uniforms.push(s[r].clone())}return this}clone(){return new this.constructor().copy(this)}}class em extends W{constructor(t,e,n=1){super(t,e),this.isInstancedInterleavedBuffer=!0,this.meshPerAttribute=n}copy(t){return super.copy(t),this.meshPerAttribute=t.meshPerAttribute,this}clone(t){const e=super.clone(t);return e.meshPerAttribute=this.meshPerAttribute,e}toJSON(t){const e=super.toJSON(t);return e.isInstancedInterleavedBuffer=!0,e.meshPerAttribute=this.meshPerAttribute,e}}class nm{constructor(t,e,n,i,s,r=!1){this.isGLBufferAttribute=!0,this.name="",this.buffer=t,this.type=e,this.itemSize=n,this.elementSize=i,this.count=s,this.normalized=r,this.version=0}set needsUpdate(t){t===!0&&this.version++}setBuffer(t){return this.buffer=t,this}setType(t,e){return this.type=t,this.elementSize=e,this}setItemSize(t){return this.itemSize=t,this}setCount(t){return this.count=t,this}}const sf=new Me;class im{constructor(t,e,n=0,i=1/0){this.ray=new Vn(t,e),this.near=n,this.far=i,this.camera=null,this.layers=new wr,this.params={Mesh:{},Line:{threshold:1},LOD:{},Points:{threshold:1},Sprite:{}}}set(t,e){this.ray.set(t,e)}setFromCamera(t,e){e.isPerspectiveCamera?(this.ray.origin.setFromMatrixPosition(e.matrixWorld),this.ray.direction.set(t.x,t.y,.5).unproject(e).sub(this.ray.origin).normalize(),this.camera=e):e.isOrthographicCamera?(this.ray.origin.set(t.x,t.y,(e.near+e.far)/(e.near-e.far)).unproject(e),this.ray.direction.set(0,0,-1).transformDirection(e.matrixWorld),this.camera=e):Re("Raycaster: Unsupported camera type: "+e.type)}setFromXRController(t){return sf.identity().extractRotation(t.matrixWorld),this.ray.origin.setFromMatrixPosition(t.matrixWorld),this.ray.direction.set(0,0,-1).applyMatrix4(sf),this}intersectObject(t,e=!0,n=[]){return Mh(t,this,n,e),n.sort(rf),n}intersectObjects(t,e=!0,n=[]){for(let i=0,s=t.length;i<s;i++)Mh(t[i],this,n,e);return n.sort(rf),n}}function rf(u,t){return u.distance-t.distance}function Mh(u,t,e,n){let i=!0;if(u.layers.test(t.layers)&&u.raycast(t,e)===!1&&(i=!1),i===!0&&n===!0){const s=u.children;for(let r=0,a=s.length;r<a;r++)Mh(s[r],t,e,!0)}}class sm{constructor(t=!0){this.autoStart=t,this.startTime=0,this.oldTime=0,this.elapsedTime=0,this.running=!1,le("THREE.Clock: This module has been deprecated. Please use THREE.Timer instead.")}start(){this.startTime=performance.now(),this.oldTime=this.startTime,this.elapsedTime=0,this.running=!0}stop(){this.getElapsedTime(),this.running=!1,this.autoStart=!1}getElapsedTime(){return this.getDelta(),this.elapsedTime}getDelta(){let t=0;if(this.autoStart&&!this.running)return this.start(),0;if(this.running){const e=performance.now();t=(e-this.oldTime)/1e3,this.oldTime=e,this.elapsedTime+=t}return t}}class rm{constructor(t=1,e=0,n=0){this.radius=t,this.phi=e,this.theta=n}set(t,e,n){return this.radius=t,this.phi=e,this.theta=n,this}copy(t){return this.radius=t.radius,this.phi=t.phi,this.theta=t.theta,this}makeSafe(){return this.phi=pe(this.phi,1e-6,Math.PI-1e-6),this}setFromVector3(t){return this.setFromCartesianCoords(t.x,t.y,t.z)}setFromCartesianCoords(t,e,n){return this.radius=Math.sqrt(t*t+e*e+n*n),this.radius===0?(this.theta=0,this.phi=0):(this.theta=Math.atan2(t,n),this.phi=Math.acos(pe(e/this.radius,-1,1))),this}clone(){return new this.constructor().copy(this)}}class am{constructor(t=1,e=0,n=0){this.radius=t,this.theta=e,this.y=n}set(t,e,n){return this.radius=t,this.theta=e,this.y=n,this}copy(t){return this.radius=t.radius,this.theta=t.theta,this.y=t.y,this}setFromVector3(t){return this.setFromCartesianCoords(t.x,t.y,t.z)}setFromCartesianCoords(t,e,n){return this.radius=Math.sqrt(t*t+n*n),this.theta=Math.atan2(t,n),this.y=e,this}clone(){return new this.constructor().copy(this)}}class af{constructor(t,e,n,i){af.prototype.isMatrix2=!0,this.elements=[1,0,0,1],t!==void 0&&this.set(t,e,n,i)}identity(){return this.set(1,0,0,1),this}fromArray(t,e=0){for(let n=0;n<4;n++)this.elements[n]=t[n+e];return this}set(t,e,n,i){const s=this.elements;return s[0]=t,s[2]=e,s[1]=n,s[3]=i,this}}const of=new At;class om{constructor(t=new At(1/0,1/0),e=new At(-1/0,-1/0)){this.isBox2=!0,this.min=t,this.max=e}set(t,e){return this.min.copy(t),this.max.copy(e),this}setFromPoints(t){this.makeEmpty();for(let e=0,n=t.length;e<n;e++)this.expandByPoint(t[e]);return this}setFromCenterAndSize(t,e){const n=of.copy(e).multiplyScalar(.5);return this.min.copy(t).sub(n),this.max.copy(t).add(n),this}clone(){return new this.constructor().copy(this)}copy(t){return this.min.copy(t.min),this.max.copy(t.max),this}makeEmpty(){return this.min.x=this.min.y=1/0,this.max.x=this.max.y=-1/0,this}isEmpty(){return this.max.x<this.min.x||this.max.y<this.min.y}getCenter(t){return this.isEmpty()?t.set(0,0):t.addVectors(this.min,this.max).multiplyScalar(.5)}getSize(t){return this.isEmpty()?t.set(0,0):t.subVectors(this.max,this.min)}expandByPoint(t){return this.min.min(t),this.max.max(t),this}expandByVector(t){return this.min.sub(t),this.max.add(t),this}expandByScalar(t){return this.min.addScalar(-t),this.max.addScalar(t),this}containsPoint(t){return t.x>=this.min.x&&t.x<=this.max.x&&t.y>=this.min.y&&t.y<=this.max.y}containsBox(t){return this.min.x<=t.min.x&&t.max.x<=this.max.x&&this.min.y<=t.min.y&&t.max.y<=this.max.y}getParameter(t,e){return e.set((t.x-this.min.x)/(this.max.x-this.min.x),(t.y-this.min.y)/(this.max.y-this.min.y))}intersectsBox(t){return t.max.x>=this.min.x&&t.min.x<=this.max.x&&t.max.y>=this.min.y&&t.min.y<=this.max.y}clampPoint(t,e){return e.copy(t).clamp(this.min,this.max)}distanceToPoint(t){return this.clampPoint(t,of).distanceTo(t)}intersect(t){return this.min.max(t.min),this.max.min(t.max),this.isEmpty()&&this.makeEmpty(),this}union(t){return this.min.min(t.min),this.max.max(t.max),this}translate(t){return this.min.add(t),this.max.add(t),this}equals(t){return t.min.equals(this.min)&&t.max.equals(this.max)}}const lf=new O,Ro=new O,Zs=new O,$s=new O,Sh=new O,Mp=new O,Sp=new O;class lm{constructor(t=new O,e=new O){this.start=t,this.end=e}set(t,e){return this.start.copy(t),this.end.copy(e),this}copy(t){return this.start.copy(t.start),this.end.copy(t.end),this}getCenter(t){return t.addVectors(this.start,this.end).multiplyScalar(.5)}delta(t){return t.subVectors(this.end,this.start)}distanceSq(){return this.start.distanceToSquared(this.end)}distance(){return this.start.distanceTo(this.end)}at(t,e){return this.delta(e).multiplyScalar(t).add(this.start)}closestPointToPointParameter(t,e){lf.subVectors(t,this.start),Ro.subVectors(this.end,this.start);const n=Ro.dot(Ro);let s=Ro.dot(lf)/n;return e&&(s=pe(s,0,1)),s}closestPointToPoint(t,e,n){const i=this.closestPointToPointParameter(t,e);return this.delta(n).multiplyScalar(i).add(this.start)}distanceSqToLine3(t,e=Mp,n=Sp){const i=10000000000000001e-32;let s,r;const a=this.start,l=t.start,c=this.end,d=t.end;Zs.subVectors(c,a),$s.subVectors(d,l),Sh.subVectors(a,l);const m=Zs.dot(Zs),g=$s.dot($s),_=$s.dot(Sh);if(m<=i&&g<=i)return e.copy(a),n.copy(l),e.sub(n),e.dot(e);if(m<=i)s=0,r=_/g,r=pe(r,0,1);else{const v=Zs.dot(Sh);if(g<=i)r=0,s=pe(-v/m,0,1);else{const M=Zs.dot($s),C=m*g-M*M;C!==0?s=pe((M*_-v*g)/C,0,1):s=0,r=(M*s+_)/g,r<0?(r=0,s=pe(-v/m,0,1)):r>1&&(r=1,s=pe((M-v)/m,0,1))}}return e.copy(a).addScaledVector(Zs,s),n.copy(l).addScaledVector($s,r),e.distanceToSquared(n)}applyMatrix4(t){return this.start.applyMatrix4(t),this.end.applyMatrix4(t),this}equals(t){return t.start.equals(this.start)&&t.end.equals(this.end)}clone(){return new this.constructor().copy(this)}}const cf=new O;class cm extends De{constructor(t,e){super(),this.light=t,this.matrixAutoUpdate=!1,this.color=e,this.type="SpotLightHelper";const n=new y,i=[0,0,0,0,0,1,0,0,0,1,0,1,0,0,0,-1,0,1,0,0,0,0,1,1,0,0,0,0,-1,1];for(let r=0,a=1,l=32;r<l;r++,a++){const c=r/l*Math.PI*2,d=a/l*Math.PI*2;i.push(Math.cos(c),Math.sin(c),1,Math.cos(d),Math.sin(d),1)}n.setAttribute("position",new Ut(i,3));const s=new Pn({fog:!1,toneMapped:!1});this.cone=new Si(n,s),this.add(this.cone),this.update()}dispose(){this.cone.geometry.dispose(),this.cone.material.dispose()}update(){this.light.updateWorldMatrix(!0,!1),this.light.target.updateWorldMatrix(!0,!1),this.parent?(this.parent.updateWorldMatrix(!0),this.matrix.copy(this.parent.matrixWorld).invert().multiply(this.light.matrixWorld)):this.matrix.copy(this.light.matrixWorld),this.matrixWorld.copy(this.light.matrixWorld);const t=this.light.distance?this.light.distance:1e3,e=t*Math.tan(this.light.angle);this.cone.scale.set(e,e,t),cf.setFromMatrixPosition(this.light.target.matrixWorld),this.cone.lookAt(cf),this.color!==void 0?this.cone.material.color.set(this.color):this.cone.material.color.copy(this.light.color)}}const Wi=new O,Io=new Me,bh=new Me;class hm extends Si{constructor(t){const e=hf(t),n=new y,i=[],s=[];for(let c=0;c<e.length;c++){const d=e[c];d.parent&&d.parent.isBone&&(i.push(0,0,0),i.push(0,0,0),s.push(0,0,0),s.push(0,0,0))}n.setAttribute("position",new Ut(i,3)),n.setAttribute("color",new Ut(s,3));const r=new Pn({vertexColors:!0,depthTest:!1,depthWrite:!1,toneMapped:!1,transparent:!0});super(n,r),this.isSkeletonHelper=!0,this.type="SkeletonHelper",this.root=t,this.bones=e,this.matrix=t.matrixWorld,this.matrixAutoUpdate=!1;const a=new re(255),l=new re(65280);this.setColors(a,l)}updateMatrixWorld(t){const e=this.bones,n=this.geometry,i=n.getAttribute("position");bh.copy(this.root.matrixWorld).invert();for(let s=0,r=0;s<e.length;s++){const a=e[s];a.parent&&a.parent.isBone&&(Io.multiplyMatrices(bh,a.matrixWorld),Wi.setFromMatrixPosition(Io),i.setXYZ(r,Wi.x,Wi.y,Wi.z),Io.multiplyMatrices(bh,a.parent.matrixWorld),Wi.setFromMatrixPosition(Io),i.setXYZ(r+1,Wi.x,Wi.y,Wi.z),r+=2)}n.getAttribute("position").needsUpdate=!0,super.updateMatrixWorld(t)}setColors(t,e){const i=this.geometry.getAttribute("color");for(let s=0;s<i.count;s+=2)i.setXYZ(s,t.r,t.g,t.b),i.setXYZ(s+1,e.r,e.g,e.b);return i.needsUpdate=!0,this}dispose(){this.geometry.dispose(),this.material.dispose()}}function hf(u){const t=[];u.isBone===!0&&t.push(u);for(let e=0;e<u.children.length;e++)t.push(...hf(u.children[e]));return t}class um extends kn{constructor(t,e,n){const i=new vo(e,4,2),s=new ei({wireframe:!0,fog:!1,toneMapped:!1});super(i,s),this.light=t,this.color=n,this.type="PointLightHelper",this.matrix=this.light.matrixWorld,this.matrixAutoUpdate=!1,this.update()}dispose(){this.geometry.dispose(),this.material.dispose()}update(){this.light.updateWorldMatrix(!0,!1),this.color!==void 0?this.material.color.set(this.color):this.material.color.copy(this.light.color)}}const bp=new O,uf=new re,ff=new re;class fm extends De{constructor(t,e,n){super(),this.light=t,this.matrix=t.matrixWorld,this.matrixAutoUpdate=!1,this.color=n,this.type="HemisphereLightHelper";const i=new _o(e);i.rotateY(Math.PI*.5),this.material=new ei({wireframe:!0,fog:!1,toneMapped:!1}),this.color===void 0&&(this.material.vertexColors=!0);const s=i.getAttribute("position"),r=new Float32Array(s.count*3);i.setAttribute("color",new se(r,3)),this.add(new kn(i,this.material)),this.update()}dispose(){this.children[0].geometry.dispose(),this.children[0].material.dispose()}update(){const t=this.children[0];if(this.color!==void 0)this.material.color.set(this.color);else{const e=t.geometry.getAttribute("color");uf.copy(this.light.color),ff.copy(this.light.groundColor);for(let n=0,i=e.count;n<i;n++){const s=n<i/2?uf:ff;e.setXYZ(n,s.r,s.g,s.b)}e.needsUpdate=!0}this.light.updateWorldMatrix(!0,!1),t.lookAt(bp.setFromMatrixPosition(this.light.matrixWorld).negate())}}class dm extends Si{constructor(t=10,e=10,n=4473924,i=8947848){n=new re(n),i=new re(i);const s=e/2,r=t/e,a=t/2,l=[],c=[];for(let g=0,_=0,v=-a;g<=e;g++,v+=r){l.push(-a,0,v,a,0,v),l.push(v,0,-a,v,0,a);const M=g===s?n:i;M.toArray(c,_),_+=3,M.toArray(c,_),_+=3,M.toArray(c,_),_+=3,M.toArray(c,_),_+=3}const d=new y;d.setAttribute("position",new Ut(l,3)),d.setAttribute("color",new Ut(c,3));const m=new Pn({vertexColors:!0,toneMapped:!1});super(d,m),this.type="GridHelper"}dispose(){this.geometry.dispose(),this.material.dispose()}}class pm extends Si{constructor(t=10,e=16,n=8,i=64,s=4473924,r=8947848){s=new re(s),r=new re(r);const a=[],l=[];if(e>1)for(let m=0;m<e;m++){const g=m/e*(Math.PI*2),_=Math.sin(g)*t,v=Math.cos(g)*t;a.push(0,0,0),a.push(_,0,v);const M=m&1?s:r;l.push(M.r,M.g,M.b),l.push(M.r,M.g,M.b)}for(let m=0;m<n;m++){const g=m&1?s:r,_=t-t/n*m;for(let v=0;v<i;v++){let M=v/i*(Math.PI*2),C=Math.sin(M)*_,w=Math.cos(M)*_;a.push(C,0,w),l.push(g.r,g.g,g.b),M=(v+1)/i*(Math.PI*2),C=Math.sin(M)*_,w=Math.cos(M)*_,a.push(C,0,w),l.push(g.r,g.g,g.b)}}const c=new y;c.setAttribute("position",new Ut(a,3)),c.setAttribute("color",new Ut(l,3));const d=new Pn({vertexColors:!0,toneMapped:!1});super(c,d),this.type="PolarGridHelper"}dispose(){this.geometry.dispose(),this.material.dispose()}}const df=new O,Po=new O,pf=new O;class mm extends De{constructor(t,e,n){super(),this.light=t,this.matrix=t.matrixWorld,this.matrixAutoUpdate=!1,this.color=n,this.type="DirectionalLightHelper",e===void 0&&(e=1);let i=new y;i.setAttribute("position",new Ut([-e,e,0,e,e,0,e,-e,0,-e,-e,0,-e,e,0],3));const s=new Pn({fog:!1,toneMapped:!1});this.lightPlane=new ds(i,s),this.add(this.lightPlane),i=new y,i.setAttribute("position",new Ut([0,0,0,0,0,1],3)),this.targetLine=new ds(i,s),this.add(this.targetLine),this.update()}dispose(){this.lightPlane.geometry.dispose(),this.lightPlane.material.dispose(),this.targetLine.geometry.dispose(),this.targetLine.material.dispose()}update(){this.light.updateWorldMatrix(!0,!1),this.light.target.updateWorldMatrix(!0,!1),df.setFromMatrixPosition(this.light.matrixWorld),Po.setFromMatrixPosition(this.light.target.matrixWorld),pf.subVectors(Po,df),this.lightPlane.lookAt(Po),this.color!==void 0?(this.lightPlane.material.color.set(this.color),this.targetLine.material.color.set(this.color)):(this.lightPlane.material.color.copy(this.light.color),this.targetLine.material.color.copy(this.light.color)),this.targetLine.lookAt(Po),this.targetLine.scale.z=pf.length()}}const Lo=new O,en=new ph;class gm extends Si{constructor(t){const e=new y,n=new Pn({color:16777215,vertexColors:!0,toneMapped:!1}),i=[],s=[],r={};a("n1","n2"),a("n2","n4"),a("n4","n3"),a("n3","n1"),a("f1","f2"),a("f2","f4"),a("f4","f3"),a("f3","f1"),a("n1","f1"),a("n2","f2"),a("n3","f3"),a("n4","f4"),a("p","n1"),a("p","n2"),a("p","n3"),a("p","n4"),a("u1","u2"),a("u2","u3"),a("u3","u1"),a("c","t"),a("p","c"),a("cn1","cn2"),a("cn3","cn4"),a("cf1","cf2"),a("cf3","cf4");function a(v,M){l(v),l(M)}function l(v){i.push(0,0,0),s.push(0,0,0),r[v]===void 0&&(r[v]=[]),r[v].push(i.length/3-1)}e.setAttribute("position",new Ut(i,3)),e.setAttribute("color",new Ut(s,3)),super(e,n),this.type="CameraHelper",this.camera=t,this.camera.updateProjectionMatrix&&this.camera.updateProjectionMatrix(),this.matrix=t.matrixWorld,this.matrixAutoUpdate=!1,this.pointMap=r,this.update();const c=new re(16755200),d=new re(16711680),m=new re(43775),g=new re(16777215),_=new re(3355443);this.setColors(c,d,m,g,_)}setColors(t,e,n,i,s){const a=this.geometry.getAttribute("color");return a.setXYZ(0,t.r,t.g,t.b),a.setXYZ(1,t.r,t.g,t.b),a.setXYZ(2,t.r,t.g,t.b),a.setXYZ(3,t.r,t.g,t.b),a.setXYZ(4,t.r,t.g,t.b),a.setXYZ(5,t.r,t.g,t.b),a.setXYZ(6,t.r,t.g,t.b),a.setXYZ(7,t.r,t.g,t.b),a.setXYZ(8,t.r,t.g,t.b),a.setXYZ(9,t.r,t.g,t.b),a.setXYZ(10,t.r,t.g,t.b),a.setXYZ(11,t.r,t.g,t.b),a.setXYZ(12,t.r,t.g,t.b),a.setXYZ(13,t.r,t.g,t.b),a.setXYZ(14,t.r,t.g,t.b),a.setXYZ(15,t.r,t.g,t.b),a.setXYZ(16,t.r,t.g,t.b),a.setXYZ(17,t.r,t.g,t.b),a.setXYZ(18,t.r,t.g,t.b),a.setXYZ(19,t.r,t.g,t.b),a.setXYZ(20,t.r,t.g,t.b),a.setXYZ(21,t.r,t.g,t.b),a.setXYZ(22,t.r,t.g,t.b),a.setXYZ(23,t.r,t.g,t.b),a.setXYZ(24,e.r,e.g,e.b),a.setXYZ(25,e.r,e.g,e.b),a.setXYZ(26,e.r,e.g,e.b),a.setXYZ(27,e.r,e.g,e.b),a.setXYZ(28,e.r,e.g,e.b),a.setXYZ(29,e.r,e.g,e.b),a.setXYZ(30,e.r,e.g,e.b),a.setXYZ(31,e.r,e.g,e.b),a.setXYZ(32,n.r,n.g,n.b),a.setXYZ(33,n.r,n.g,n.b),a.setXYZ(34,n.r,n.g,n.b),a.setXYZ(35,n.r,n.g,n.b),a.setXYZ(36,n.r,n.g,n.b),a.setXYZ(37,n.r,n.g,n.b),a.setXYZ(38,i.r,i.g,i.b),a.setXYZ(39,i.r,i.g,i.b),a.setXYZ(40,s.r,s.g,s.b),a.setXYZ(41,s.r,s.g,s.b),a.setXYZ(42,s.r,s.g,s.b),a.setXYZ(43,s.r,s.g,s.b),a.setXYZ(44,s.r,s.g,s.b),a.setXYZ(45,s.r,s.g,s.b),a.setXYZ(46,s.r,s.g,s.b),a.setXYZ(47,s.r,s.g,s.b),a.setXYZ(48,s.r,s.g,s.b),a.setXYZ(49,s.r,s.g,s.b),a.needsUpdate=!0,this}update(){const t=this.geometry,e=this.pointMap,n=1,i=1;let s,r;if(en.projectionMatrixInverse.copy(this.camera.projectionMatrixInverse),this.camera.reversedDepth===!0)s=1,r=0;else if(this.camera.coordinateSystem===Nn)s=-1,r=1;else if(this.camera.coordinateSystem===Pi)s=0,r=1;else throw new Error("THREE.CameraHelper.update(): Invalid coordinate system: "+this.camera.coordinateSystem);an("c",e,t,en,0,0,s),an("t",e,t,en,0,0,r),an("n1",e,t,en,-n,-i,s),an("n2",e,t,en,n,-i,s),an("n3",e,t,en,-n,i,s),an("n4",e,t,en,n,i,s),an("f1",e,t,en,-n,-i,r),an("f2",e,t,en,n,-i,r),an("f3",e,t,en,-n,i,r),an("f4",e,t,en,n,i,r),an("u1",e,t,en,n*.7,i*1.1,s),an("u2",e,t,en,-n*.7,i*1.1,s),an("u3",e,t,en,0,i*2,s),an("cf1",e,t,en,-n,0,r),an("cf2",e,t,en,n,0,r),an("cf3",e,t,en,0,-i,r),an("cf4",e,t,en,0,i,r),an("cn1",e,t,en,-n,0,s),an("cn2",e,t,en,n,0,s),an("cn3",e,t,en,0,-i,s),an("cn4",e,t,en,0,i,s),t.getAttribute("position").needsUpdate=!0}dispose(){this.geometry.dispose(),this.material.dispose()}}function an(u,t,e,n,i,s,r){Lo.set(i,s,r).unproject(n);const a=t[u];if(a!==void 0){const l=e.getAttribute("position");for(let c=0,d=a.length;c<d;c++)l.setXYZ(a[c],Lo.x,Lo.y,Lo.z)}}const Do=new A;class _m extends Si{constructor(t,e=16776960){const n=new Uint16Array([0,1,1,2,2,3,3,0,4,5,5,6,6,7,7,4,0,4,1,5,2,6,3,7]),i=new Float32Array(24),s=new y;s.setIndex(new se(n,1)),s.setAttribute("position",new se(i,3)),super(s,new Pn({color:e,toneMapped:!1})),this.object=t,this.type="BoxHelper",this.matrixAutoUpdate=!1,this.update()}update(){if(this.object!==void 0&&Do.setFromObject(this.object),Do.isEmpty())return;const t=Do.min,e=Do.max,n=this.geometry.attributes.position,i=n.array;i[0]=e.x,i[1]=e.y,i[2]=e.z,i[3]=t.x,i[4]=e.y,i[5]=e.z,i[6]=t.x,i[7]=t.y,i[8]=e.z,i[9]=e.x,i[10]=t.y,i[11]=e.z,i[12]=e.x,i[13]=e.y,i[14]=t.z,i[15]=t.x,i[16]=e.y,i[17]=t.z,i[18]=t.x,i[19]=t.y,i[20]=t.z,i[21]=e.x,i[22]=t.y,i[23]=t.z,n.needsUpdate=!0,this.geometry.computeBoundingSphere()}setFromObject(t){return this.object=t,this.update(),this}copy(t,e){return super.copy(t,e),this.object=t.object,this}dispose(){this.geometry.dispose(),this.material.dispose()}}class xm extends Si{constructor(t,e=16776960){const n=new Uint16Array([0,1,1,2,2,3,3,0,4,5,5,6,6,7,7,4,0,4,1,5,2,6,3,7]),i=[1,1,1,-1,1,1,-1,-1,1,1,-1,1,1,1,-1,-1,1,-1,-1,-1,-1,1,-1,-1],s=new y;s.setIndex(new se(n,1)),s.setAttribute("position",new Ut(i,3)),super(s,new Pn({color:e,toneMapped:!1})),this.box=t,this.type="Box3Helper",this.geometry.computeBoundingSphere()}updateMatrixWorld(t){const e=this.box;e.isEmpty()||(e.getCenter(this.position),e.getSize(this.scale),this.scale.multiplyScalar(.5),super.updateMatrixWorld(t))}dispose(){this.geometry.dispose(),this.material.dispose()}}class vm extends ds{constructor(t,e=1,n=16776960){const i=n,s=[1,-1,0,-1,1,0,-1,-1,0,1,1,0,-1,1,0,-1,-1,0,1,-1,0,1,1,0],r=new y;r.setAttribute("position",new Ut(s,3)),r.computeBoundingSphere(),super(r,new Pn({color:i,toneMapped:!1})),this.type="PlaneHelper",this.plane=t,this.size=e;const a=[1,1,0,-1,1,0,-1,-1,0,1,1,0,-1,-1,0,1,-1,0],l=new y;l.setAttribute("position",new Ut(a,3)),l.computeBoundingSphere(),this.add(new kn(l,new ei({color:i,opacity:.2,transparent:!0,depthWrite:!1,toneMapped:!1})))}updateMatrixWorld(t){this.position.set(0,0,0),this.scale.set(.5*this.size,.5*this.size,1),this.lookAt(this.plane.normal),this.translateZ(-this.plane.constant),super.updateMatrixWorld(t)}dispose(){this.geometry.dispose(),this.material.dispose(),this.children[0].geometry.dispose(),this.children[0].material.dispose()}}const mf=new O;let Uo,Th;class ym extends De{constructor(t=new O(0,0,1),e=new O(0,0,0),n=1,i=16776960,s=n*.2,r=s*.2){super(),this.type="ArrowHelper",Uo===void 0&&(Uo=new y,Uo.setAttribute("position",new Ut([0,0,0,0,1,0],3)),Th=new lo(.5,1,5,1),Th.translate(0,-.5,0)),this.position.copy(e),this.line=new ds(Uo,new Pn({color:i,toneMapped:!1})),this.line.matrixAutoUpdate=!1,this.add(this.line),this.cone=new kn(Th,new ei({color:i,toneMapped:!1})),this.cone.matrixAutoUpdate=!1,this.add(this.cone),this.setDirection(t),this.setLength(n,s,r)}setDirection(t){if(t.y>.99999)this.quaternion.set(0,0,0,1);else if(t.y<-.99999)this.quaternion.set(1,0,0,0);else{mf.set(t.z,0,-t.x).normalize();const e=Math.acos(t.y);this.quaternion.setFromAxisAngle(mf,e)}}setLength(t,e=t*.2,n=e*.2){this.line.scale.set(1,Math.max(1e-4,t-e),1),this.line.updateMatrix(),this.cone.scale.set(n,e,n),this.cone.position.y=t,this.cone.updateMatrix()}setColor(t){this.line.material.color.set(t),this.cone.material.color.set(t)}copy(t){return super.copy(t,!1),this.line.copy(t.line),this.cone.copy(t.cone),this}dispose(){this.line.geometry.dispose(),this.line.material.dispose(),this.cone.geometry.dispose(),this.cone.material.dispose()}}class Mm extends Si{constructor(t=1){const e=[0,0,0,t,0,0,0,0,0,0,t,0,0,0,0,0,0,t],n=[1,0,0,1,.6,0,0,1,0,.6,1,0,0,0,1,0,.6,1],i=new y;i.setAttribute("position",new Ut(e,3)),i.setAttribute("color",new Ut(n,3));const s=new Pn({vertexColors:!0,toneMapped:!1});super(i,s),this.type="AxesHelper"}setColors(t,e,n){const i=new re,s=this.geometry.attributes.color.array;return i.set(t),i.toArray(s,0),i.toArray(s,3),i.set(e),i.toArray(s,6),i.toArray(s,9),i.set(n),i.toArray(s,12),i.toArray(s,15),this.geometry.attributes.color.needsUpdate=!0,this}dispose(){this.geometry.dispose(),this.material.dispose()}}class Sm{constructor(){this.type="ShapePath",this.color=new re,this.subPaths=[],this.currentPath=null}moveTo(t,e){return this.currentPath=new Kc,this.subPaths.push(this.currentPath),this.currentPath.moveTo(t,e),this}lineTo(t,e){return this.currentPath.lineTo(t,e),this}quadraticCurveTo(t,e,n,i){return this.currentPath.quadraticCurveTo(t,e,n,i),this}bezierCurveTo(t,e,n,i,s,r){return this.currentPath.bezierCurveTo(t,e,n,i,s,r),this}splineThru(t){return this.currentPath.splineThru(t),this}toShapes(t){function e(w){const L=[];for(let D=0,N=w.length;D<N;D++){const et=w[D],rt=new ks;rt.curves=et.curves,L.push(rt)}return L}function n(w,L){const D=L.length;let N=!1;for(let et=D-1,rt=0;rt<D;et=rt++){let vt=L[et],at=L[rt],Mt=at.x-vt.x,St=at.y-vt.y;if(Math.abs(St)>Number.EPSILON){if(St<0&&(vt=L[rt],Mt=-Mt,at=L[et],St=-St),w.y<vt.y||w.y>at.y)continue;if(w.y===vt.y){if(w.x===vt.x)return!0}else{const Gt=St*(w.x-vt.x)-Mt*(w.y-vt.y);if(Gt===0)return!0;if(Gt<0)continue;N=!N}}else{if(w.y!==vt.y)continue;if(at.x<=w.x&&w.x<=vt.x||vt.x<=w.x&&w.x<=at.x)return!0}}return N}const i=pi.isClockWise,s=this.subPaths;if(s.length===0)return[];let r,a,l;const c=[];if(s.length===1)return a=s[0],l=new ks,l.curves=a.curves,c.push(l),c;let d=!i(s[0].getPoints());d=t?!d:d;const m=[],g=[];let _=[],v=0,M;g[v]=void 0,_[v]=[];for(let w=0,L=s.length;w<L;w++)a=s[w],M=a.getPoints(),r=i(M),r=t?!r:r,r?(!d&&g[v]&&v++,g[v]={s:new ks,p:M},g[v].s.curves=a.curves,d&&v++,_[v]=[]):_[v].push({h:a,p:M[0]});if(!g[0])return e(s);if(g.length>1){let w=!1,L=0;for(let D=0,N=g.length;D<N;D++)m[D]=[];for(let D=0,N=g.length;D<N;D++){const et=_[D];for(let rt=0;rt<et.length;rt++){const vt=et[rt];let at=!0;for(let Mt=0;Mt<g.length;Mt++)n(vt.p,g[Mt].p)&&(D!==Mt&&L++,at?(at=!1,m[Mt].push(vt)):w=!0);at&&m[D].push(vt)}}L>0&&w===!1&&(_=m)}let C;for(let w=0,L=g.length;w<L;w++){l=g[w].s,c.push(l),C=_[w];for(let D=0,N=C.length;D<N;D++)l.holes.push(C[D].h)}return c}}class bm extends Kn{constructor(t,e=null){super(),this.object=t,this.domElement=e,this.enabled=!0,this.state=-1,this.keys={},this.mouseButtons={LEFT:null,MIDDLE:null,RIGHT:null},this.touches={ONE:null,TWO:null}}connect(t){if(t===void 0){le("Controls: connect() now requires an element.");return}this.domElement!==null&&this.disconnect(),this.domElement=t}disconnect(){}dispose(){}update(){}}function Tp(u,t){const e=u.image&&u.image.width?u.image.width/u.image.height:1;return e>t?(u.repeat.x=1,u.repeat.y=e/t,u.offset.x=0,u.offset.y=(1-u.repeat.y)/2):(u.repeat.x=t/e,u.repeat.y=1,u.offset.x=(1-u.repeat.x)/2,u.offset.y=0),u}function Ap(u,t){const e=u.image&&u.image.width?u.image.width/u.image.height:1;return e>t?(u.repeat.x=t/e,u.repeat.y=1,u.offset.x=(1-u.repeat.x)/2,u.offset.y=0):(u.repeat.x=1,u.repeat.y=e/t,u.offset.x=0,u.offset.y=(1-u.repeat.y)/2),u}function Ep(u){return u.repeat.x=1,u.repeat.y=1,u.offset.x=0,u.offset.y=0,u}function gf(u,t,e,n){const i=wp(n);switch(e){case ua:return u*t;case cr:return u*t/i.components*i.byteLength;case hr:return u*t/i.components*i.byteLength;case pa:return u*t*2/i.components*i.byteLength;case ma:return u*t*2/i.components*i.byteLength;case fa:return u*t*3/i.components*i.byteLength;case Ci:return u*t*4/i.components*i.byteLength;case ga:return u*t*4/i.components*i.byteLength;case _a:case xa:return Math.floor((u+3)/4)*Math.floor((t+3)/4)*8;case va:case ya:return Math.floor((u+3)/4)*Math.floor((t+3)/4)*16;case Sa:case El:return Math.max(u,16)*Math.max(t,8)/4;case Ma:case Al:return Math.max(u,8)*Math.max(t,8)/2;case wl:case Cl:case Il:case Pl:return Math.floor((u+3)/4)*Math.floor((t+3)/4)*8;case Rl:case Ll:case Dl:return Math.floor((u+3)/4)*Math.floor((t+3)/4)*16;case Ul:return Math.floor((u+3)/4)*Math.floor((t+3)/4)*16;case Nl:return Math.floor((u+4)/5)*Math.floor((t+3)/4)*16;case Fl:return Math.floor((u+4)/5)*Math.floor((t+4)/5)*16;case Ol:return Math.floor((u+5)/6)*Math.floor((t+4)/5)*16;case Bl:return Math.floor((u+5)/6)*Math.floor((t+5)/6)*16;case zl:return Math.floor((u+7)/8)*Math.floor((t+4)/5)*16;case Vl:return Math.floor((u+7)/8)*Math.floor((t+5)/6)*16;case kl:return Math.floor((u+7)/8)*Math.floor((t+7)/8)*16;case Gl:return Math.floor((u+9)/10)*Math.floor((t+4)/5)*16;case Hl:return Math.floor((u+9)/10)*Math.floor((t+5)/6)*16;case Wl:return Math.floor((u+9)/10)*Math.floor((t+7)/8)*16;case Xl:return Math.floor((u+9)/10)*Math.floor((t+9)/10)*16;case ql:return Math.floor((u+11)/12)*Math.floor((t+9)/10)*16;case Yl:return Math.floor((u+11)/12)*Math.floor((t+11)/12)*16;case Zl:case $l:case Jl:return Math.ceil(u/4)*Math.ceil(t/4)*16;case Kl:case Ql:return Math.ceil(u/4)*Math.ceil(t/4)*8;case jl:case tc:return Math.ceil(u/4)*Math.ceil(t/4)*16}throw new Error(`Unable to determine texture byte length for ${e} format.`)}function wp(u){switch(u){case Is:case na:return{byteLength:1,components:1};case sa:case ia:case aa:return{byteLength:2,components:1};case oa:case la:return{byteLength:2,components:4};case Zi:case ra:case gi:return{byteLength:4,components:1};case ca:case ha:return{byteLength:4,components:3}}throw new Error(`Unknown texture type ${u}.`)}class Tm{static contain(t,e){return Tp(t,e)}static cover(t,e){return Ap(t,e)}static fill(t){return Ep(t)}static getByteLength(t,e,n,i){return gf(t,e,n,i)}}typeof __THREE_DEVTOOLS__<"u"&&__THREE_DEVTOOLS__.dispatchEvent(new CustomEvent("register",{detail:{revision:qr}})),typeof window<"u"&&(window.__THREE__?le("WARNING: Multiple instances of Three.js being imported."):window.__THREE__=qr)})}]);

//# sourceMappingURL=1627.aed499ff.chunk.js.map