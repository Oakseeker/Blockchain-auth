package main

import (
	"fmt"
	"crypto/sha256"
	"crypto/ecdsa"
	"crypto/elliptic" 
	crand "crypto/rand"
	//"log"
	mrand "math/rand"
	"os"
	"time"
	"strconv"
    "math/big"
	"bytes"
	"encoding/hex"
)


//Defining a struct type
type RAUTH struct {
	Gid, V, M ,RPW       []byte
	ID, PWD 			  string

}
var RApatient RAUTH
var RAphc RAUTH
var RAgovt RAUTH


var privatekey, err = ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
var pubkey ecdsa.PublicKey
var flowkey int =0 // 0= registration phase 1=login phase


//functional variable
var PvtKeyStr string
var CIDi, CIDj, wPpub, yPpub []byte
var wPpubdash, CIDjdash, C1dash, C2dash, yPpubdash []byte
var C1, C2, C3, C4 []byte
var C4dash []byte
var Ck, Cu, Cp []byte
var seco, seco1, seco2 int
var SKghc, SKphc, SKp []byte

func initialize() {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	pubkey = privatekey.PublicKey
	fmt.Printf("--ECC Parameters--\n")
	fmt.Printf(" Name: %s\n",elliptic.P256().Params().Name)
	fmt.Printf(" N: %x\n",elliptic.P256().Params().N)
	fmt.Printf(" P: %x\n",elliptic.P256().Params().P)
	fmt.Printf(" Gx: %v\n",elliptic.P256().Params().Gx)
	fmt.Printf(" Gy: %x\n",elliptic.P256().Params().Gy)
	fmt.Printf(" Bitsize: %x\n\n",elliptic.P256().Params().BitSize)
	fmt.Printf("\nPrivate key  %x", privatekey.D)
	fmt.Printf("\nPublic key (Alice) (%x,%x)", pubkey.X,pubkey.Y)
	fmt.Printf("\n")
}
func max(a, b int) int {
    if a < b {
        return b
    }
    return a
}
func timeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    fmt.Printf("%s took %s", name, elapsed)
}

func insertREFAUTHpatient(gid []byte, v []byte, m []byte) {
	RApatient.Gid = gid
	RApatient.M = m
	RApatient.V = v
}

func insertREFAUTHphc(gid []byte, v []byte, m []byte) {
	RAphc.Gid = gid
	RAphc.M = m
	RAphc.V = v
}

func insertREFAUTHgovt(gid []byte, v []byte, m []byte) {
	RAgovt.Gid = gid
	RAgovt.M = m
	RAgovt.V = v
}

func ScalMulandXor( scalar []byte ) ( xorstring []byte ){
	defer timeTrack(time.Now(), "scalarmul")
	var b bytes.Buffer
	var c bytes.Buffer
	ax, ay := elliptic.Curve.ScalarMult(elliptic.P256(), pubkey.X, pubkey.Y, scalar) //elliptic.P256().Params().Gx, elliptic.P256().Params().Gy, scalar)
	b.WriteString(ax.String())	
	c.WriteString(ay.String())
	newstr := b.String()
	newstr1 :=c.String()
	lenghtofstring := len(newstr)
	xorstring = make([]byte, lenghtofstring)
	for i := 0; i < lenghtofstring; i++ {
		xorstring[i] = newstr[i] ^ newstr1[i] 
		 }
	return  
}

func simplexor( xor1 string, xor2 string) ( xorstring []byte){
		
	lenghtofstring1 := len(xor1)
	lenghtofstring2 := len(xor2)
	a :=max(lenghtofstring1,lenghtofstring2)
	c := "00"
	data, err := hex.DecodeString(c)
	if err != nil {
  	  panic(err)
	}
	var b bytes.Buffer
	if a > lenghtofstring1 {
				dif := a- lenghtofstring1
				for j := 1; j <=(dif) ; j++ {
				b.WriteString(string(data))
				}
				b.WriteString(xor1)
				xor1 = b.String()
				lenghtofstring1 = len(xor1)
	} else {
				dif := a- lenghtofstring2
				for j := 1; j <=(dif) ; j++ {
				b.WriteString(string(data))
				}
				b.WriteString(xor2)
				xor2 = b.String()
				lenghtofstring1 = len(xor2)
		}
	xorstring = make([]byte, lenghtofstring1)
	for i := 0; i < lenghtofstring1; i++ {
		xorstring[i] = xor1[i] ^ xor2[i]
	 	}
	return
}

