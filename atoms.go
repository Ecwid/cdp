package cdp

// Atom JS functions
const (
	atomClearInput       = `function(){("INPUT"===this.nodeName||"TEXTAREA"===this.nodeName)?this.value="":this.innerText=""}`
	atomGetInnerText     = `function(){return this.value||this.innerText}`
	atomDispatchEvents   = `function(l){for(const e of l)this.dispatchEvent(new Event(e,{'bubbles':!0}))}`
	atomSelect           = `function(a){const b=Array.from(this.options);this.value=void 0;for(const c of b)if(c.selected=a.includes(c.value),c.selected&&!this.multiple)break}`
	atomGetSelected      = `function(){return Array.from(this.options).filter(a=>a.selected).map(a=>a.value)}`
	atomGetSelectedText  = `function(){return Array.from(this.options).filter(a=>a.selected).map(a=>a.innerText)}`
	atomSelectContains   = `function(c){const a=Array.from(this.options);return c.length==a.filter(a=>c.includes(a.value)).length}`
	atomCheckBox         = `function(c){this.checked=c}`
	atomChecked          = `function(){return this.checked}`
	atomGetComputedStyle = `function(s){return getComputedStyle(this)[s]}`
	atomSetAttr          = `function(a,v){this.setAttribute(a,v)}`
	atomGetAttr          = `function(a){return this.getAttribute(a)}`
	atomIsVisible        = `function(){const b=this.getBoundingClientRect(),c=window.getComputedStyle(this);return c&&"hidden"!==c.visibility&&!c.disabled&&!!(b.top||b.bottom||b.width||b.height)}`
	atomClickDone        = `function(){return this._cc}`
	atomPreventMissClick = `function(){this._cc=!1,tt=this,z=function(b){for(var c=b;c;c=c.parentNode)if(c==tt)return!0;return!1},i=function(b){if (z(b.target)) {tt._cc=!0;} else {b.stopPropagation();b.preventDefault()}},document.addEventListener("click",i,{capture:!0,once:!0})}`
	atomMutationObserver = `function(b,d,c){return new Promise(e=>{const f=new MutationObserver(b=>{for(var c of b){e(c.type),f.disconnect();break}});f.observe(this,{attributes:b,childList:d,subtree:c})})}`
)
