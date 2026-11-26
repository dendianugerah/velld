"use client";

import { useEffect, useState } from "react";
import styles from "./backup-story.module.css";

type V = [number, number];

// Isometric projection: +x runs right-down, +y runs left-down, +z is up.
const P = (x: number, y: number, z = 0): V => [
  Math.round((x - y) * 0.866 * 10) / 10,
  Math.round(((x + y) * 0.5 - z) * 10) / 10,
];
const pts = (...v: V[]) => v.map((p) => p.join(",")).join(" ");
const line = (...v: V[]) => "M" + v.map((p) => p.join(" ")).join("L");

type Part = { x: number; y: number; z?: number; w: number; d: number; h: number };

function Box({ x, y, z = 0, w, d, h }: Part) {
  const t = z + h;
  return (
    <>
      <polygon className={styles.side} points={pts(P(x, y + d, z), P(x + w, y + d, z), P(x + w, y + d, t), P(x, y + d, t))} />
      <polygon className={styles.side} points={pts(P(x + w, y, z), P(x + w, y + d, z), P(x + w, y + d, t), P(x + w, y, t))} />
      <polygon className={styles.top} points={pts(P(x, y, t), P(x + w, y, t), P(x + w, y + d, t), P(x, y + d, t))} />
    </>
  );
}

function Cylinder({ x, y, z, r, h, bands = 0 }: { x: number; y: number; z: number; r: number; h: number; bands?: number }) {
  const [bx, by] = P(x, y, z);
  const ty = by - h;
  const rx = r * 1.2247;
  const ry = r * 0.7071;
  const arcs = Array.from({ length: bands }, (_, i) => by - (h * (i + 1)) / (bands + 1));
  return (
    <>
      <path className={styles.side} d={"M" + (bx - rx) + " " + ty + "L" + (bx - rx) + " " + by + "A" + rx + " " + ry + " 0 0 0 " + (bx + rx) + " " + by + "L" + (bx + rx) + " " + ty + "Z"} />
      {arcs.map((ay) => (
        <path key={ay} className={styles.detail} d={"M" + (bx - rx) + " " + ay + "A" + rx + " " + ry + " 0 0 0 " + (bx + rx) + " " + ay} />
      ))}
      <ellipse className={styles.top} cx={bx} cy={ty} rx={rx} ry={ry} />
    </>
  );
}

function Tag({ at, text, tone }: { at: V; text: string; tone?: "ok" | "warn" }) {
  const width = text.length * 6.6 + 18;
  return (
    <g className={styles.tag} data-tone={tone} transform={"translate(" + at[0] + " " + at[1] + ")"}>
      <rect x={-width / 2} y={-11} width={width} height={22} rx={4} />
      <text x={0} y={3.8} textAnchor="middle">{text}</text>
    </g>
  );
}

function Callout({ letter, anchor, at }: { letter: string; anchor: V; at: V }) {
  const dx = anchor[0] - at[0];
  const dy = anchor[1] - at[1];
  const length = Math.hypot(dx, dy) || 1;
  return (
    <g className={styles.callout}>
      <path d={line([at[0] + (dx / length) * 11, at[1] + (dy / length) * 11], anchor)} />
      <circle className={styles.anchor} cx={anchor[0]} cy={anchor[1]} r={2.4} />
      <circle cx={at[0]} cy={at[1]} r={11} />
      <text x={at[0]} y={at[1] + 3.6} textAnchor="middle">{letter}</text>
    </g>
  );
}

// Data paths, in world coordinates, follow the code's order of operations.
const routes = {
  connect: line(P(75, 335, 41), P(75, 205, 41)),
  dump: line(P(90, 170, 14), P(130, 170, 14), P(130, 130, 14)),
  write: line(P(150, 110, 14), P(192, 110, 14), P(192, 85, 14)),
  upload: line(P(192, 55, 44), P(192, 55, 172), P(192, -46, 172)),
  verify: line(P(226, 55, 14), P(290, 55, 14), P(290, 110, 14), P(338, 110, 14)),
};
const packets: Record<number, string[]> = {
  1: [routes.connect + line(P(75, 205, 41), P(75, 205, 14)).replace("M", "L") + routes.dump.replace("M", "L") + routes.write.replace("M", "L")],
  2: [routes.upload],
  3: [routes.verify],
};

