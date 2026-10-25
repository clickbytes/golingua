// Package interlingua contains independently authored A1.11 case-order meanings.
package interlingua

import (
 "fmt"
 "strconv"
 "strings"
 "github.com/clickbytes/golingua/internal/ids"
 "github.com/clickbytes/golingua/internal/meaning"
)

type caseOrderSpec struct { index int; predicate, roles, temporal, gap string }

// Semantic specs omit all bilingual surface text. Role mentions create distinct
// referent nodes even when their ConceptIDs match; black is a separate untimed
// state rooted in the graph and bound to the exact marked referent.
var caseOrderSpecs = []caseOrderSpec{
	{index: 1, predicate: "seeing", roles: "experiencer=cat:one:black,stimulus=dog:one", temporal: "present", gap: ""},
	{index: 2, predicate: "on_support", roles: "figure=book:one:black,landmark=table:one", temporal: "present", gap: "CORE has supports_physical, but no on-support assertion with figure/landmark roles in Meaning."},
	{index: 3, predicate: "seeing", roles: "experiencer=teacher:one,stimulus=dog:one:black", temporal: "present", gap: ""},
	{index: 4, predicate: "seeing", roles: "experiencer=doctor:one,stimulus=book:one:black", temporal: "present", gap: ""},
	{index: 5, predicate: "seeing", roles: "experiencer=child:one,stimulus=window:one:black", temporal: "present", gap: ""},
	{index: 6, predicate: "seeing", roles: "experiencer=dog:one,stimulus=door:one:black", temporal: "present", gap: ""},
	{index: 7, predicate: "seeing", roles: "experiencer=doctor:one,stimulus=cat:many:black", temporal: "present", gap: ""},
	{index: 8, predicate: "seeing", roles: "experiencer=cat:many:black,stimulus=dog:one", temporal: "present", gap: ""},
	{index: 9, predicate: "seeing", roles: "experiencer=teacher:one,stimulus=bird:many:black", temporal: "present", gap: ""},
	{index: 10, predicate: "seeing", roles: "experiencer=bird:many:black,stimulus=dog:one", temporal: "present", gap: ""},
	{index: 11, predicate: "seeing", roles: "experiencer=mother:one,stimulus=bag:one:black", temporal: "present", gap: ""},
	{index: 12, predicate: "hearing", roles: "experiencer=father:one,stimulus=cat:one:black", temporal: "present", gap: ""},
	{index: 13, predicate: "hearing", roles: "experiencer=dog:one:black,stimulus=cat:one", temporal: "present", gap: ""},
	{index: 14, predicate: "open_state", roles: "bearer=bag:one:black", temporal: "present", gap: "CORE has opening_object EVENT; open-state is a distinct untimed state and is not encoded."},
	{index: 15, predicate: "seeing", roles: "experiencer=doctor:one,stimulus=cat:one", temporal: "present", gap: ""},
	{index: 16, predicate: "hearing", roles: "experiencer=mother:one,stimulus=dog:one", temporal: "present", gap: ""},
	{index: 17, predicate: "seeing", roles: "experiencer=dog:one,stimulus=teacher:one", temporal: "present", gap: ""},
	{index: 18, predicate: "hearing", roles: "experiencer=cat:one,stimulus=doctor:one", temporal: "present", gap: ""},
	{index: 19, predicate: "seeing", roles: "experiencer=child:one,stimulus=mother:one", temporal: "present", gap: ""},
	{index: 20, predicate: "hearing", roles: "experiencer=teacher:one,stimulus=father:one", temporal: "present", gap: ""},
	{index: 21, predicate: "seeing", roles: "experiencer=doctor:one,stimulus=cat:one:black", temporal: "present", gap: ""},
	{index: 22, predicate: "hearing", roles: "experiencer=mother:one,stimulus=dog:one:black", temporal: "present", gap: ""},
	{index: 23, predicate: "seeing", roles: "experiencer=child:one,stimulus=book:one:black", temporal: "present", gap: ""},
	{index: 24, predicate: "seeing", roles: "experiencer=dog:one,stimulus=doctor:many", temporal: "present", gap: ""},
	{index: 25, predicate: "hearing", roles: "experiencer=cat:one,stimulus=teacher:many", temporal: "present", gap: ""},
	{index: 26, predicate: "seeing", roles: "experiencer=doctor:many,stimulus=dog:one", temporal: "present", gap: ""},
	{index: 27, predicate: "hearing", roles: "experiencer=child:many,stimulus=cat:one", temporal: "present", gap: ""},
	{index: 28, predicate: "helping", roles: "helper=mother:one,beneficiary=child:one", temporal: "present", gap: ""},
	{index: 29, predicate: "helping", roles: "helper=mother:one,beneficiary=child:one", temporal: "present", gap: ""},
	{index: 30, predicate: "thanking", roles: "thanker=child:one,recipient=mother:one", temporal: "present", gap: ""},
	{index: 31, predicate: "giving", roles: "giver=teacher:one,recipient=child:one,theme=book:one", temporal: "present", gap: ""},
	{index: 32, predicate: "giving", roles: "giver=teacher:one,recipient=child:one,theme=book:one", temporal: "present", gap: ""},
	{index: 33, predicate: "giving", roles: "giver=teacher:one,recipient=child:one,theme=book:one", temporal: "present", gap: ""},
	{index: 34, predicate: "giving", roles: "giver=mother:one,recipient=doctor:one,theme=key_tool:one", temporal: "present", gap: ""},
	{index: 35, predicate: "giving", roles: "giver=father:one,recipient=mother:one,theme=bag:one", temporal: "present", gap: ""},
	{index: 36, predicate: "helping", roles: "helper=child:many,beneficiary=teacher:one", temporal: "present", gap: ""},
	{index: 37, predicate: "using_object", roles: "user=doctor:one,tool=key_tool:one", temporal: "present", gap: "CORE has no using_object predicate with user/tool roles; the tool concept alone cannot encode the action."},
	{index: 38, predicate: "cutting_object", roles: "agent=teacher:one,material=bread:mass,instrument=knife_utensil:one", temporal: "present", gap: "CORE cutting_object omits the instrument role; knife_utensil is only a referent concept."},
	{index: 39, predicate: "writing_instrument", roles: "writer=mother:one,instrument=pencil:one", temporal: "present", gap: "No writing-with-instrument predicate/role is licensed. writing_document would add durable-document meaning."},
	{index: 40, predicate: "conversation", roles: "speaker=teacher:one,partner=doctor:one", temporal: "present", gap: "No reciprocal conversation predicate is licensed; speaking would assert a different one-way event."},
	{index: 41, predicate: "accompaniment", roles: "figure=child:one,companion=mother:one", temporal: "present", gap: "No accompaniment relation/predicate is licensed; another social or ownership assertion would change meaning."},
	{index: 42, predicate: "looking", roles: "agent=dog:one,target=cat:one", temporal: "present", gap: ""},
}

