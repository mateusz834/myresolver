package myresolver

import (
	"encoding/binary"
	"io"
	"net"
	"net/netip"
	"strconv"
	"time"

	"github.com/mateusz834/dnsmsg"
)

func ListenUDPDNS(srcAddr netip.AddrPort, db *IPDB, callback func(q dnsmsg.Question[dnsmsg.ParserName], srcAddr netip.Addr)) error {
	udpConn, err := net.ListenUDP("udp", net.UDPAddrFromAddrPort(srcAddr))
	if err != nil {
		return err
	}
	defer udpConn.Close()

	buf := make([]byte, 512)
	resBuf := make([]byte, 0, 512)
	for {
		n, addr, err := udpConn.ReadFromUDPAddrPort(buf)
		if err != nil {
			return err
		}

		response := handleResponse(addr.Addr(), buf[:n], resBuf[:0], db, callback)
		if response != nil {
			if _, err := udpConn.WriteToUDPAddrPort(response, addr); err != nil {
				return err
			}
		}
	}
}

func ListenTCPDNS(srcAddr netip.AddrPort, db *IPDB, callback func(q dnsmsg.Question[dnsmsg.ParserName], srcAddr netip.Addr)) error {
	tcpConn, err := net.ListenTCP("tcp", net.TCPAddrFromAddrPort(srcAddr))
	if err != nil {
		return err
	}
	defer tcpConn.Close()

	for {
		conn, err := tcpConn.AcceptTCP()
		if err != nil {
			// TODO: out of file descriptors handle.
			return err
		}

		go func(conn *net.TCPConn) {
			defer conn.Close()

			buf := make([]byte, 512)
			resBuf := make([]byte, 0, 512)

			for {
				if err := conn.SetDeadline(time.Now().Add(time.Second * 5)); err != nil {
					return
				}

				if _, err := io.ReadFull(conn, buf[:2]); err != nil {
					return
				}

				length := binary.BigEndian.Uint16(buf[:2])
				if len(buf) < int(length) {
					buf = make([]byte, length)
				}

				if _, err := io.ReadFull(conn, buf[:length]); err != nil {
					return
				}

				addr, _ := netip.AddrFromSlice(conn.LocalAddr().(*net.TCPAddr).IP)
				if msg := handleResponse(addr, buf[:length], resBuf, db, callback); msg != nil {
					m := append(binary.BigEndian.AppendUint16(nil, uint16(len(msg))), msg...)
					if _, err := conn.Write(m); err != nil {
						return
					}
				} else {
					return
				}
			}
		}(conn)
	}

}

