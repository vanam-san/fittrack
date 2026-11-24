package com.wails.app;

import android.app.Activity;
import android.content.Intent;
import android.database.Cursor;
import android.net.Uri;
import android.provider.DocumentsContract;
import android.provider.OpenableColumns;
import android.webkit.JavascriptInterface;
import android.webkit.WebView;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.util.UUID;
import org.json.JSONObject;

/** Copies between Android document URIs and FitTrack's private staging files. */
public final class FitTrackFiles {
    private static final int DOCUMENT_REQUEST = 6410;
    private final Activity activity;
    private final WebView webView;
    private final File imports;
    private final File exports;
    private Integer request;
    private File source;

    FitTrackFiles(Activity activity, WebView webView) {
        this.activity = activity;
        this.webView = webView;
        imports = new File(activity.getCacheDir(), "fittrack-imports");
        exports = new File(activity.getFilesDir(), "exports");
    }

    @JavascriptInterface
    public void pick(int id, String operation, String path) {
        activity.runOnUiThread(() -> {
            if (request != null) {
                reply(id, "", "A document operation is already in progress.");
                return;
            }
            boolean saving = "save".equals(operation);
            File staged = saving ? stagedFile(exports, path) : null;
            if (saving && (staged == null || !staged.isFile())) {
                reply(id, "", "FitTrack could not open the staged export.");
                return;
            }

            Intent intent = new Intent(saving ? Intent.ACTION_CREATE_DOCUMENT : Intent.ACTION_OPEN_DOCUMENT);
            intent.addCategory(Intent.CATEGORY_OPENABLE);
            intent.putExtra(Intent.EXTRA_LOCAL_ONLY, true);
            intent.setType(saving ? mimeType(staged.getName()) : "*/*");
            if (saving) intent.putExtra(Intent.EXTRA_TITLE, staged.getName());
            request = id;
            source = staged;
            try {
                activity.startActivityForResult(intent, DOCUMENT_REQUEST);
            } catch (RuntimeException failure) {
                remove(staged);
                finish("", "FitTrack could not open Android's document picker.");
            }
        });
    }

    boolean onActivityResult(int code, int result, Intent data) {
        if (code != DOCUMENT_REQUEST) return false;
        if (request == null) return true;
        File staged = source;
        Uri uri = data == null ? null : data.getData();
        if (result != Activity.RESULT_OK || uri == null) {
            remove(staged);
            finish("", "");
            return true;
        }
        if (!"content".equals(uri.getScheme())) {
            remove(staged);
            finish("", "Choose a document from Android's file picker.");
            return true;
        }

        new Thread(() -> {
            String resultPath = "";
            String error = "";
            try {
                if (staged != null) {
                    save(staged, uri);
                    resultPath = staged.getName();
                } else {
                    resultPath = read(uri).getAbsolutePath();
                }
            } catch (IOException | RuntimeException failure) {
                error = staged == null
                    ? "FitTrack could not read the selected document."
                    : "FitTrack could not write the selected document.";
            } finally {
                remove(staged);
            }
            String path = resultPath;
            String message = error;
            activity.runOnUiThread(() -> finish(path, message));
        }, "FitTrack documents").start();
        return true;
    }

    @JavascriptInterface
    public void discard(String path) {
        remove(stagedFile(imports, path));
    }

    private void save(File file, Uri uri) throws IOException {
        try (InputStream input = new FileInputStream(file);
             OutputStream output = activity.getContentResolver().openOutputStream(uri, "wt")) {
            copy(input, output);
        } catch (IOException | RuntimeException failure) {
            try {
                DocumentsContract.deleteDocument(activity.getContentResolver(), uri);
            } catch (Exception ignored) {
                // Some providers do not support deleting a partially written document.
            }
            throw failure;
        }
    }

    private File read(Uri uri) throws IOException {
        String name = "document";
        try (Cursor cursor = activity.getContentResolver().query(
                uri, new String[] {OpenableColumns.DISPLAY_NAME}, null, null, null)) {
            if (cursor != null && cursor.moveToFirst()) {
                String displayName = cursor.getString(0);
                if (displayName != null) name = new File(displayName).getName();
            }
        }
        if (name.isEmpty() || ".".equals(name) || "..".equals(name)) {
            throw new IOException("Invalid document name");
        }
        File folder = new File(imports, UUID.randomUUID().toString());
        if (!folder.mkdirs()) throw new IOException("Could not create import folder");
        File file = new File(folder, name);
        try (InputStream input = activity.getContentResolver().openInputStream(uri);
             OutputStream output = new FileOutputStream(file)) {
            copy(input, output);
        } catch (IOException | RuntimeException failure) {
            remove(file);
            throw failure;
        }
        return file;
    }

    private static void copy(InputStream input, OutputStream output) throws IOException {
        if (input == null || output == null) throw new IOException("Could not open document");
        byte[] buffer = new byte[64 * 1024];
        int count;
        while ((count = input.read(buffer)) != -1) output.write(buffer, 0, count);
    }

    private static String mimeType(String name) {
        if (name.endsWith(".json")) return "application/json";
        if (name.endsWith(".csv")) return "text/csv";
        return "application/octet-stream";
    }

    private static File stagedFile(File root, String path) {
        try {
            File file = new File(path).getCanonicalFile();
            File folder = file.getParentFile();
            if (folder != null && root.getCanonicalFile().equals(folder.getParentFile())) return file;
        } catch (IOException ignored) {
        }
        return null;
    }

    private static void remove(File file) {
        if (file == null) return;
        file.delete();
        file.getParentFile().delete();
    }

    private void finish(String path, String error) {
        int id = request;
        request = null;
        source = null;
        reply(id, path, error);
    }

    private void reply(int id, String path, String error) {
        webView.evaluateJavascript("window.onFitTrackDocument?.(" + id + ","
            + JSONObject.quote(path) + "," + JSONObject.quote(error) + ")", null);
    }
}
