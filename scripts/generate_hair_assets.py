#!/usr/bin/env python3
"""
WOOHWAHAE — Hair Asset Generator (Blender headless)

Usage:
    blender --background --python scripts/generate_hair_assets.py

Output:
    website/assets/3d/hair_{style}.glb  (5 files)

Requirements: Blender 3.6+ or 4.x
"""

import bpy
import bmesh
import math
import os
from mathutils import Vector

# ─── PATHS ────────────────────────────────────────────────────────────────────
ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
OUT = os.path.join(ROOT, "website", "assets", "3d")
os.makedirs(OUT, exist_ok=True)

# ─── HEAD DIMENSIONS (Blender units) ──────────────────────────────────────────
HRX = 0.85   # width
HRY = 0.80   # depth (front–back)
HRZ = 1.05   # height

# ─── HAIR STYLE PARAMS ────────────────────────────────────────────────────────
# length   : 0.0 (buzz) → 1.0 (very long)
# bow      : lateral volume beyond head radius (0=tight, 0.35=voluminous)
# taper    : how much bottom narrows (0=straight fall, 1=sharp taper)
# crown    : crown lift above head top (0=flat, 0.3=lifted)
# depth    : front-back thickness of hair mesh (Blender units)
STYLES = {
    "short_minimal":  dict(length=0.13, bow=0.01, taper=0.72, crown=0.04, depth=0.15),
    "medium_natural": dict(length=0.50, bow=0.13, taper=0.30, crown=0.18, depth=0.26),
    "medium_texture": dict(length=0.48, bow=0.24, taper=0.20, crown=0.22, depth=0.30),
    "long_classic":   dict(length=0.85, bow=0.09, taper=0.46, crown=0.25, depth=0.23),
    "long_bold":      dict(length=0.88, bow=0.34, taper=0.09, crown=0.07, depth=0.36),
}


# ─── SCENE HELPERS ────────────────────────────────────────────────────────────

def reset_scene():
    bpy.ops.object.select_all(action="SELECT")
    bpy.ops.object.delete()
    for collection in [bpy.data.meshes, bpy.data.materials, bpy.data.curves]:
        for block in list(collection):
            if block.users == 0:
                collection.remove(block)


def _make_material(name: str, color_rgba: tuple, roughness: float) -> bpy.types.Material:
    mat = bpy.data.materials.new(name)
    mat.use_nodes = True
    bsdf = mat.node_tree.nodes["Principled BSDF"]
    bsdf.inputs["Base Color"].default_value = color_rgba
    bsdf.inputs["Roughness"].default_value = roughness
    # Specular: Blender 3.x = index 5, 4.x = "Specular IOR Level"
    try:
        bsdf.inputs["Specular IOR Level"].default_value = 0.05
    except KeyError:
        try:
            bsdf.inputs[5].default_value = 0.05
        except (IndexError, KeyError):
            pass
    return mat


# ─── HEAD MESH ────────────────────────────────────────────────────────────────

def make_head() -> bpy.types.Object:
    bpy.ops.mesh.primitive_uv_sphere_add(segments=32, ring_count=24, radius=1.0)
    obj = bpy.context.active_object
    obj.name = "Head"
    obj.scale = (HRX, HRY, HRZ)
    bpy.ops.object.transform_apply(scale=True)
    bpy.ops.object.shade_smooth()
    mat = _make_material("M_Head", (0.88, 0.80, 0.72, 1.0), roughness=0.85)
    obj.data.materials.append(mat)
    return obj


# ─── HAIR MESH ────────────────────────────────────────────────────────────────

