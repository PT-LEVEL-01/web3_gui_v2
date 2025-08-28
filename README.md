Analisis
File ini berisi panduan teknis untuk:

Kompilasi program Go menjadi executable.
Deploy node blockchain (node awal dan node tambahan).
Konfigurasi file dan parameter.
Cara menjalankan node (init, load, restart).
RPC testing menggunakan Postman.
Proses upgrade dan pembersihan data node.
Terjemahan & Langkah-langkah Detail
1. Kompilasi Proyek Menjadi Program Eksekusi
Langkah:

File utama (main) ada di: example/peer_root/firstPeer.go
Buka terminal, masuk ke folder peer_root.
Jalankan perintah berikut untuk kompilasi:
Jika berhasil, akan muncul file peer_root.exe.
2. Deploy Node Awal (Founder Node)
Langkah:

Buat folder baru bernama peer1.
Salin file peer_root.exe ke folder peer1.
Di dalam peer1, buat folder conf.
Masuk ke folder conf, buat file config.json dengan isi:
Penjelasan parameter:
AreaName: Nama jaringan, harus sama di semua node agar bisa terhubung.
ip: IP lokal/server. Untuk server publik, gunakan IP publik.
port: Port jaringan P2P.
WebAddr & WebPort: Untuk RPC.
RpcUser & RpcPassword: Username & password RPC.
NetType: Untuk jaringan publik gunakan release, untuk lokal gunakan not release.
AddrPre: Prefix alamat wallet.
Jalankan node dengan perintah:
Perintah ini akan menginisialisasi node dan menimpa data lama.
Akan muncul file keystore.key di folder conf (kunci node).
Akan muncul folder wallet di peer1 (data blockchain).
Node sudah aktif dan siap diuji dengan RPC.
3. Pengujian RPC
Langkah:

Buka aplikasi Postman.
Buat request POST ke: 127.0.0.1:3081/rpc
Di Headers, tambahkan:
user: test
password: testp
Di Body, pilih format raw, isi:
Klik Send untuk mengirim request.
4. Deploy Node Kedua (Node Tambahan)
Langkah:

Pastikan node awal (peer1) sudah berjalan.
Salin folder peer1 menjadi peer2.
Hapus folder wallet dan file keystore.key di peer2/conf.
Di peer2/conf, buat file nodeEntry.json:
Isi dengan alamat node yang akan dihubungi.
Jalankan node kedua dengan perintah:
Tanpa parameter.
Akan muncul folder wallet dan file keystore.key baru.
Node kedua berfungsi sebagai node biasa (tidak mining).
Untuk menambah node lain, ulangi langkah ini.
Bisa menambah password wallet saat start:
5. Node Restart (Load Data Lama)
Langkah:

Pastikan semua node dalam keadaan mati.
Pilih satu node yang bisa diakses (biasanya node mining).
Jalankan node dengan perintah:
Pastikan file nodeEntry.json di node lain berisi alamat node yang aktif.
Node lain jalankan tanpa parameter.
6. Node Hot Upgrade (Restart Node)
Langkah:

Minimal satu node tetap aktif.
Pastikan file nodeEntry.json berisi alamat node yang aktif.
Jalankan node yang ingin di-upgrade dengan perintah:
Tanpa parameter.
7. Bersihkan Data Node dan Restart
Langkah:

Matikan semua node.
Hapus folder wallet di semua node.
File keystore.key boleh disimpan.
Ikuti langkah deploy node dari awal.
8. Ringkasan Perintah Start Node
Founder node: ./peer_root.exe init (reset data, mulai dari awal).
Restart semua node dengan data lama: satu node pakai load, lainnya tanpa parameter.
Restart node (upgrade): tanpa parameter.
Jika ada bagian yang ingin dijelaskan lebih detail, silakan sebutkan!