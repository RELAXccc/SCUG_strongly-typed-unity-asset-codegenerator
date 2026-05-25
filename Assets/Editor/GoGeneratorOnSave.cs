using UnityEngine;
using UnityEditor;
using System.Diagnostics;
using System.IO;
using System.Linq;
using System.Text;
using System.Collections.Generic;

/// <summary>
/// Unity Editor integration for SCUG.
/// Automatically triggers the Go generator when a prefab is saved.
/// </summary>
public class GoGeneratorOnSave : UnityEditor.AssetModificationProcessor
{
    /// <summary>
    /// If true, triggers AssetDatabase.Refresh() after a successful Go run, which causes a Domain Reload.
    /// Set to false to avoid the reload wait; logs will be printed immediately.
    /// </summary>
    public static bool RefreshAssetDatabase = false;

    private static List<BackgroundProcessState> _activeProcesses = new List<BackgroundProcessState>();

    /// <summary>
    /// Hook called by Unity whenever an asset is about to be saved.
    /// </summary>
    static string[] OnWillSaveAssets(string[] paths)
    {
        // Only trigger for changed prefab files.
        string[] savedPrefabs = paths.Where(path => path.EndsWith(".prefab", System.StringComparison.OrdinalIgnoreCase)).ToArray();

        if (savedPrefabs.Length > 0)
        {
            // Delay the call to ensure Unity has finished its own save operations.
            EditorApplication.delayCall += () => RunGoGenerator(savedPrefabs);
        }

        return paths;
    }

    /// <summary>
    /// Manual trigger for the generator via the Tools menu.
    /// </summary>
    [MenuItem("Tools/Run Go Generator")]
    public static void RunGoGeneratorMenu()
    {
        // ... (rest of implementation)
        string[] selectedPrefabs = Selection.objects
            .Select(o => AssetDatabase.GetAssetPath(o))
            .Where(p => p.EndsWith(".prefab", System.StringComparison.OrdinalIgnoreCase))
            .ToArray();

        RunGoGenerator(selectedPrefabs.Length > 0 ? selectedPrefabs : null);
    }

    public static void RunGoGenerator(string[] changedFiles, string taskName = "Targeted")
    {
        string projectRoot = Directory.GetParent(Application.dataPath).FullName;
        string generatorPath = Path.Combine(projectRoot, "generator");

        if (!Directory.Exists(generatorPath))
        {
            UnityEngine.Debug.LogError($"[Go Generator] Could not find folder at: {generatorPath}");
            return;
        }

        ProcessStartInfo startInfo = new ProcessStartInfo();
        startInfo.WorkingDirectory = generatorPath;

        string exeName = Application.platform == RuntimePlatform.WindowsEditor ? "scug.exe" : "scug";
        string exePath = Path.Combine(generatorPath, exeName);

        if (!File.Exists(exePath))
        {
            if (Application.platform == RuntimePlatform.WindowsEditor)
            {
                startInfo.FileName = "cmd.exe";
                startInfo.Arguments = $"/c go run cmd/scug/main.go";
            }
            else
            {
                startInfo.FileName = "go";
                startInfo.Arguments = "run cmd/scug/main.go";
            }
        }
        else
        {
            startInfo.FileName = exePath;
            startInfo.Arguments = "";
        }

        if (changedFiles != null && changedFiles.Length > 0)
        {
            string fileArgs = string.Join(" ", changedFiles.Select(p => $"\"{p}\""));
            startInfo.Arguments += (string.IsNullOrEmpty(startInfo.Arguments) ? "" : " ") + fileArgs;
        }

        startInfo.UseShellExecute = false;
        startInfo.RedirectStandardOutput = true;
        startInfo.RedirectStandardError = true;
        startInfo.CreateNoWindow = true;

        try
        {
            Process process = new Process();
            process.StartInfo = startInfo;
            
            var state = new BackgroundProcessState {
                Process = process,
                StartTime = (float)EditorApplication.timeSinceStartup,
                Files = changedFiles,
                TaskName = taskName,
                Output = new StringBuilder(),
                Error = new StringBuilder()
            };

            process.OutputDataReceived += (s, e) => { if (e.Data != null) state.Output.AppendLine(e.Data); };
            process.ErrorDataReceived += (s, e) => { if (e.Data != null) state.Error.AppendLine(e.Data); };

            process.Start();
            process.BeginOutputReadLine();
            process.BeginErrorReadLine();
            
            _activeProcesses.Add(state);
            
            if (_activeProcesses.Count == 1)
            {
                EditorApplication.update += MonitorProcesses;
            }
            
            UnityEngine.Debug.Log($"[Go Generator] Started {taskName} task in background...");
        }
        catch (System.Exception e)
        {
            UnityEngine.Debug.LogError($"[Go Generator] Exception occurred: {e.Message}");
        }
    }