// CaseOrderExpansion returns a licensed present-frame graph. Known CORE gaps
// have no substitute graph: callers receive ok=false for those full meanings.
// IDs are a1.11.case.NNN.en|pl; child role variants use .offspring.
func CaseOrderExpansion(id string) (IdentificationMeaning, bool, error) {
 parts:=strings.Split(id,".")
 if len(parts)<5 || len(parts)>6 || parts[0]!="a1" || parts[1]!="11" || parts[2]!="case" || (parts[4]!="en" && parts[4]!="pl") { return IdentificationMeaning{},false,nil }
 n,err:=strconv.Atoi(parts[3]); if err!=nil || n<1 || n>42 { return IdentificationMeaning{},false,nil }
 variant:=""; if len(parts)==6 { if parts[5]!="offspring" { return IdentificationMeaning{},false,nil }; variant=parts[5] }
 var spec caseOrderSpec; for _,s:=range caseOrderSpecs { if s.index==n {spec=s;break} }
 if spec.index==0 || spec.gap!="" { return IdentificationMeaning{},false,nil }
 roles:=strings.Split(spec.roles,","); hasChild:=false
 for _,raw:=range roles { if strings.Contains(raw,"=child:") {hasChild=true} }
 if variant!="" && !hasChild { return IdentificationMeaning{},false,nil }
 predicate,err:=ids.NewConceptID(spec.predicate); if err!=nil {return IdentificationMeaning{},false,fmt.Errorf("case predicate: %w",err)}
 nodes:=make([]meaning.Node,0,len(roles)*2+1); participants:=make([]meaning.Participant,0,len(roles)); blackStates:=make([]meaning.Node,0)
 for i,raw:=range roles {
  pair:=strings.SplitN(strings.TrimSpace(raw),"=",2); if len(pair)!=2 {return IdentificationMeaning{},false,fmt.Errorf("bad role %q",raw)}
  d:=strings.Split(pair[1],":"); if len(d)<2 {return IdentificationMeaning{},false,fmt.Errorf("bad role descriptor %q",raw)}
  concept:=d[0]; count:=d[1]; isBlack:=len(d)>2 && d[2]=="black"
  if concept=="child" && variant=="offspring" {concept="child_of_person"}
  cid,e:=ids.NewConceptID(concept); if e!=nil{return IdentificationMeaning{},false,fmt.Errorf("case referent: %w",e)}
  rid:=meaning.NodeID(fmt.Sprintf("r%d",i+1)); ref,e:=meaning.NewReferent(rid,cid); if e!=nil{return IdentificationMeaning{},false,e}
  if count=="one" {ref.Multiplicity=meaning.MultiplicityOne} else if count=="many" {ref.Multiplicity=meaning.MultiplicityMany}
  if spec.index==6 && parts[4]=="pl" && concept=="door" {ref.Multiplicity=""}
  nodes=append(nodes,ref); participants=append(participants,meaning.Participant{Role:meaning.Role(pair[0]),Target:rid})
  if isBlack { black,e:=ids.NewConceptID("black"); if e!=nil{return IdentificationMeaning{},false,e}; st,e:=meaning.NewState(meaning.NodeID(fmt.Sprintf("black%d",i+1)),black,[]meaning.Participant{{Role:"bearer",Target:rid}}); if e!=nil{return IdentificationMeaning{},false,e}; blackStates=append(blackStates,st) }
 }
 event,err:=meaning.NewEvent("pred",predicate,participants); if err!=nil{return IdentificationMeaning{},false,err}; event.TemporalFrame=meaning.TemporalFramePresent
 nodes=append([]meaning.Node{event},nodes...); nodes=append(nodes,blackStates...)
 roots:=[]meaning.NodeID{"pred"}; for _,st:=range blackStates {roots=append(roots,st.ID)}
 graph,err:=meaning.Build(roots,nodes,meaning.DefaultBudget()); if err!=nil{return IdentificationMeaning{},false,err}
 return IdentificationMeaning{Graph:graph,Context:meaning.RequestContext{ID:"a1-11-case"}},true,nil
}

// CaseOrderExpansionIDs enumerates one ID per licensed interpretation.
func CaseOrderExpansionIDs() []string { var out []string; for _,s:=range caseOrderSpecs { if s.gap!="" {continue}; for _,lang:=range []string{"en","pl"} {base:=fmt.Sprintf("a1.11.case.%03d.%s",s.index,lang);out=append(out,base);if strings.Contains(s.roles,"child:"){out=append(out,base+".offspring")}} };return out }