func hashing( b string) ( ha []byte){
	defer timeTrack(time.Now(), "hashing")
	h := sha256.New()
	h.Write([]byte(b))
	ha = h.Sum(nil)
	return
}

func Passwordchange () {
	//change password
	RPWCheck := hashing(RApatient.ID + RApatient.PWD) // This is RPW_star 
	fmt.Printf("\n RPW*​ =h​​ (ID​​ || PW​ ) = %x",RPWCheck)
	//Finding value for R_star: R​*​ = ID​​ ⊕ h​ (m​ || RPW​ )
	var b bytes.Buffer 
	b.WriteString(string(RPWCheck) )
	b.WriteString(string(RApatient.M)) 
	temp := hashing(b.String())
	// Xor it with ID
	b.Reset()
	b.WriteString(string(temp))
	Rstar := simplexor(b.String() , RApatient.ID)
	fmt.Printf("\n R​*​ = ID​​ ⊕ h​ (m​ || RPW​ ) = %x", Rstar)//Rstar calculated

	//V​.P = h​(R​ * || ID​ || RPW​*​ ).P​ pub 
	//This is same as V = V* where V*=h​(R​ * || ID​ || RPW​*​ ).x
	b.Reset()
	b.WriteString(string(Rstar)) 
	b.WriteString(RApatient.ID) 
	b.WriteString(string(RApatient.RPW)) 
	temp = hashing(b.String())
	b.Reset()
	b.WriteString(string(temp)) //completed h​(R​ * || ID​ || RPW​*​ )
	mul := []byte(PvtKeyStr)
	hash := temp
	m2 := new(big.Int)
	m3 := new(big.Int)
	m2.SetBytes(hash)
	m3.SetBytes(mul)
	Vstar := (new(big.Int).Mul(m2, m3)).Bytes()
	//Vstar := vstore.Bytes()
	fmt.Printf("\n \" Vstar \" : %x ",Vstar)
	// Verifies the equation
	res := bytes.Compare(RApatient.V, Vstar) 
    if res == 0 { 
		fmt.Println("\n !..ID and Password of the Patient is Correct..!") 
    } else { 
		fmt.Println("\n !..ID and Password of the Patient is Not Correct. Cannot update the Password..!") 
		os.Exit(0)
	}

	// for updation of password
	fmt.Println("\n Enter the new string: ")
    var first string    
	fmt.Scanln(&first)
	RApatient.PWD = first
	GIDNewPatient := RegCentre(RApatient.ID, RApatient.RPW, 1) //GID
	fmt.Printf("\n New GID stored in Patient: %x", GIDNewPatient)
	fmt.Println("\n Password change complete")
}


