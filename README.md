# SCUG: Strongly-Typed Unity Asset Codegenerator

![Coverage](https://img.shields.io/badge/Coverage-Unknown-lightgrey)

SCUG (Super Cool Unity Generator) is a high-performance, standalone Go tool that bridges the gap between Unity's loosely-typed asset system and C#'s strong-typing requirements. It automatically generates safe, performant C# wrappers for your Prefabs, Scenes, Tags, and Resources.

Stop using fragile transform.Find() strings and GameObject.FindWithTag() literals. SCUG gives you compile-time safety and sub-second feedback loops.

---

## Key Features

*   Deep Prefab Hierarchy: Generates nested classes that mirror your .prefab structure.
*   Strongly-Typed Tags & Maps: Replaces fragile enums with readonly struct patterns that preserve Unity's exact string capitalization (e.g., Tags.Wall.ToString() returns "wall", not "Wall").
*   JSON Serialization Support: Includes built-in System.Text.Json converters so structs behave like enums/strings during serialization, supporting both string and integer tokens for backward compatibility.
*   Global Resource Tree: Generates the Res static class for everything in Resources/, including typed access to Sprites, Materials, and Audio.
*   GUID Script Resolution: Automatically maps Unity's internal GUIDs to your C# namespaces and classes.
*   Incremental & Background Performance: Uses a persistent JSON cache and fast hashing to run in the background without freezing the Unity Editor.
*   Automatic Array Grouping: Children with numeric suffixes (e.g., Item0, Item1) are automatically exposed as Item_Array.

---

## Why SCUG?

*   Replace Fragile Lookups: Move away from string-based transform.Find("Child/Path") and GetComponent<T>() boilerplate.
*   Strongly-Typed API: Access your prefab hierarchy directly through C# properties (e.g., LevelBar.Image).
*   Compile-Time Safety: Catch broken hierarchy references and missing components at compile time, not runtime.
*   Cross-Platform: A standalone Go tool that runs natively on Windows, Linux, and macOS.
*   Fast & Lightweight: Incremental builds and background execution mean you never wait for the editor to "catch up".

---

## Project Status

### Currently Working
- [x] Deep Prefab Parsing: Reconstructs full GameObject/Transform hierarchies.
- [x] Nested Prefab Support: Seamlessly links wrappers for nested prefabs.
- [x] Smart Script Resolution: Automatically maps Unity GUIDs to C# classes.
- [x] Component Deduplication: Handles multiple components of the same type.
- [x] Automatic Array Grouping: Groups children with numeric suffixes (e.g., Item0, Item1) into arrays.
- [x] Incremental Builds: Persistent JSON cache and hashing for sub-second updates.
- [x] Scenes & Tags Generation: Generates Map and Tags wrappers from Unity settings.
- [x] JSON Compatibility: Built-in support for System.Text.Json serialization.
- [x] Universal Resource Generation: Generates the Res static tree for all files in Resources/.

### Roadmap
- [ ] Scene Parsing: Support for generating wrappers for entire Unity Scenes.
- [ ] UI Toolkit Support: Targeted generation for UXML/USS based UI.

---

## Usage Examples

### 1. Prefab Interaction
Instead of:
```csharp
var goldText = transform.Find("TopBar/Currency/Label").GetComponent<TextMeshProUGUI>();
```
Use SCUG:
```csharp
var overlay = CanvasOverlay.Get(this.gameObject);
overlay.TopBar.Currency.Label.TextMeshProUGUI.text = "100";
```

### 2. Tags & Scenes (Exact Strings)
Unity tags are often lowercase or snake_case. SCUG preserves this perfectly:
```csharp
// Implicit conversion to string works!
if (other.CompareTag(Tags.Defensebuildings)) { ... } 

// ToString() returns exactly "defensebuildings"
string tagValue = Tags.Defensebuildings.ToString(); 

// Maps work the same way
Map.Load(Maps.Social);
```

### 3. Static Resources
```csharp
// Fully typed paths, no more typos in Resources.Load paths
var icon = Res.Icons.UiIcons.GoldToken;
var sfx = Res.Audio.Sfx.Click;
```

---

## Technical Details

SCUG is architected for performance and reliability, operating entirely outside the Unity main thread to ensure a smooth development experience.

### Architecture

*   **Go-Based Parser**: A custom, high-performance YAML block parser that reads `.prefab` files. It avoids the overhead of loading assets through the Unity API.
*   **GUID Resolution**: Scans your `Assets` folder (cached) to build a mapping between Unity's internal GUIDs and C# types.
*   **Code Generation**: Uses a non-reflective, string-builder approach to emit clean, predictable C# code.

### Architectural Decisions

#### Why Structs instead of Enums for Tags/Maps?
Standard C# enums cannot override .ToString(). In Unity, tags and scene names are often case-sensitive strings. SCUG uses a readonly struct pattern so that:
1.  You get the IntelliSense and Type Safety of an enum.
2.  You get Implicit String Conversion (you can pass the struct where a string is expected).
3.  .ToString() returns the actual value defined in Unity (e.g., "wall").
4.  Serialization: Includes a JsonConverter for System.Text.Json compatibility, handling both string and integer representations.

#### Why SCUG does NOT statically cache Resources.Load
SCUG does not cache Sprites or Materials in static variables. This is a deliberate decision to prioritize Memory Safety:
1. Unity's Internal Cache: Resources.Load is fast because Unity keeps assets in an internal C++ dictionary once loaded.
2. Avoiding Memory Leaks: Static variables are GC Roots. If SCUG cached everything, Unity could never unload assets via Resources.UnloadUnusedAssets(), leading to OOM crashes on mobile.

---

## Installation & Requirements

### Requirements
*   Go: Version 1.26 or higher.
*   Unity: Tested with version 2021.3+ and 6000.x.

### Installation
1.  Place the generator folder in your Unity project root.
2.  Install Go 1.26+.
3.  Copy the scripts from Assets/Editor to your project's Assets/Editor folder.

### Recommended Usage (Unity Editor)
We highly recommend using the included Unity Editor Scripts. They handle everything automatically:
*   Automatic Build: Compiles the generator binary on Unity startup.
*   Background Generation: Triggers a run whenever you save a prefab, running in the background to avoid freezing the editor.
*   Automatic Cleanup: Deletes the temporary executable and cache file when you close Unity.

### Standalone Usage
You can also run the generator manually from your terminal:
```bash
# Full scan and prime cache
go run cmd/scug/main.go

# Targeted generation for specific files
go run cmd/scug/main.go "Assets/Resources/MyPrefab.prefab"
```

---

## Configuration (scug.json)

Place a scug.json in your project root to customize behavior:

```json
{
  "disable_cache": false,
  "cache_file": "scug_cache.json",
  "resources_dir": "Assets/Resources",
  "output_dir": "Assets/Scripts/v2/UX/generated",
  "scan_dirs": [
    "Assets/Scripts",
    "Library/PackageCache",
    "Packages"
  ],
  "workers": 16
}
```

---

## Source Control (.gitignore)

To keep your repository clean, ensure the following are added to your .gitignore:
```ignore
# SCUG Temporary Files
generator/scug
generator/scug.exe
generator/scug_cache.json
```
*(The provided root `.gitignore` in this repo already includes these.)*

---

## Contributing
We welcome contributions! Please ensure any changes to the Go source include corresponding tests in the *_test.go files.

SCUG: Access your hierarchy directly. Never deal with broken references again.