const usage: Record<string, number[]> = {
  database: [0, 1],
  tunnel: [0, 1],
  tool: [0, 1],
  gzip: [1],
  folder: [1, 2, 3],
  bucket: [2],
  container: [3],
};

export const figureCallouts: { anchor: V; at: V }[][] = [
  [
    { anchor: P(75, 268, 48), at: [-250, 70] },
    { anchor: [-160, 8], at: [-232, -10] },
    { anchor: [12, 57], at: [70, 20] },
  ],
  [
    { anchor: P(60, 170, 52), at: [-140, -6] },
    { anchor: P(130, 110, 38), at: [14, 18] },
    { anchor: P(196, 55, 42), at: [150, 30] },
  ],
  [
    { anchor: P(192, -80, 196), at: [300, -170] },
    { anchor: P(192, 55, 120), at: [60, -40] },
    { anchor: P(196, 55, 42), at: [236, 146] },
  ],
  [
    { anchor: P(405, 110, 34), at: [316, 196] },
    { anchor: P(316, 110, 14), at: [110, 245] },
    { anchor: [280, 122], at: [312, 84] },
  ],
];

function useReducedMotion() {
  const [reduced, setReduced] = useState(true);
  useEffect(() => {
    const query = window.matchMedia("(prefers-reduced-motion: reduce)");
    const update = () => setReduced(query.matches);
    update();
    query.addEventListener("change", update);
    return () => query.removeEventListener("change", update);
  }, []);
  return reduced;
}