func RegCentre(IDfunc string, RPWfunc []byte, org int) (GIDtemp []byte) {
	// 1=patient 2=PHC and 3=GovtHos
	var  Vtemp , Mtemp []byte
	var temp []byte// to store hash value from the function
	//generate random number e
	randnum := mrand.Intn(1000)
	fmt.Println(" Random number generated:",randnum)
	//convert private key to string 
	PvtKeyStr = string(privatekey.D.Bytes()) 
	//fmt.Printf("\n Private key in string: %x", PvtKeyStr)
	var b bytes.Buffer 
	b.WriteString(PvtKeyStr) 
	//fmt.Printf("\n old String: %x", b.String()) 
	stunt := strconv.Itoa(randnum)  //strconv.FormatInt(e, 16) 
	// its adding the ascii value for concatination.
	//fmt.Println("\n Random no. in string", stunt)
	b.WriteString(stunt)
	//fmt.Printf("\n Private key concatenated with random no.(num in ascii form): %x", b.String()) 
	fmt.Printf("\n Finding \"m​ = h​(x​ || e)\" :")
	Mtemp = hashing(b.String())
	fmt.Printf("\n M is %x\n", Mtemp)


	//R​= ID ⊕ h​(m​ || ​RPW​)
	fmt.Println("\n Finding \"R​= ID ⊕ h​(m​ || ​RPW​)\" :  ")
	// convert mx to string 
	//mxstring := fmt.Sprintf("%x", M) // mx is a type big *int. this is converted to string
	b.Reset()//to reset the bytes.Buffer. Used for concatenation of the strings
	b.WriteString(string(RPWfunc))
	b.WriteString(string(Mtemp))
	temp = hashing(b.String())
	b.Reset()
	b.WriteString(string(temp))// completed hashing of h​(m​ || ​RPW​)
	//Xoring
	Rtemp := simplexor(b.String(), IDfunc) //string(xorstring)
	fmt.Printf("\n \" R​=ID ⊕ h​(m​ || ​RPW​) \" : %x", Rtemp)


	//V​ = h​(R​ || ID​ || RPW​).x
	fmt.Printf("\n Now Finding \"V​ = h​(R​ || ID​ || RPW​).x \":")
	b.Reset()//reset bytes.buffer
	b.WriteString(string(Rtemp))
	fmt.Printf("\n r is %x", b.String())
	b.WriteString(IDfunc)
	b.WriteString(string(RPWfunc))// concatenate all three
	temp = hashing(b.String())//start hashing it
	b.Reset()
	b.WriteString(string(temp))//Hashed is h​(R​ || ID​ || RPW​)
	fmt.Printf("\n \" V hash \" : %x ",b.String())
	// multiply with X i.e x is Private key
	mul := []byte(PvtKeyStr)
	hash := temp
	m2 := new(big.Int)
	m3 := new(big.Int)
	m2.SetBytes(hash)
	m3.SetBytes(mul)
	vstore := new(big.Int).Mul(m2, m3)
	Vtemp = vstore.Bytes()
	fmt.Printf("\n \" V​ = h​(R​ || ID​ || RPW​).x \" : %x ",Vtemp)


	fmt.Printf("\n\n Computing GID now.")
	sgg := sha256.New()
	sgg.Write(Vtemp)//start hashing it
	GIDtemp =sgg.Sum(nil)
	fmt.Printf("\n \" GID  \" %x ",GIDtemp)
	fmt.Println("\n All parameters are calculated and stored.")
	if org == 1 { //patient
		insertREFAUTHpatient(GIDtemp, Vtemp , Mtemp)
		fmt.Println("\n !..{GID​ i​ , V​ i​ , m​ i​ } of the Patient is updated..!") 
    } else if org == 2 { //PHC
		insertREFAUTHphc(GIDtemp, Vtemp, Mtemp)
		fmt.Println("\n !..{GID​j , V​j , mj​ } of the PHC is updated..! ")
	} else { //GovtHos
		insertREFAUTHgovt(GIDtemp, Vtemp, Mtemp)
		fmt.Println("\n !..{GID​k , V​k​ , m​k​ } of the GovtHosp is updated..! ")
	}
	return
}