    private class BackgroundProcessState
    {
        public Process Process;
        public float StartTime;
        public string[] Files;
        public string TaskName;
        public StringBuilder Output;
        public StringBuilder Error;
        public bool IsFinished;
    }

    private static void MonitorProcesses()
    {
        for (int i = _activeProcesses.Count - 1; i >= 0; i--)
        {
            var state = _activeProcesses[i];
            
            // Safety timeout (5 minutes)
            if ((float)EditorApplication.timeSinceStartup - state.StartTime > 300f)
            {
                UnityEngine.Debug.LogError($"[Go Generator] {state.TaskName} task TIMEOUT. Killing process.");
                state.Process.Kill();
                state.IsFinished = true;
            }
            else if (state.Process.HasExited)
            {
                state.IsFinished = true;
            }

            if (state.IsFinished)
            {
                HandleProcessExit(state);
                _activeProcesses.RemoveAt(i);
            }
        }

        if (_activeProcesses.Count == 0)
        {
            EditorApplication.update -= MonitorProcesses;
        }
    }

    private static void HandleProcessExit(BackgroundProcessState state)
    {
        int exitCode = state.Process.ExitCode;
        float duration = (float)EditorApplication.timeSinceStartup - state.StartTime;
        string outputStr = state.Output.ToString();
        string errorStr = state.Error.ToString();

        if (exitCode == 0)
        {
            bool triggeredRefresh = false;
            if (RefreshAssetDatabase && state.Files != null)
            {
                SessionState.SetBool("GoGen_HasPendingLogs", true);
                SessionState.SetString("GoGen_Output", outputStr);
                SessionState.SetString("GoGen_Error", errorStr);
                AssetDatabase.Refresh();
                triggeredRefresh = true;
            }
            
            if (!triggeredRefresh)
            {
                if (!string.IsNullOrEmpty(outputStr))
                    UnityEngine.Debug.Log($"[Go Generator] {state.TaskName} Success ({duration:F2}s):\n{outputStr}");
                
                if (!string.IsNullOrEmpty(errorStr))
                    UnityEngine.Debug.LogWarning($"[Go Generator] {state.TaskName} Output:\n{errorStr}");
            }
        }
        else
        {
            string detail = $"[Go Generator] {state.TaskName} FAILED after {duration:F2}s (Exit Code {exitCode})\n";
            detail += "--------------------------------------------------\n";
            if (!string.IsNullOrEmpty(outputStr)) detail += $"STDOUT:\n{outputStr}\n";
            if (!string.IsNullOrEmpty(errorStr)) detail += $"STDERR:\n{errorStr}\n";
            detail += "--------------------------------------------------";
            UnityEngine.Debug.LogError(detail);
        }

        state.Process.Dispose();
    }

    // =========================================================================
    // POST-COMPILE LOGGING
    // This attribute tells Unity to run this method immediately after a Domain 
    // Reload (which happens right after AssetDatabase.Refresh finishes compiling).
    // =========================================================================
    [InitializeOnLoadMethod]
    private static void PrintPendingLogsAfterCompile()
    {
        if (SessionState.GetBool("GoGen_HasPendingLogs", false))
        {
            SessionState.SetBool("GoGen_HasPendingLogs", false);
            string output = SessionState.GetString("GoGen_Output", "");
            string error = SessionState.GetString("GoGen_Error", "");

            if (!string.IsNullOrEmpty(output))
            {
                UnityEngine.Debug.Log($"[Go Generator] Success:\n{output}");
            }
            
            if (!string.IsNullOrEmpty(error))
            {
                UnityEngine.Debug.LogWarning($"[Go Generator] Output:\n{error}");
            }
        }
    }
}
