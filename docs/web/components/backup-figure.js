/* Custom figure drawn with hairline-create. Engine: Lucas Marques, MIT.
 * Scroll replaces pointer selection in the homepage; the bench keeps a pointer tour.
 */
export function mount({ stage, svg, read, interactive = true }, value) {
  const {
    Cam, fit, proj, facing, rings, rrect, prism, ringAt, run, hull,
    poly, open, seg, clamp, lerp, mk, solid, put, tween, tset, tval,
    tdone, pointer, register, disposer,
  } = HL;
  const bag = disposer();
  const C = Cam(45, 0.5, 1.62);
  fit(C, [[-88, -45, 0], [88, 65, 0], [-88, 65, 0], [88, -45, 85]], 200, 166);
  const P = proj(C), front = facing(C);
  const g = mk("g", {}, svg);
  const base = solid(g);
  const [foot, inset] = rings(-88, -45, 88, 65, 5, 1.5);
  put(base, prism(P, front, foot, inset, -6, 0));
  const guides = mk("path", { class: "nf dash" }, g);
  const source = mk("g", {}, g), copies = mk("g", {}, g), pages = mk("g", {}, g);
  const cylinder = solid(source);
  const seams = mk("path", { class: "nf lo" }, source);
  const lid = mk("path", { class: "nf" }, source);
  const remote = solid(copies);
  const remoteSlot = mk("path", { class: "nf lo" }, copies);
  const remotePlatter = solid(copies);
  const local = solid(pages);
  const localSlot = mk("path", { class: "nf lo" }, pages);
  const retained = Array.from({ length: 5 }, () => solid(pages));
  const sheets = Array.from({ length: 5 }, () => {
    const group = mk("g", {}, pages);
    return { body: solid(group), rules: mk("path", { class: "nf lo" }, group) };
  });
  const [localRing, localInner] = rings(-11, 9, 38, 49, 4, 1.2);
  const [remoteRing, remoteInner] = rings(38, -35, 78, -2, 4, 1.2);
  const drumRing = rrect(-71, -13, -27, 31, 22, 16);
  const drumInner = rrect(-69.8, -11.8, -28.2, 29.8, 20.8, 16);
  const [platterRing, platterInner] = rings(42, -31, 74, -6, 3, 1);
  const clocks = [tween(0), tween(0), tween(0), tween(0), tween(0)];
  let fan = value, selected = 0, last = "";

  function draw(now) {
    const q = clocks.map((clock) => tval(clock, now));
    const signature = q.join(",") + ":" + fan;
    if (signature === last) return;
    last = signature;
    const pack = clamp(q[0], 0, 1), send = clamp(q[0] - 1, 0, 1);
    const height = lerp(55, 39, pack);
    put(cylinder, prism(P, front, drumRing, drumInner, 3, height));
    seams.setAttribute("d", [0.27, 0.53, 0.78].map((t) =>
      open(ringAt(P, run(drumRing, front), lerp(3, height, t)))
    ).join(""));
    lid.setAttribute("d", poly(ringAt(P, rrect(-65, -7, -33, 25, 16, 16), height)));
    put(local, prism(P, front, localRing, localInner, 2, 14));
    localSlot.setAttribute("d", open([P(-2, 49, 8), P(28, 49, 8)]));
    const remoteHeight = lerp(5, 27, send);
    put(remote, prism(P, front, remoteRing, remoteInner, 2, remoteHeight));
    remoteSlot.setAttribute("d", seg(P(47, -2, remoteHeight - 4), P(69, -2, remoteHeight - 4)));
    put(remotePlatter, prism(P, front, platterRing, platterInner, remoteHeight + 2, remoteHeight + 4 + 5 * send));
    guides.setAttribute("d", open([P(-48, 20, 0.5), P(-48, 40, 0.5), P(12, 40, 0.5)]) +
      open([P(16, 12, 0.5), P(16, -18, 0.5), P(58, -18, 0.5)]));
    sheets.forEach((sheet, i) => {
      put(retained[i], prism(P, front, ...rings(-8, 12, 33, 45, 2, 0.8),
        16 + i * 2.6 * pack, 16 + i * 2.6 * pack + 1.2));
      const a = clamp(q[i], 0, 1), b = clamp(q[i] - 1, 0, 1);
      const x = lerp(-31 + i * 7, -8, a) + b * 49;
      const y = lerp(-2 + i * 2, 12, a) - b * 38;
      const z = lerp(38 + i * fan, 16 + i * 2.6, a) + b * 17;
      const w = lerp(36, 41, a) - b * 12, h = lerp(29, 33, a) - b * 11;
      const [ring, inner] = rings(x, y, x + w, y + h, 2, 0.8);
      put(sheet.body, prism(P, front, ring, inner, z, z + 1.2));
      sheet.rules.setAttribute("d", [7, 12, 17].map((v, j) =>
        seg(P(x + 5, y + v, z + 1.3), P(x + w - 5 - j * 2, y + v, z + 1.3))
      ).join(""));
      sheet.body.sil.classList.toggle("hi", selected === 1 && i === 4);
    });
    cylinder.sil.classList.toggle("hi", selected === 0);
    remote.sil.classList.toggle("hi", selected === 2);
    // The remote store is behind the file until the file arrives on its platter.
    if (pages.previousSibling !== copies) copies.after(pages);
  }

  const loop = register(stage, (_dt, now) => {
    draw(now);
    return clocks.some((clock) => !tdone(clock, now));
  });
  bag.add(loop.unregister);
  function setStage(index) {
    selected = clamp(index, 0, 2);
    clocks.forEach((clock, i) => tset(clock, selected, performance.now(), Math.abs(i - 2) * 35));
    last = "";
    loop.wake();
  }
  if (interactive) bag.add(pointer(stage, {
    move: ([x, y]) => {
      if (y < 55 || y > 265) { read.textContent = "rest"; setStage(0); return; }
      const index = x < 155 ? 0 : x > 245 ? 2 : 1;
      read.textContent = ["dump", "file", "copy"][index];
      setStage(index);
    },
    leave: () => { read.textContent = "rest"; setStage(0); },
  }));
  bag.add(() => svg.replaceChildren());
  draw(performance.now());
  return {
    set: (v) => { fan = clamp(v, 3, 9); last = ""; loop.wake(); },
    setStage,
    destroy: bag.dispose,
  };
}

if (typeof hairline === "function") hairline({
  name: "backup",
  means: "A database opens into sheets, settles into a local file, then sends a copy to a second store.",
  rules: [3, 6, 7, 8, 9, 10],
  range: [3, 6, 9],
  tour: [[120, 160], [200, 160], [280, 160], null],
  mount,
});