func Patient() { 
	var temp []byte//storing hash from the function
	if(flowkey==0) {
		//Registration Phase
		RApatient.ID = "Nikhil1234"
		RApatient.PWD = "1234Nikhil"
		//RApatient.ID = "PatientAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQg8q/rcryQTAd0WOiPlgl6E0KFQwrffniIqZLr+e6pU8ahRANCAARTApuTeiaNtjWY7vaSnjHb+6/yvy3aK4Via0Zlm4au+QonDCbbE4oXNwm4QLKoHiQHsbBV3NCwMwuVmu8XpALi"
		//RApatient.PWD = "PatientHKoZIzj0CAQYIKoZIzj0DAQcDQgAEUwKbk3omjbY1mO72kp4x2/uv8r8t2iuFYmtGZZuGrvkKJwwm2xOKFzcJuECyqB4kB7GwVdzQsDMLlZrvF6QC4g=="
		RApatient.RPW = hashing(RApatient.ID + RApatient.PWD)  // store the hash back in RPW variable
		fmt.Printf("\n RPW: %x ", RApatient.RPW)
		fmt.Printf("\n ID and RPW is updated by the patient. RC is called for further process.")
	    //** RegCentre(ID,RPW)   //RegCentre()
		GIDPatient := RegCentre(RApatient.ID, RApatient.RPW, 1) //GID
		fmt.Printf("\n GID stored in Patient: %x", GIDPatient)
		return
		// calling the RC i.e ID and RPW are passed to RC
	}//flowkey =0

	//i.e login phase 
	if(flowkey==1) {
		//Login Phase
		RPWCheck := hashing(RApatient.ID + RApatient.PWD) // This is RPW_star 
		fmt.Printf("\n RPW*​ =h​​ (ID​​ || PW​ ) = %x",RPWCheck)
		//Finding value for R_star: R​*​ = ID​​ ⊕ h​ (m​ || RPW​ )
		var b bytes.Buffer 
		b.WriteString(string(RPWCheck) )
		b.WriteString(string(RApatient.M)) 
		temp = hashing(b.String())
		// Xor it with ID
		b.Reset()
		b.WriteString(string(temp))
		Rstar := simplexor(b.String() , RApatient.ID)
		fmt.Printf("\n R​*​ = ID​​ ⊕ h​ (m​ || RPW​ ) = %x", Rstar)//Rstar calculated

		//V​.P = h​(R​ * || ID​ || RPW​*​ ).P​ pub 
		//This is same as V = V* where V*=h​(R​ * || ID​ || RPW​*​ ).x
		b.Reset()
		b.WriteString(string(Rstar)) 
		b.WriteString(RApatient.ID) 
		b.WriteString(string(RApatient.RPW)) 
		temp = hashing(b.String())
		b.Reset()
		b.WriteString(string(temp)) //completed h​(R​ * || ID​ || RPW​*​ )
		mul := []byte(PvtKeyStr)
		hash := temp
		m2 := new(big.Int)
		m3 := new(big.Int)
		m2.SetBytes(hash)
		m3.SetBytes(mul)
		Vstar := (new(big.Int).Mul(m2, m3)).Bytes()
		//Vstar := vstore.Bytes()
		fmt.Printf("\n \" Vstar \" : %x ",Vstar)
		// Verifies the equation
		res := bytes.Compare(RApatient.V, Vstar) 
      	if res == 0 { 
			fmt.Println("\n !..ID and Password of the Patient is Correct..!") 
    	} else { 
			fmt.Println("\n !..ID and Password of the Patient is Not Correct..!") 
			os.Exit(0)
		} 

		//C​u​ = w.P​ pub ⊕ ​ h​(GID​​ || m​ || T​ ),
		//Patient​ Generates random number w
		randnumw := mrand.Intn(100)
		fmt.Println(" Random number generated:",randnumw)
		stunt := strconv.Itoa(randnumw)  //Random number  in string format is called stunt 
		// its adding the ascii value 
		b.Reset()
		b.WriteString(stunt)
		fmt.Println(" Random number generated: ",b.String())
		wPpub = ScalMulandXor([]byte(b.String()))
		fmt.Printf(" wPpub: %X ",wPpub)
		dtime := time.Now()
		seco = dtime.Second()
		fmt.Println("\n Current date and time is: ", dtime.String())
		fmt.Println("Current Second is T1: ", seco)
		b.Reset()
		b.WriteString(string(RApatient.Gid))
		b.WriteString(string(RApatient.M))
		stunt = strconv.Itoa(seco)// time T1 in ascii
		b.WriteString(stunt)
		hash = hashing(b.String())
		Cu = simplexor(string(hash) , string(wPpub) )
		fmt.Printf("\n Cu: %X",Cu)
		

		//CIDi ​= h​(w.P​ pub || ​T​1 ) ⊕ RPW​
		b.Reset()
		b.WriteString(string(wPpub))
		b.WriteString(stunt)//time T1
		hash = hashing(b.String())
		CIDi = simplexor(string(hash) , string(RApatient.RPW))
		fmt.Printf("\n CIDi : %x", CIDi )
		

		//C1= h​(CID​ || m​ || w.P​ pub​ )
		b.Reset()
		b.WriteString(string(CIDi))
		b.WriteString(string(RApatient.M))
		b.WriteString(string(wPpub))
		C1 = hashing(b.String())
		fmt.Printf("\n C1 : %x" , C1)
		fmt.Println("\n Parametes are calculates and sending to PHC server for further processing")
		//calling PHC
		PHC()
		
		fmt.Println("\n Patient receives mutual authentication message {M​2​ ,C4​ } from PHC server")
		//Patient receives mutual authentication message {M​2​ ,C4​ } from PHC server and computes
		//y.P​pub​’​ = Ck ⊕ h​(GID​k​ || m​k​ || T​3​ )
		b.Reset()
		b.WriteString(string(RAgovt.Gid))
		b.WriteString(string(RAgovt.M))
		stunt = strconv.Itoa(seco2)// time in ascii
		b.WriteString(stunt)
		b.WriteString(string(stunt))
		temp = hashing(b.String())
		yPpubdash = simplexor(string(Ck) , string(temp) )
		fmt.Printf("\n y.P​pub​’​ = Ck ⊕ h​(GID​k​ || m​k​ || T​3​ ): %x",yPpubdash)

		//SK​p​ = h​(y.P​pub​’​ || ​ w.P​pub​ || m​i​ || m​j​ || m​k​ )
		b.Reset()
		b.WriteString(string(yPpubdash))
		b.WriteString(string(wPpub))
		b.WriteString(string(RApatient.M))
		b.WriteString(string(RAphc.M))
		b.WriteString(string(RAgovt.M))
		SKp = hashing(b.String())
		fmt.Printf("\n SK​p​ = h​(y.P​pub​’​ || ​ w.P​pub​ || m​i​ || m​j​ || m​k​ ): %x",SKp)

		//C​4​’ = h​(SK​p​ || C​3​ || y.P​pub || w.P​pub​’)
		b.Reset()
		b.WriteString(string(SKp))
		b.WriteString(string(C3))
		b.WriteString(string(yPpub))
		b.WriteString(string(wPpubdash))
		C4dash = hashing(b.String())
		fmt.Printf("\n C​4​’ = h​(SK​p​ || C​3​ || y.P​pub || w.P​pub​’): %x", C4dash)

		//Compare C​4​’ with received C​4
		res = bytes.Compare(C4dash, C4) 
      	if res == 0 { 
			fmt.Println("\n !..C​4​’ matches with received C​4..!") 
			fmt.Println("\n Patient, PHC server and the GH server established the connection successfully")
    	} else { 
			fmt.Println("\n !..C​4​’ doesnt match with received C​4..!") 
			os.Exit(0)
		} 

		return	
	}//flowkey =1

}//function Patient

