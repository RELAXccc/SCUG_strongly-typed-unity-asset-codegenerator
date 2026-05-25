# SCUG: Strongly-Typed Unity Asset Codegenerator

Stop using fragile `transform.Find()` string lookups. **SCUG** (Super Cool Unity Generator) automatically creates safe C# wrappers for your Prefabs, Scenes, UI elements, and more. Get direct hierarchical access and extended helper methods to prevent broken references and typos.

## Why SCUG?

*   **Replace Fragile Lookups:** Move away from string-based `transform.Find("Child/Path")` and `GetComponent<T>()` boilerplate.
*   **Strongly-Typed API:** Access your prefab hierarchy directly through C# properties (e.g., `LevelBar.Image`).
*   **Compile-Time Safety:** Catch broken hierarchy references and missing components at compile time, not runtime.
*   **Cross-Platform:** A standalone Go tool that runs natively on Windows, Linux, and macOS.
*   **Fast & Lightweight:** Incremental builds and background execution mean you never wait for the editor to "catch up".

## Project Status

### Currently Working
- [x] **Deep Prefab Parsing**: Reconstructs full GameObject/Transform hierarchies.
- [x] **Nested Prefab Support**: Seamlessly links wrappers for nested prefabs.
- [x] **Smart Script Resolution**: Automatically maps Unity GUIDs to C# classes.
- [x] **Component Deduplication**: Handles multiple components of the same type.
- [x] **Automatic Array Grouping**: Groups children with numeric suffixes (e.g., `Item0`, `Item1`) into arrays.
- [x] **Incremental Builds**: Persistent JSON cache and hashing for sub-second updates.

### Roadmap (Future Goals)
- [ ] **Scene Parsing**: Support for generating wrappers for entire Unity Scenes.
- [ ] **Full Resource Scanning**: Make every asset in the `Resources` folder accessible.
- [ ] **UI Toolkit Support**: Targeted generation for UXML/USS based UI.

## Requirements

*   **Go**: Version **1.26** or higher.
*   **Unity**: Tested with version **6000.3.15f1** (Revision: `c1aa84e375f6`).

## Installation

1.  Place the `generator` folder in your Unity project root.
2.  Install [Go 1.26+](https://go.dev/dl/).
3.  Copy the scripts from `Assets/Editor` to your project's `Assets/Editor` folder.

### Recommended Usage
We highly recommend using the included **Unity Editor Scripts**. They handle everything automatically:
*   **Automatic Build**: Compiles the generator binary on Unity startup.
*   **Background Generation**: Triggers a targeted run whenever you save a prefab, running in the background to avoid freezing the editor.
*   **Automatic Cleanup**: Deletes the temporary executable and cache file when you close Unity to keep your project clean.

### Standalone Usage
You can also run the generator manually from your terminal for CI/CD or custom pipelines:
```bash
# Full scan and prime cache
go run cmd/scug/main.go

# Targeted generation for specific files
go run cmd/scug/main.go "Assets/Resources/MyPrefab.prefab"
```

## Source Control (.gitignore)

To keep your repository clean and avoid binary bloat, please ensure the following are added to your `.gitignore`:
```ignore
# SCUG Temporary Files
generator/scug
generator/scug.exe
generator/scug_cache.json
```
*(The provided root `.gitignore` in this repo already includes these.)*

## Example

Instead of maintaining fragile, string-based lookup code:
```csharp
// The old, fragile way
var goldText = transform.Find("TopBar/CurrencyGroup/GoldAmount").GetComponent<TextMeshProUGUI>();
goldText.text = "100";
```

You can use **SCUG** to get fully-typed, cached access with full IntelliSense support:
```csharp
using Prefabs.UIPrefabs.OverlayCanvas;

// Get a wrapper for the current GameObject
var overlay = CanvasOverlay.Get(this.gameObject);

// Navigate the hierarchy with zero string lookups
overlay.TopBar.CurrencyGroup.GoldAmount.TextMeshProUGUI.text = "100";

// SCUG also supports loading and instantiating prefabs directly!
var wrapper = CanvasOverlay.Instantiate(parentTransform);
wrapper.TopBar.SetActive(true);
```

## Technical Details

SCUG is architected for performance and reliability, operating entirely outside the Unity main thread to ensure a smooth development experience.

### Architecture

*   **Go-Based Parser**: A custom, high-performance YAML block parser that reads `.prefab` files. It avoids the overhead of loading assets through the Unity API.
*   **GUID Resolution**: Scans your `Assets` folder (cached) to build a mapping between Unity's internal GUIDs and C# types.
*   **Code Generation**: Uses a non-reflective, string-builder approach to emit clean, predictable C# code.

### C# Wrapper Structure

Generated wrappers use a `Wrapper` class as the entry point, containing:
*   `ResourcePath`: The path used for `Resources.Load`.
*   `Load()` / `Instantiate()`: Static helpers for spawning the prefab.
*   `Get()`: Static factory to wrap an existing `GameObject`.
*   **Nested Classes**: Each child GameObject becomes a nested class (e.g., `MyChild_Obj`), maintaining the hierarchy without string paths.

---

*SCUG: Access your hierarchy directly. Never deal with broken references again.*
