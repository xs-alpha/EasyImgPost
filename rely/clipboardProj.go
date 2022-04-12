package rely

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

func write(file string) error {
	cmd := exec.Command("PowerShell", "-Command", "Add-Type", "-AssemblyName",
		fmt.Sprintf("System.Windows.Forms;[Windows.Forms.Clipboard]::SetImage([System.Drawing.Image]::FromFile('%s'));", file))
	b, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, string(b))
	}
	return nil
}

func read() (io.Reader, error) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, err
	}
	f.Close()
	defer os.Remove(f.Name())

	cmd := exec.Command("PowerShell", "-Command", "Add-Type", "-AssemblyName",
		fmt.Sprintf("System.Windows.Forms;$clip=[Windows.Forms.Clipboard]::GetImage();if ($clip -ne $null) { $clip.Save('%s') };", f.Name()))
	b, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s: %s", err, string(b))
	}

	r := new(bytes.Buffer)
	f, err = os.Open(f.Name())
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if _, err := io.Copy(r, f); err != nil {
		return nil, err
	}

	return r, nil
}

// Write write image to clipboard
func WriteIntoClipBoard(r io.Reader) error {
	f, err := writeTemp(r)
	if err != nil {
		return err
	}
	defer os.Remove(f)
	return write(f)
}

// Read read image from clipboard
func ReadFromClipBoard() (io.Reader, error) {
	return read()
}

func writeTemp(r io.Reader) (string, error) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}
	return f.Name(), nil
}