func PHC() {
	var temp []byte//storing hash from the function
	var b bytes.Buffer 
	if(flowkey == 0) {
		//Registration Phase
		RAphc.ID = "PHC1234"
		RAphc.PWD = "1234PHC"
		//RAphc.ID = "PhcMIGHAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQg8q/rcryQTAd0WOiPlgl6E0KFQwrffniIqZLr+e6pU8ahRANCAARTApuTeiaNtjWY7vaSnjHb+6/yvy3aK4Via0Zlm4au+QonDCbbE4oXNwm4QLKoHiQHsbBV3NCwMwuVmu8XpALi"
		//RAphc.PWD = "PhcMFkwHKoZIzj0CAQYIKoZIzj0DAQcDQgAEUwKbk3omjbY1mO72kp4x2/uv8r8t2iuFYmtGZZuGrvkKJwwm2xOKFzcJuECyqB4kB7GwVdzQsDMLlZrvF6QC4g=="
		RAphc.RPW = hashing(RAphc.ID + RAphc.PWD)  // store the hash back in RPW variable
		fmt.Printf("\n RPW: %x ", RAphc.RPW)
		fmt.Printf("\n ID and RPW is updated by the PHC. RC is called for further process.")
	    //** RegCentre(ID,RPW)   //RegCentre()
		GIDphc := RegCentre(RAphc.ID, RAphc.RPW, 2) //GID
		fmt.Printf("\n GID stored in PHC: %x", GIDphc)
		return
		// calling the RC i.e ID and RPW are passed to RC
	}//flowkey =0
	
	if(flowkey == 1) {
		//Login phase
		dtime := time.Now()
		seco1 = dtime.Second()
		result := seco1 - seco
		fmt.Println("\n Current date and time is: ", dtime.String())
		fmt.Println("Current Second is T1: ", seco)
		if result > 20 { 
			fmt.Println("\n Recieved message from Patient server late. Dropping M1") 
			os.Exit(0)
    	} else { 
			// w.P​ pub = C​u ⊕ h​(GID​ i​ || m​ i​ || T​ 1​ )
			b.WriteString(string(RApatient.Gid))
			b.WriteString(string(RApatient.M))
			stunt := strconv.Itoa(seco)// time in ascii
			b.WriteString(stunt)
			temp = hashing(b.String())
			wPpub = simplexor(string(temp) , string(Cu) )
			fmt.Printf("\n Updated wPup: %X", wPpub)


			//C​ p​ = w.P​ pub ⊕ ​ h​ 2​ (GID​ j​ || m​ j​ || T​ 2​ )
			b.Reset()
			b.WriteString(string(RAphc.Gid))
			b.WriteString(string(RAphc.M))
			stunt = strconv.Itoa(seco1)// time in ascii
			b.WriteString(stunt)
			temp = hashing(b.String())
			Cp = simplexor(string(temp) , string(wPpub) )
			fmt.Printf("\n C​p​ = w.P​ pub ⊕ ​h​(GID​j​ || mj​ || T2​ ): %X", Cp)

			//CID​j​ = h​(C​u || ​C​p​ || w.P​pub​ )
			b.Reset()
			b.WriteString(string(Cu))
			b.WriteString(string(Cp))
			b.WriteString(string(wPpub))
			CIDj = hashing(b.String())
			fmt.Printf("\n CID​j​ = h​(C​u || ​C​p​ || w.P​pub​ ): %X", CIDj)

			//C​2​ = h​(w.P​pub || ​C​1​ || CID​i || CID​i || ​T​1​ || T​2​ )
			b.Reset()
			b.WriteString(string(wPpub))
			b.WriteString(string(C1))
			b.WriteString(string(CIDi))
			b.WriteString(string(CIDj))
			stunt = strconv.Itoa(seco1)// time T2 in ascii
			b.WriteString(stunt)
			stunt = strconv.Itoa(seco)// time T1 in ascii
			b.WriteString(stunt)
			temp = hashing(b.String())
			C2 = hashing(b.String())
			fmt.Printf("\n C​2​ = h​(w.P​pub || ​C​1​ || CID​i || CID​i || ​T​1​ || T​2​ ): %X", C2)
			fmt.Println("\n Sends {M​ 1​ , C​ p​ , C​ 2​ , T​ 2​ } to the GH server ")

			//calling GH server
			GovtHosp()

			fmt.Println("\n Return with M2 = {C​3​ , C​k​ , T​3​ }​ ​ to PHC server")
			//return with M2 = {C​ 3​ , C​ k​ , T​ 3​ }​ ​ to PHC server.
			//y.P​pub​’​ ​= ​​C​k​ ⊕ h​(GID​k​ || m​k​ || T​3​ )
			b.Reset()
			b.WriteString(string(RAgovt.Gid))
			b.WriteString(string(RAgovt.M))
			stunt = strconv.Itoa(seco2)// time T3 in ascii
			b.WriteString(stunt)
			b.WriteString(string(stunt))
			temp = hashing(b.String())
			yPpubdash = simplexor( string(Ck), string(temp))
			fmt.Printf("\n y.P​pub​’ =​ C​k​ ⊕ h​(GID​k​ || m​k​ || T​3​ ): %x",yPpubdash)
			
			//SK​phc​ = h​(y.P​pub​’​ || ​w.P​pub​’ || m​i​ || m​j​ || m​k​ )
			b.Reset()
			b.WriteString(string(yPpubdash))
			b.WriteString(string(wPpubdash))
			b.WriteString(string(RApatient.M))
			b.WriteString(string(RAphc.M))
			b.WriteString(string(RAgovt.M))
			SKphc = hashing(b.String())
			fmt.Printf("\n SK​phc​ = h​(y.P​pub​’​ || ​w.P​pub​’ || m​i​ || m​j​ || m​k​ ): %x ", SKphc)

			//C​4 = h​(SK​phc​ || C​3​ || y.P​pub || w.P​pub​ ’)
			b.Reset()
			b.WriteString(string(SKphc))
			b.WriteString(string(C3))
			b.WriteString(string(yPpub))
			b.WriteString(string(wPpubdash))
			C4 = hashing(b.String())
			fmt.Printf("\n C​4 = h​(SK​phc​ || C​3​ || y.P​pub || w.P​pub​ ’): %x ", C4)

			//Sends {M​ 2​ , C​ 4​ } to the patient
			fmt.Println("\n Sends {M​2​ , C​4​ } to the patient")
			return
		}
	}//flowkey=1

}//PHC

