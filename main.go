package main

import (
    "net/http"
    "fmt"
    "encoding/json"
    "io"
    "os"
    "strconv"
    "os/exec"
    "math/rand"
    "time"
    )

type request struct {
        TypeID int `json:"typeID"`
}

type proof struct {
        Context []string `json:"@context"`
        Type string `json:"type"`
        Created string `json:"created"`
        Domain string `json:"domain"`
        Nonce string `json:"nonce"`
        ProofOfPurpose string `json:"proofPurpose"`
        VerificationMethod string `json:"verificationMethod"`
        ProofValue string `json:"proofValue"`
}

type VC struct {
        TypeID int `json:"typeID"`
        Type string `json:"type"`
        ID string `json:"id"`
        Proof proof `json:"proof"`
}

func createVC(typeID int, vc *VC)  {
	vc.TypeID = typeID
	switch(typeID) {
	case 1:
		vc.Type = "birth certificate"
	case 2:
		vc.Type = "passport"
	case 3:
		vc.Type = "driving license"
	case 4:
		vc.Type = "diploma"
	default:
		vc.Type = "invalid"
	}
	s1 := rand.NewSource(time.Now().UnixNano())
        r1 := rand.New(s1)
        randomNum := r1.Intn(100000)
	vc.ID = strconv.Itoa(randomNum)
}

func createDID(){
        s1 := rand.NewSource(time.Now().UnixNano())
        r1 := rand.New(s1)
        randomNum := r1.Intn(100)
        refName := "issuer-key-"+strconv.Itoa(randomNum)
        cmd := exec.Command("algoid", "create", refName)
        out, err := cmd.Output()
        if err != nil {
                fmt.Println(err)
        }
        f, err := os.Create("did.ref")

        if err != nil {
                fmt.Println(err)
        }
        defer f.Close()

        _, err = f.WriteString(refName)
        if err != nil {
                fmt.Println(err)
        }

        //sync or publish
        cmd = exec.Command("algoid", "sync", refName)
        out, err = cmd.Output()
        if err != nil {
                fmt.Println(err)
        }
        fmt.Println(string(out))
}


func createSignature() {

        content, err := os.ReadFile("did.ref")
        if (err != nil){
                //create did
                createDID()
                content, err = os.ReadFile("did.ref")
        }
        if (err != nil){
                fmt.Println(err)
        }
        //did.key contains referencename
        referenceName := string(content)
        //get signature using algoid - algoid sign <referencename> -i "vp"
        ///command := "algoid sign "+referencename+" -i \"vp\""
        cmd := exec.Command("algoid", "sign", referenceName, "-i", "vp")
        out, err := cmd.Output()
        if err != nil {
                fmt.Println(err)
        }
        f, err := os.Create("issuer.sign")

        if err != nil {
        fmt.Println(err)
        }
        defer f.Close()

        _, err = f.WriteString(string(out))
        if err != nil {
       fmt.Println(err)
        }
}


func signVC(vc *VC){
	var sign proof
	//Check if there is a signature
	content, err := os.ReadFile("issuer.sign")
        if (err != nil){
                //create signature 
                createSignature()
                content, err = os.ReadFile("issuer.sign")
	}
        if (err != nil) {
                fmt.Println(err)
        }
	err = json.Unmarshal(content, &sign)
        if (err != nil){
                fmt.Println(err)
        }
	vc.Proof = sign
}

func IssueVC(w http.ResponseWriter, req *http.Request) {
	var r request
        var vc VC 
	b, _ := io.ReadAll(req.Body)
         err := json.Unmarshal(b, &r)
         if err != nil {
                fmt.Println(err)
         }
         docTypeID := r.TypeID

	 //Create VC
	 createVC(docTypeID, &vc)
	 //Sign VC
	 signVC(&vc)
	 //Send Back
	 w.Header().Set("Content-Type", "application/json")
	 json.NewEncoder(w).Encode(vc)
}

func main() {
    //Handlers    
    http.HandleFunc("/idme/issuer/issue/v1/vc", IssueVC)
    fmt.Printf("Running on port 8090...");
    http.ListenAndServe("131.159.209.212:8090", nil)
}

