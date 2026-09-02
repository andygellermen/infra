const $ = (selector) => document.querySelector(selector);
const all = (selector) => [...document.querySelectorAll(selector)];

if (document.body.dataset.page === "product") {
  const state = { profile: "PRIVATE", result: null, activeQuestion: null, pairs: [], sessionID: newSessionID() };
  const text = $("#analysis-text");
  const count = $("#char-count");
  const updateCount = () => { count.textContent = `${[...text.value].length.toLocaleString("de-DE")} / 10.000`; };
  text.addEventListener("input", updateCount);
  updateCount();

  all('input[name="profile"]').forEach((input) => input.addEventListener("change", () => {
    state.profile = input.value;
    $("#brand-name").textContent = input.value === "CORPORATE" ? "Sprachkompass" : "MeineSprache";
    document.title = `${$("#brand-name").textContent} · Sprach-A-Lyzer`;
  }));

  $("#analyze-button").addEventListener("click", async () => {
    const button = $("#analyze-button");
    button.disabled = true;
    $("#form-error").hidden = true;
    try {
      const response = await fetch("/api/v6/experience/analyze", {method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({text:text.value,context:$("#context").value,profile:state.profile,language_level:$("#language-level").value})});
      const payload = await response.json();
      if (!response.ok) throw new Error(payload.error?.message || "Die Analyse ist gerade nicht verfügbar.");
      state.result = payload; state.pairs = []; state.sessionID = newSessionID();
      renderResult(payload);
    } catch (error) {
      $("#form-error").textContent = error.message; $("#form-error").hidden = false;
    } finally { button.disabled = false; }
  });

  function renderResult(result) {
    $("#result-headline").textContent = result.headline;
    $("#result-summary").textContent = result.summary;
    $("#privacy-receipt").textContent = `${result.privacy.mode} · nicht gespeichert`;
    const dimensions = $("#dimensions"); dimensions.replaceChildren();
    result.dimensions.forEach((dimension) => {
      const row = element("div", "dimension-row");
      row.append(element("strong", "", dimension.label));
      const meter = element("progress", "meter"); meter.max=100; meter.value=Math.round(dimension.assessability * 100); row.append(meter);
      row.append(element("span", "dimension-value", dimension.score == null ? "offen" : `${Math.round(dimension.score)} %`)); dimensions.append(row);
    });
    const patterns = $("#patterns"); patterns.replaceChildren();
    if (!result.core_result.patterns.length) patterns.append(element("p", "empty-state", "Kein gesichertes Muster – der Core erfindet keine Deutung."));
    result.core_result.patterns.forEach((pattern) => patterns.append(element("span", "tag", readable(pattern))));
    const trace = $("#trace"); trace.replaceChildren();
    if (!result.explanation_trace.length) trace.append(element("p", "empty-state", "Keine scoring-wirksame Evidenz vorhanden."));
    result.explanation_trace.forEach((item) => { const node=element("div","trace-item"); node.append(element("strong","",`${item.label} · ${signed(item.delta)}`),element("span","",`${item.evidence} – ${item.reason}`)); trace.append(node); });
    $("#reflection-question").textContent = result.reflection_question || "Welche Formulierung fühlt sich für dich stimmiger an?";
    const alternatives = $("#alternatives"); alternatives.replaceChildren();
    if (!result.alternatives.length) alternatives.append(element("p", "empty-state", "Für diesen Text liegt noch keine redaktionell freigegebene Alternative vor."));
    result.alternatives.forEach((value) => { const button=element("button","alternative",value); button.type="button"; button.addEventListener("click",()=>{all(".alternative").forEach(x=>x.classList.remove("selected"));button.classList.add("selected");$("#feedback-status").textContent="Nur für diese Sitzung ausgewählt.";}); alternatives.append(button); });
    const questionList = $("#questions"); questionList.replaceChildren();
    result.suggested_questions.forEach((question) => { const button=element("button","question-button",question.text); button.type="button"; button.addEventListener("click",()=>selectQuestion(question,button)); questionList.append(button); });
    const notices=$("#notices"); notices.replaceChildren(); result.notices.forEach(n=>notices.append(element("span","",`✓ ${n}`)));
    $("#answer-form").hidden=true; $("#session-result").hidden=true; $("#results").hidden=false; $("#results").scrollIntoView({behavior:"smooth",block:"start"});
  }

  function selectQuestion(question, button) { state.activeQuestion=question; all(".question-button").forEach(x=>x.classList.remove("active"));button.classList.add("active");$("#active-question").textContent=question.text;$("#answer-form").hidden=false;$("#answer-text").focus(); }

  $("#answer-form").addEventListener("submit", async (event) => {
    event.preventDefault(); if (!state.activeQuestion) return;
    const answer=$("#answer-text").value.trim(); if (!answer) return;
    state.pairs.push({question_id:state.activeQuestion.question_id,answer,profile:state.profile});
    const output=$("#session-result"); output.hidden=false;
    try {
      const response=await fetch("/api/v3/sessions/compose",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({session_id:state.sessionID,profile:state.profile,pairs:state.pairs})});
      const payload=await response.json();
      if (!response.ok) throw new Error(payload.error?.message || "Die Antwort konnte nicht betrachtet werden.");
      const observation=payload.observations[payload.observations.length-1]; output.textContent=`${readable(observation.assessability)} · ${observation.qa_patterns.length ? observation.qa_patterns.map(readable).join(", ") : "Beobachtung ohne zusätzliche Deutung"}. Inferenzniveau ${payload.inference_level}.`;
      $("#answer-text").value="";
    } catch (error) {
      state.pairs.pop(); output.textContent=error.message;
    }
  });

  all(".feedback").forEach((button)=>button.addEventListener("click",()=>{$("#feedback-status").textContent=button.dataset.value==="HELPFUL"?"Danke – lokal für diese Sitzung vorgemerkt.":"Danke – es wird nichts übertragen.";}));
}

if (document.body.dataset.page === "admin") {
  const refresh = async () => { const target=$("#api-status"); target.textContent="Prüfung …"; try { const response=await fetch("/health/ready"); target.textContent=response.ok?"READY":"DEGRADED"; } catch { target.textContent="OFFLINE"; } };
  $("#refresh-status").addEventListener("click",refresh); refresh();
}

function element(tag, className="", value="") { const node=document.createElement(tag); if(className)node.className=className;if(value)node.textContent=value;return node; }
function newSessionID() { if (globalThis.crypto && typeof globalThis.crypto.randomUUID === "function") return globalThis.crypto.randomUUID(); return `session-${Date.now()}-${Math.random().toString(16).slice(2)}`; }
function readable(value) { return String(value).toLowerCase().replaceAll("_"," ").replace(/(^|\s)\S/g,(letter)=>letter.toUpperCase()); }
function signed(value) { return `${value>0?"+":""}${Number(value).toFixed(1)}`; }