func GovtHosp() {
	var temp []byte//storing hash from the function
	var b bytes.Buffer 
	if(flowkey==0) {
		//Registration Phase
		RAgovt.ID = "GovtHosp123"
		RAgovt.PWD = "123GovtHosp"
		//RAgovt.ID = "GovtHosAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQg8q/rcryQTAd0WOiPlgl6E0KFQwrffniIqZLr+e6pU8ahRANCAARTApuTeiaNtjWY7vaSnjHb+6/yvy3aK4Via0Zlm4au+QonDCbbE4oXNwm4QLKoHiQHsbBV3NCwMwuVmu8XpALi"
		//RAgovt.PWD = "GovtHosHKoZIzj0CAQYIKoZIzj0DAQcDQgAEUwKbk3omjbY1mO72kp4x2/uv8r8t2iuFYmtGZZuGrvkKJwwm2xOKFzcJuECyqB4kB7GwVdzQsDMLlZrvF6QC4g=="
		RAgovt.RPW = hashing(RAgovt.ID + RAgovt.PWD)  // store the hash back in RPW variable
		fmt.Printf("\n RPW: %x ", RAgovt.RPW)
		fmt.Printf("\n ID and RPW is updated by the Govt Hospital. RC is called for further process.")
	    //** RegCentre(ID,RPW)   //RegCentre()
		GIDgov := RegCentre(RAgovt.ID, RAgovt.RPW, 3) //GID
		fmt.Printf("\n GID stored in Govt Hospital: %x", GIDgov)
		return
		// calling the RC i.e ID and RPW are passed to RC
		
	}//flowkey =
	//login part
	if(flowkey == 1) {
		//Login phase
		dtime := time.Now()
		seco2 = dtime.Second()
		result := seco2 - seco1
		fmt.Println("\n Current date and time is: ", dtime.String())
		fmt.Println("\n Current Second is T1: ", seco2)
		if result > 20 { 
			fmt.Println("\n Recieved message from PHC server late. Dropping M1") 
			os.Exit(0)
    	} else {
			
			//w.P​pub​ ’​ = C​u ⊕ h​(GID​i​ || m​i​ || T​1​ )
			b.WriteString(string(RApatient.Gid))
			b.WriteString(string(RApatient.M))
			stunt := strconv.Itoa(seco)// time T1 in ascii
			b.WriteString(stunt)
			temp = hashing(b.String())
			wPpubdash = simplexor(string(temp) , string(Cu) )
			fmt.Printf("\n wPup’​ = C​u ⊕ h​(GID​i​ || m​i​ || T​1​ ): %X", wPpubdash)

			
			//CID​ j​ ’​ = h(C​u || C​p​ || w.P​pub​ )
			b.Reset()
			b.WriteString(string(Cu))
			b.WriteString(string(Cp))
			b.WriteString(string(wPpub))
			CIDjdash = hashing(b.String())
			fmt.Printf("\n CID​ j​ ’​ = h(C​u || C​p​ || w.P​pub​ ): %x", CIDjdash)

			// C​1’ = h​(CID​i’ || m​i​ || w.P​ pub​') 
			b.Reset()
			b.WriteString(string(CIDi))
			b.WriteString(string(RApatient.M))
			b.WriteString(string(wPpubdash))
			C1dash = hashing(b.String())
			fmt.Printf("\n C​1’ = h​(CID​i’ || m​i​ || w.P​ pub​') : %x", C1dash)

			//C2’= h(w.P​ pub​’​ || ​C1​ || CIDi || CIDj’​ || T1 || T2​ )
			b.Reset()
			b.WriteString(string(wPpubdash))
			b.WriteString(string(C1))
			b.WriteString(string(CIDi))
			b.WriteString(string(CIDjdash))
			stunt = strconv.Itoa(seco)// time T1 in ascii
			b.WriteString(stunt)
			stunt = strconv.Itoa(seco1)// time T2 in ascii
			b.WriteString(stunt)
			C2dash = hashing(b.String())
			fmt.Printf("\n C2’= h(w.P​ pub​’​ || ​C1​ || CIDi || CIDj’​ || T1 || T2​ ): %X", C2dash)
			res := bytes.Compare(C2dash, C2) 
      		if res == 0 { 
				fmt.Println("\n C​2​’ successfully matches with received C2​ ")
			} else { 
				fmt.Println("\n C​2​’ Doesn't matches with received C2​ ")
				os.Exit(0)
			} 

			//Ck​ = y.Ppub ⊕ ​ h(GIDk​ || m​k​ || T3​ )
			b.Reset()
			b.WriteString(string(RAgovt.Gid))
			b.WriteString(string(RAgovt.M))
			stunt = strconv.Itoa(seco2)// time T3 in ascii
			b.WriteString(stunt)
			temp = hashing(b.String())
			//Patient​ Generates random number y
			randnumw := mrand.Intn(100)
			fmt.Println(" Random number generated:",randnumw)
			stunt = strconv.Itoa(randnumw)  //Random number  in string format is called stunt 
			// its adding the ascii value 
			b.Reset()
			b.WriteString(stunt)
			fmt.Println(" Random number generated: ",b.String())
			yPpub = ScalMulandXor([]byte(b.String()))
			fmt.Printf(" y.Ppub: %X ",yPpub)
			Ck = simplexor(string(temp) , string(yPpub) )
			fmt.Printf("\n Ck: %X",Ck)

			//SK​ghc​ = h​(y.P​pub || ​w.P​pub​’ || m​i​ || m​j​ || m​k​ )
			b.Reset()
			b.WriteString(string(yPpub))
			b.WriteString(string(wPpubdash))
			b.WriteString(string(RApatient.M))
			b.WriteString(string(RAphc.M))
			b.WriteString(string(RAgovt.M))
			SKghc = hashing(b.String())
			fmt.Printf("\n SK​ghc​ = h​(y.P​pub || ​w.P​pub​’ || m​i​ || m​j​ || m​k​ ): %x ", SKghc)

			//C​3​ = h​(SK​ghc​ || T​3​ || y.P​pub​ )
			b.Reset()
			b.WriteString(string(SKghc))
			stunt = strconv.Itoa(seco2)// time T3 in ascii
			b.WriteString(stunt)
			b.WriteString(string(yPpub))
			C3 = hashing(b.String())
			fmt.Printf("\n C​3​ = h​(SK​ghc​ || T​3​ || y.P​pub​ ):%x", C3)
			fmt.Println("\n GH server sends a mutual authentication message M2= {C3​, Ck​, T3} to PHC server.")
			return

		} //else
	}//flowkey=1
}//Govt hospital