def _profile_right(p: dict) -> list:
    """
    Returns [(x, z), ...] for the RIGHT half of the hair silhouette.
    Coordinate system: X=right, Z=up (Blender space).
    Will be mirrored to build full closed outline.
    """
    L       = p["length"]
    bow     = p["bow"]
    taper   = p["taper"]
    crown   = p["crown"]

    top_z      = HRZ + crown * HRZ * 0.58       # crown peak (above head)
    temple_z   = HRZ * 0.28                      # hairline at temple height
    bottom_z   = temple_z - L * 2.8             # hair bottom

    temple_x   = HRX * 0.87                      # temple width
    max_x      = temple_x + bow * HRX            # maximum lateral width
    bot_x      = max_x * (1.0 - taper * 0.65)   # bottom width

    pts = [
        (0.0,         top_z),                         # crown center
        (HRX * 0.32,  top_z - 0.03),                  # crown inner slope
        (HRX * 0.76,  HRZ * 0.88),                    # upper side
        (temple_x,    temple_z + 0.14),               # near temple
        (temple_x,    temple_z),                      # temple
    ]

    if L > 0.06:
        mid_z = temple_z - (temple_z - bottom_z) * 0.38
        pts += [
            (max_x,         mid_z),
            (bot_x,         bottom_z + abs(bottom_z - temple_z) * 0.12),
            (bot_x * 0.90,  bottom_z),
        ]

    return pts


def make_hair_mesh(style: str, p: dict) -> bpy.types.Object:
    """
    Builds a flat extruded mesh representing the hair silhouette.
    Profile in XZ plane → mirrored → extruded along Y.
    """
    pts_r = _profile_right(p)
    # Bottom center closing point
    bottom_z = pts_r[-1][1] if p["length"] > 0.06 else HRZ * 0.28 - 0.08

    # Full outline: right side → center-bottom → mirrored left side (reversed, skip ends)
    center_bottom = [(0.0, bottom_z)]
    pts_l = [(-x, z) for (x, z) in reversed(pts_r[1:])]  # skip crown center (shared)
    outline = pts_r + center_bottom + pts_l

    depth = p["depth"]
    half_d = depth / 2.0

    bm = bmesh.new()

    # Front face (y = -half_d)
    front_verts = [bm.verts.new(Vector((x, -half_d, z))) for (x, z) in outline]
    bm.faces.new(front_verts)

    bm.verts.ensure_lookup_table()
    bm.faces.ensure_lookup_table()

    # Extrude front face → back face
    ret = bmesh.ops.extrude_face_region(bm, geom=list(bm.faces))
    back_verts = [e for e in ret["geom"] if isinstance(e, bmesh.types.BMVert)]
    bmesh.ops.translate(bm, verts=back_verts, vec=Vector((0.0, depth, 0.0)))

    bmesh.ops.recalc_face_normals(bm, faces=list(bm.faces))

    mesh = bpy.data.meshes.new(f"Hair_{style}")
    bm.to_mesh(mesh)
    bm.free()

    obj = bpy.data.objects.new(f"Hair_{style}", mesh)
    bpy.context.scene.collection.objects.link(obj)
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)

    bpy.ops.object.shade_smooth()

    mat = _make_material("M_Hair", (0.07, 0.06, 0.05, 1.0), roughness=0.92)
    obj.data.materials.append(mat)

    return obj


# ─── EXPORT ───────────────────────────────────────────────────────────────────

def export_glb(style: str):
    path = os.path.join(OUT, f"hair_{style}.glb")
    bpy.ops.export_scene.gltf(
        filepath=path,
        export_format="GLB",
        export_apply=True,
        export_materials="EXPORT",
        use_visible=True,
    )
    size_kb = os.path.getsize(path) // 1024
    print(f"  ✓ {path}  ({size_kb} KB)")


# ─── MAIN ─────────────────────────────────────────────────────────────────────

def main():
    print(f"\n{'─'*56}")
    print(f"  WOOHWAHAE Hair Asset Generator")
    print(f"  Output → {OUT}")
    print(f"{'─'*56}\n")

    for style, params in STYLES.items():
        print(f"  Generating: {style}")
        reset_scene()
        make_head()
        make_hair_mesh(style, params)
        export_glb(style)

    print(f"\n  Done. {len(STYLES)} files generated.\n")


main()