export function BackupFigure({ active }: { active: number }) {
  const reduced = useReducedMotion();
  const state = (name: string) => {
    const stages = usage[name];
    if (stages.includes(active)) return "on";
    return stages[0] < active ? "past" : "off";
  };
  const flow = (stage: number) => (active === stage ? "on" : stage < active ? "past" : "off");
  const at = (stage: number) => (active === stage ? "on" : "off");

  return (
    <svg className={styles.svg} viewBox="-345 -185 675 495" data-stage={active} aria-hidden="true">
      <g className={styles.part} data-state={state("bucket")}>
        <Cylinder x={192} y={-80} z={150} r={34} h={46} bands={1} />
        <text className={styles.label} x={P(192, -80, 150)[0]} y={P(192, -80, 150)[1] + 40} textAnchor="middle">s3://bucket</text>
      </g>

      <g className={styles.base}>
        <Box x={0} y={0} w={240} d={220} h={12} />
        <text className={styles.label} x={P(150, 220, 0)[0]} y={P(150, 220, 0)[1] + 30} textAnchor="middle">your server · velld</text>
      </g>

      <path className={styles.flow} data-state={flow(1)} d={routes.dump} />
      <path className={styles.flow} data-state={flow(1)} d={routes.write} />
      <path className={styles.flow} data-state={flow(3)} d={routes.verify} />

      <g className={styles.part} data-state={state("folder")}>
        <Box x={160} y={25} z={12} w={66} d={60} h={8} />
        <Box x={170} y={35} z={20} w={46} d={40} h={4} />
        <Box x={170} y={35} z={26} w={46} d={40} h={4} />
        <g className={styles.file} data-purge={active === 2 ? "on" : "off"}>
          <Box x={170} y={35} z={32} w={46} d={40} h={4} />
        </g>
        <text className={styles.label} x={P(226, 25, 20)[0] + 10} y={P(226, 25, 20)[1] + 2}>backups/&lt;connection&gt;/</text>
      </g>

      <g className={styles.part} data-state={state("gzip")}>
        <Box x={110} y={90} z={12} w={40} d={40} h={26} />
        <text className={styles.face} x={P(130, 130, 22)[0]} y={P(130, 130, 22)[1] + 4} textAnchor="middle">gzip</text>
      </g>

      <g className={styles.part} data-state={state("tool")}>
        <Box x={30} y={140} z={12} w={60} d={60} h={40} />
        <path className={styles.detail} d={line(P(40, 200, 40), P(80, 200, 40))} />
        <path className={styles.detail} d={line(P(40, 200, 32), P(66, 200, 32))} />
        <text className={styles.face} x={P(60, 200, 24)[0]} y={P(60, 200, 24)[1] + 4} textAnchor="middle">dump</text>
      </g>

      <g className={styles.part} data-state={state("tunnel")}>
        <Box x={68} y={200} z={27} w={14} d={135} h={14} />
      </g>
      <path className={styles.flow} data-state={flow(0) === "on" || active === 1 ? "on" : flow(0)} d={routes.connect} />

      <g className={styles.part} data-state={state("database")}>
        <Box x={30} y={320} w={90} d={90} h={10} />
        <Cylinder x={75} y={365} z={10} r={30} h={56} bands={2} />
        <text className={styles.label} x={P(30, 410, 0)[0] + 10} y={P(30, 410, 0)[1] + 26}>database host</text>
      </g>

      <g className={styles.part} data-state={state("container")}>
        <Box x={320} y={60} w={100} d={100} h={8} />
        <g className={styles.ephemeral}>
          <Box x={335} y={75} z={8} w={70} d={70} h={50} />
        </g>
        {[90, 105, 120, 135].map((y) => (
          <path key={y} className={styles.detail} d={line(P(405, y, 14), P(405, y, 52))} />
        ))}
        <Cylinder x={370} y={110} z={58} r={16} h={20} bands={1} />
        <text className={styles.label} x={P(420, 160, 0)[0] + 4} y={P(420, 160, 0)[1] + 22} textAnchor="middle">throwaway container</text>
      </g>

      <path className={styles.flow} data-state={flow(2)} d={routes.upload} />

      <g className={styles.overlay} data-state={at(0)}>
        <g className={styles.card}>
          <Box x={42} y={158} z={70} w={36} d={24} h={2} />
        </g>
        <Tag at={[P(60, 170, 104)[0], P(60, 170, 104)[1] - 12]} text="PGPASSWORD=••••" />
        <Tag at={[P(75, 365, 66)[0], P(75, 365, 66)[1] - 30]} text="server 16" />
        <Tag at={[P(60, 170, 52)[0] + 160, P(60, 170, 52)[1] - 6]} text="pg_dump 17 ≥ 16" tone="ok" />
        <Tag at={[P(75, 268, 41)[0] + 20, P(75, 268, 41)[1] + 34]} text="127.0.0.1:<port>" />
      </g>
      <g className={styles.overlay} data-state={at(1)}>
        <Tag at={[P(226, 85, 0)[0] + 40, P(226, 85, 0)[1] + 34]} text="app_20261126_211400_<id>.sql.gz" />
        <Tag at={[P(60, 170, 52)[0] - 10, P(60, 170, 52)[1] - 34]} text="pg_dump · mysqldump · mongodump" />
        <Tag at={[P(130, 110, 38)[0] + 10, P(130, 110, 38)[1] - 30]} text="stdout → gzip" />
      </g>
      <g className={styles.overlay} data-state={at(2)}>
        <Tag at={[P(192, -80, 196)[0] - 120, P(192, -80, 196)[1] + 8]} text="<prefix>/<connection>/<file>" />
        <Tag at={[P(192, 55, 120)[0] - 62, P(192, 55, 120)[1] + 18]} text="fail → warning" tone="warn" />
      </g>
      <g className={styles.overlay} data-state={at(3)}>
        <Tag at={[P(370, 110, 78)[0] - 10, P(370, 110, 78)[1] - 40]} text="tables ✓  rows ✓" tone="ok" />
      </g>

      {figureCallouts.map((set, stage) => (
        <g key={stage} className={styles.overlay} data-state={at(stage)}>
          {set.map((c, i) => <Callout key={i} letter={"ABC"[i]} anchor={c.anchor} at={c.at} />)}
        </g>
      ))}

      {!reduced && packets[active]?.map((d) =>
        [0, 1, 2].map((i) => (
          <circle key={active + "-" + i} className={styles.packet} r={2.6}>
            <animateMotion dur="2.4s" begin={i * 0.8 + "s"} repeatCount="indefinite" path={d} />
          </circle>
        )),
      )}
    </svg>
  );
}