func main() {
	// insert: func(gid []byte, m []byte, v []byte) {
	// 	Gidst = gid 
	// 	Vst =  v
	// 	Mst = m
	
	// }
	fmt.Printf("Initializing the parameters: \n")
	initialize()
	flowkey = 0
	fmt.Printf("\n Starting Registration Phase: \n")
	Patient()
	fmt.Printf("\n Registration Phase for Patient is completed: \n")
	PHC()
	fmt.Printf("\n Registration Phase for PHC is completed: \n")
	GovtHosp()
	fmt.Printf("\n Registration Phase for Govt Hospital is completed: \n")
	fmt.Printf("\n\n\n Start Login Phase.Patient to PHC")
	flowkey =1
	Patient()

	
}

//Example/ Extra Codes
//Declaring a variable of a `struct` type
		// var p Person // // All the struct fields are initialized with their zero value
		// fmt.Println(p)

		// // Declaring and initializing a struct using a struct literal
		// p1 := Person{ 4, 3, 6}
		// fmt.Println("Person1: ", p1)

//for exor code
		// func main() {
		// 	string1 := "WhatAmIDoingHere?"
		// 	string2 := "thisisjustatest !"

		// 	n := len(string1)
		// 	b := make([]byte, n)
		// 	for i := 0; i < n; i++ {
		// 		b[i] = string1[i] ^ string2[i]
		// 	}
		// 	fmt.Printf("%x\n", string(b))
		// }