func handleResponse(addr netip.Addr, msg []byte, resBuf []byte, db *IPDB, callback func(q dnsmsg.Question[dnsmsg.ParserName], srcAddr netip.Addr)) []byte {
	if callback == nil {
		callback = func(q dnsmsg.Question[dnsmsg.ParserName], srcAddr netip.Addr) {}
	}

	p, _ := dnsmsg.NewParser(msg)
	hdr, err := p.Header()
	if err != nil {
		return nil
	}

	if hdr.QDCount != 1 || !hdr.Flags.Query() ||
		hdr.Flags.OpCode() != dnsmsg.OpCodeQuery ||
		hdr.Flags.RCode() != dnsmsg.RCodeSuccess {
		return reject(hdr, resBuf)
	}

	q, err := p.Question()
	if err != nil {
		return malformed(hdr, resBuf)
	}

	if q.Class != dnsmsg.ClassIN {
		return reject(hdr, resBuf)
	}

	var resFlags dnsmsg.Flags
	resFlags.SetResponse()
	resFlags.SetRCode(dnsmsg.RCodeSuccess)
	resFlags.SetBit(dnsmsg.BitAA, true)
	if hdr.Flags.Bit(dnsmsg.BitRD) {
		resFlags.SetBit(dnsmsg.BitRD, true)
	}

	name := q.Name.AsRawName()

	b := dnsmsg.StartBuilder(resBuf, hdr.ID, resFlags)
	b.Question(dnsmsg.Question[dnsmsg.RawName]{
		Name:  name,
		Class: q.Class,
		Type:  q.Type,
	})
	b.StartAnswers()

	if addr.Is4In6() {
		addr = netip.AddrFrom4(addr.As4())
	}

	switch q.Type {
	case dnsmsg.TypeA:
		if addr.Is4() {
			b.ResourceA(dnsmsg.ResourceHeader[dnsmsg.RawName]{
				Name:  name,
				Type:  dnsmsg.TypeA,
				Class: dnsmsg.ClassIN,
				TTL:   60,
			}, dnsmsg.ResourceA{
				A: addr.As4(),
			})
		} else if addr.Is6() {
			a6 := addr.As16()
			for i := 0; i < 16; i += 4 {
				b.ResourceA(dnsmsg.ResourceHeader[dnsmsg.RawName]{
					Name:  name,
					Type:  dnsmsg.TypeA,
					Class: dnsmsg.ClassIN,
					TTL:   60,
				}, dnsmsg.ResourceA{
					A: ([4]byte)(a6[i : i+4]),
				})
			}
		}
		callback(q, addr)
	case dnsmsg.TypeAAAA:
		b.ResourceAAAA(dnsmsg.ResourceHeader[dnsmsg.RawName]{
			Name:  name,
			Type:  dnsmsg.TypeAAAA,
			Class: dnsmsg.ClassIN,
			TTL:   60,
		}, dnsmsg.ResourceAAAA{
			AAAA: addr.As16(),
		})
		callback(q, addr)
	case dnsmsg.TypeTXT:
		txt := make([]byte, 0, 128)
		txt = append(txt, "Query resolved by: '"...)
		txt = addr.AppendTo(txt)
		txt = append(txt, '\'')

		b.ResourceTXT(dnsmsg.ResourceHeader[dnsmsg.RawName]{
			Name:  name,
			Type:  dnsmsg.TypeTXT,
			Class: dnsmsg.ClassIN,
			TTL:   60,
		}, dnsmsg.ResourceTXT{
			TXT: [][]byte{txt},
		})

		if db != nil {
			asn, desc, err := db.LookupIP(addr)
			if err == nil && asn != 0 && desc != "" {
				txt = append(txt[:0], "ASN "...)
				txt = strconv.AppendUint(txt, asn, 10)
				txt = append(txt, ": '"...)
				txt = append(txt, desc...)
				txt = append(txt, '\'')

				b.ResourceTXT(dnsmsg.ResourceHeader[dnsmsg.RawName]{
					Name:  name,
					Type:  dnsmsg.TypeTXT,
					Class: dnsmsg.ClassIN,
					TTL:   60,
				}, dnsmsg.ResourceTXT{
					TXT: [][]byte{txt},
				})
			}
		}

	default:
	}

	return b.Bytes()
}

func reject(hdr dnsmsg.Header, resBuf []byte) []byte {
	var resFlags dnsmsg.Flags
	resFlags.SetResponse()
	resFlags.SetRCode(dnsmsg.RCodeRefused)
	if hdr.Flags.Bit(dnsmsg.BitRD) {
		resFlags.SetBit(dnsmsg.BitRD, true)
	}
	b := dnsmsg.StartBuilder(resBuf, hdr.ID, resFlags)
	return b.Bytes()
}

func malformed(hdr dnsmsg.Header, resBuf []byte) []byte {
	var resFlags dnsmsg.Flags
	resFlags.SetResponse()
	resFlags.SetRCode(dnsmsg.RCodeFormatError)
	if hdr.Flags.Bit(dnsmsg.BitRD) {
		resFlags.SetBit(dnsmsg.BitRD, true)
	}
	b := dnsmsg.StartBuilder(resBuf, hdr.ID, resFlags)
	return b.Bytes()
}
