# Deployment Guide: Homelab VM + Cloudflare Tunnel

This guide explains how to deploy the **Gamer Journal Wrapped** application to a homelab Linux VM and expose it securely to the internet using Cloudflare Tunnels.

## 1. Architecture Overview
The application is designed to be self-contained:
- **Backend (Go)**: Serves both the JSON API and the static frontend files.
- **Frontend (React)**: Compiled into static assets served by the Go binary.
- **Database**: Uses an internal memory-backed SQL server that bridges to Airtable.
- **Exposure**: Cloudflare Tunnel (cloudflared) provides a secure HTTPS entry point without opening firewall ports.

---

## 2. Deployment via Ansible (Recommended)
Ansible automates the build, packaging, upload, and service management in a single command.

### Prerequisites
- **Ansible** installed on your local machine.
- **Go** and **Node.js/npm** installed locally (for the build phase).
- SSH access to the VM (ideally with SSH keys).

### Configuration
1. **API Keys**: Open `deploy/vars.yml` and fill in your `airtable_api_key` and `serper_api_key`.
2. **Inventory**: Check `deploy/inventory.ini` to ensure the VM IP (`192.168.86.80`) and user (`ubuntu`) are correct.

### Run Deployment
```bash
ansible-playbook -i deploy/inventory.ini deploy/playbook.yml
```
This command will:
1. Build the frontend and backend **locally**.
2. Package everything into a tarball.
3. Upload, extract, and configure the `.env` on the VM.
4. Setup and restart the `systemd` service.

---

## 3. Cloudflared Tunnel Setup
To access your dashboard from a custom domain (e.g., `wrapped.yourdomain.com`):

### Installation on VM
1. [Download and install](https://pkg.cloudflare.com/) `cloudflared` on your VM.
2. **Authenticate**:
   ```bash
   cloudflared tunnel login
   ```
3. **Create Tunnel**:
   ```bash
   cloudflared tunnel create gamer-wrapped
   ```
4. **Configure DNS**:
   ```bash
   cloudflared tunnel route dns gamer-wrapped wrapped.yourdomain.com
   ```

### Configuration File
Create a configuration file (e.g., `~/.cloudflared/config.yml`):
```yaml
tunnel: <tunnel-id>
credentials-file: /home/ubuntu/.cloudflared/<tunnel-id>.json

ingress:
  - hostname: wrapped.yourdomain.com
    service: http://localhost:8080
  - service: http_status:404
```

### Run as Service
Install the tunnel as a system service to ensure it starts on boot:
```bash
sudo cloudflared service install
sudo systemctl start cloudflared
```
---

## 5. Managing the Service
On the VM, you can manage the application using standard `systemctl` commands:
```bash
sudo systemctl status gamer-wrapped
sudo systemctl restart gamer-wrapped
journalctl -u gamer-wrapped -f # View logs
```