//some code for concatination
	//b.WriteString(e)
		//b10 := strconv.AppendInt(privatekey.D.Bytes(), e )
		//b10 := []byte(privatekey.D.Bytes())
		//b12 := []byte(e.Bytes())
		//sgg := sha256.New()
		//sgg.Write([]byte(example1))
		//fmt.Println("\n example 1 ", example1) //sgg.Sum(nil)
		///sgg2 := sha256.New()
		//sgg2.Write([]byte(example2))
		//fmt.Printf("\n example 2 %x", b10 ) //

// scalar multiplication code

		// def scalarmult(P, n):
		//        if n == 1:
		//            return P # identity element
		//        if n & 1:
		//            # n is odd we add point
		//            return point_add(P, scalarmult(P, n - 1)
		//        else:
		//            return scalarmult(point_double(P), n >> 1) # we double point and multiply by n/2


//curve param
		// type CurveParams struct {
		//     // P is the prime used in the secp256k1 field.
		//     P   *big.Int

		//     // N is the order of the secp256k1 curve group generated by the base point.
		//     N   *big.Int

		//     // Gx and Gy are the x and y coordinate of the base point, respectively.
		//     Gx, Gy *big.Int

		//     // BitSize is the size of the underlying secp256k1 field in bits.
		//     BitSize int

		//     // H is the cofactor of the secp256k1 curve.
		//     H   int

		//     // ByteSize is simply the bit size / 8 and is provided for convenience
		//     // since it is calculated repeatedly.
		//     ByteSize int
		// }

//Dont require scalar multiplication with P. 28/04/2020
		//mx, my := elliptic.Curve.ScalarMult(elliptic.P256(), elliptic.P256().Params().Gx, elliptic.P256().Params().Gy, hashed)

//Convert Big int to string 
		//bigint := big.NewInt(123)
		//bigstr := bigint.String()

//Simplified exor
	// func simplexor( xor1 string, xor2 string) ( xorstring []byte){
	// 	//didnt take care of unequal length
	// 	lenghtofstring1 := len(xor1)
	// 	xorstring = make([]byte, lenghtofstring1)
	// 	for i := 0; i < lenghtofstring1; i++ {
	// 		xorstring[i] = xor1[i] ^ xor2[i]
	// 	 	}
	// 	return
	// }